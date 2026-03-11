package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/openclaw/clawtool/internal/core"
	lifecyclelogs "github.com/openclaw/clawtool/internal/logs"
	"github.com/openclaw/clawtool/internal/schema"
	"github.com/openclaw/clawtool/internal/state"
	"gopkg.in/yaml.v3"
)

const (
	workspaceDirName = ".clawtool"
	profilesDirName  = "profiles"
)

// Service provides local workspace operations. / Service 提供本地工作区操作。
type Service struct {
	RootDir string
}

// NewService creates a workspace service. / NewService 创建工作区服务。
func NewService(rootDir string) Service {
	return Service{RootDir: rootDir}
}

// Detect collects local environment facts. / Detect 收集本地环境事实。
func (s Service) Detect() (core.DetectResult, error) {
	workingDir, err := filepath.Abs(s.RootDir)
	if err != nil {
		return core.DetectResult{}, err
	}

	homeDir, _ := os.UserHomeDir()
	executablePath, _ := os.Executable()
	shell := detectShell()
	openClawPath, _ := exec.LookPath("openclaw")

	return core.DetectResult{
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		WorkingDir:        workingDir,
		HomeDir:           homeDir,
		Shell:             shell,
		ExecutablePath:    executablePath,
		OpenClawPath:      openClawPath,
		OpenClawInstalled: openClawPath != "",
		WorkspacePath:     s.workspacePath(),
		ProfilesPath:      s.profilesPath(),
		StatePath:         state.DefaultStatePath(workingDir),
	}, nil
}

// Doctor performs deterministic health checks. / Doctor 执行确定性的健康检查。
func (s Service) Doctor() (core.DoctorResult, error) {
	detectResult, err := s.Detect()
	if err != nil {
		return core.DoctorResult{}, err
	}

	findings := []core.DoctorFinding{
		{
			Code:               "doctor.installation.openclaw",
			Severity:           ternarySeverity(detectResult.OpenClawInstalled, core.SeverityPass, core.SeverityWarn),
			MessageKey:         ternaryKey(detectResult.OpenClawInstalled, "doctor.finding.openclaw.present", "doctor.finding.openclaw.missing"),
			RemediationHintKey: "doctor.hint.openclaw.install",
		},
	}

	findings = append(findings, s.checkDirectoryWritable(s.RootDir, "doctor.fs.root_writable", "doctor.hint.root_writable"))
	findings = append(findings, s.checkDirectoryWritable(s.workspacePath(), "doctor.fs.workspace_writable", "doctor.hint.workspace_init"))

	profilesPath := s.profilesPath()
	if _, err := os.Stat(profilesPath); err == nil {
		findings = append(findings, core.DoctorFinding{
			Code:               "doctor.profiles.directory",
			Severity:           core.SeverityPass,
			MessageKey:         "doctor.finding.profiles.present",
			RemediationHintKey: "doctor.hint.none",
		})
	} else {
		findings = append(findings, core.DoctorFinding{
			Code:               "doctor.profiles.directory",
			Severity:           core.SeverityWarn,
			MessageKey:         "doctor.finding.profiles.missing",
			RemediationHintKey: "doctor.hint.workspace_init",
		})
	}

	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Code < findings[j].Code
	})

	return core.DoctorResult{Findings: findings}, nil
}

// Init initializes local workspace state. / Init 初始化本地工作区状态。
func (s Service) Init() (core.InitResult, error) {
	paths := []string{
		s.workspacePath(),
		s.profilesPath(),
		filepath.Join(s.workspacePath(), "backups"),
		filepath.Join(s.workspacePath(), "cache"),
		filepath.Join(s.workspacePath(), "state"),
		filepath.Join(s.workspacePath(), "logs"),
	}

	result := core.InitResult{
		WorkspacePath:  s.workspacePath(),
		DefaultProfile: filepath.Join(s.profilesPath(), "default.yaml"),
		StatePath:      state.DefaultStatePath(s.RootDir),
	}

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			result.ExistingPaths = append(result.ExistingPaths, path)
			continue
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return core.InitResult{}, err
		}
		result.CreatedPaths = append(result.CreatedPaths, path)
	}

	if err := s.ensureDefaultProfile(); err != nil {
		return core.InitResult{}, err
	}
	if err := s.ensureState(); err != nil {
		return core.InitResult{}, err
	}
	_ = s.logLifecycle("init", "workspace.ready", "success", "", "default")

	return result, nil
}

// Plan computes a deterministic local execution plan. / Plan 计算确定性的本地执行计划。
func (s Service) Plan() (core.Plan, error) {
	profile, _, err := s.activeProfile()
	if err != nil {
		return core.Plan{}, err
	}

	steps := []core.PlanStep{
		{
			ID:          "plan.ensure.workspace",
			Kind:        "filesystem",
			Description: "Ensure workspace directories exist",
		},
		{
			ID:          "plan.render.profile",
			Kind:        "config",
			Description: fmt.Sprintf("Render effective configuration for profile %s", profile.Name),
		},
		{
			ID:          "plan.persist.state",
			Kind:        "state",
			Description: "Persist lifecycle state",
		},
	}

	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.Plan{}, err
	}

	return core.Plan{
		Version:         "v1",
		Profile:         profile.Name,
		GeneratedConfig: s.effectiveConfigPath(profile.Name),
		StatePath:       state.DefaultStatePath(s.RootDir),
		PlanRecordPath:  filepath.Join(s.workspacePath(), "state", "last-plan.json"),
		RequiresBackup:  record.LastApplyAt != nil,
		VerificationSteps: []string{
			"verify.workspace",
			"verify.profile",
			"verify.generated_config",
			"verify.plan_record",
		},
		Changes:      s.planChanges(profile),
		ContentDiffs: s.planContentDiffs(profile),
		Steps:        steps,
	}, nil
}

// Apply executes the local plan and persists state. / Apply 执行本地计划并持久化状态。
func (s Service) Apply() (core.ApplyResult, error) {
	if _, err := s.Init(); err != nil {
		return core.ApplyResult{}, err
	}

	plan, err := s.Plan()
	if err != nil {
		return core.ApplyResult{}, err
	}

	profile, _, err := s.activeProfile()
	if err != nil {
		return core.ApplyResult{}, err
	}

	configData, err := yaml.Marshal(profile)
	if err != nil {
		return core.ApplyResult{}, err
	}
	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.ApplyResult{}, err
	}

	changed := true
	backupRecord := core.BackupRecord{}
	backupRecord, err = s.backupPreviousConfig(profile.Name)
	if err != nil {
		return core.ApplyResult{}, err
	}
	if existing, err := os.ReadFile(plan.GeneratedConfig); err == nil {
		if bytes.Equal(existing, configData) {
			changed = false
		} else if backupRecord.Path == "" {
			backupRecord, err = s.backupGeneratedConfig(profile.Name, existing)
			if err != nil {
				return core.ApplyResult{}, err
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(plan.GeneratedConfig), 0o755); err != nil {
		return core.ApplyResult{}, err
	}
	if changed {
		if err := os.WriteFile(plan.GeneratedConfig, configData, 0o644); err != nil {
			return core.ApplyResult{}, err
		}
	}

	planRecordPath := plan.PlanRecordPath
	planData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return core.ApplyResult{}, err
	}
	if err := os.WriteFile(planRecordPath, planData, 0o644); err != nil {
		return core.ApplyResult{}, err
	}

	now := time.Now().UTC()
	record.Version = "v1"
	record.CurrentProfile = profile.Name
	record.LastApplyAt = &now
	record.LastApplyResult = "success"
	if backupRecord.Path != "" {
		record.Backups = append(record.Backups, backupRecord)
	}
	if err := state.Save(state.DefaultStatePath(s.RootDir), record); err != nil {
		return core.ApplyResult{}, err
	}
	_ = s.logLifecycle("apply", "apply.completed", "success", "", profile.Name)
	return core.ApplyResult{
		Plan:            plan,
		AppliedAt:       now,
		StatePath:       state.DefaultStatePath(s.RootDir),
		GeneratedConfig: plan.GeneratedConfig,
		PlanRecordPath:  planRecordPath,
		BackupPath:      backupRecord.Path,
		Changed:         changed,
	}, nil
}

// Verify validates the current local lifecycle state. / Verify 验证当前本地生命周期状态。
func (s Service) Verify() (core.VerifyResult, error) {
	profile, _, err := s.activeProfile()
	if err != nil {
		return core.VerifyResult{}, err
	}

	plan, err := s.Plan()
	if err != nil {
		return core.VerifyResult{}, err
	}

	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.VerifyResult{}, err
	}

	findings := []core.VerifyFinding{
		{
			Code:       "verify.profile.active",
			Severity:   core.SeverityPass,
			MessageKey: "verify.finding.profile.active",
		},
	}

	if _, err := os.Stat(s.workspacePath()); err == nil {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.workspace.present",
			Severity:   core.SeverityPass,
			MessageKey: "verify.finding.workspace.present",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.workspace.present",
			Severity:   core.SeverityFail,
			MessageKey: "verify.finding.workspace.missing",
		})
	}

	if _, err := os.Stat(plan.GeneratedConfig); err == nil {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.generated_config.present",
			Severity:   core.SeverityPass,
			MessageKey: "verify.finding.generated_config.present",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.generated_config.present",
			Severity:   core.SeverityFail,
			MessageKey: "verify.finding.generated_config.missing",
		})
	}

	if _, err := os.Stat(plan.PlanRecordPath); err == nil {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.plan_record.present",
			Severity:   core.SeverityPass,
			MessageKey: "verify.finding.plan_record.present",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.plan_record.present",
			Severity:   core.SeverityWarn,
			MessageKey: "verify.finding.plan_record.missing",
		})
	}

	if record.LastApplyResult == "success" {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.last_apply.success",
			Severity:   core.SeverityPass,
			MessageKey: "verify.finding.last_apply.success",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "verify.last_apply.success",
			Severity:   core.SeverityWarn,
			MessageKey: "verify.finding.last_apply.missing",
		})
	}

	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Code < findings[j].Code
	})

	result := verifyResult(findings)
	_ = s.logLifecycle("verify", "verify.completed", result, "", profile.Name)

	return core.VerifyResult{
		Profile:  profile.Name,
		Findings: findings,
	}, nil
}

// Inspect returns detailed lifecycle metadata. / Inspect 返回详细生命周期元数据。
func (s Service) Inspect() (core.InspectResult, error) {
	detectResult, err := s.Detect()
	if err != nil {
		return core.InspectResult{}, err
	}

	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.InspectResult{}, err
	}
	profiles, err := s.ListProfiles()
	if err != nil {
		return core.InspectResult{}, err
	}

	return core.InspectResult{
		CurrentProfile:  record.CurrentProfile,
		Profiles:        profiles.Profiles,
		InstallPath:     detectResult.OpenClawPath,
		ConfigPath:      s.effectiveConfigPath(record.CurrentProfile),
		StatePath:       detectResult.StatePath,
		PlanRecordPath:  filepath.Join(s.workspacePath(), "state", "last-plan.json"),
		LastApplyAt:     record.LastApplyAt,
		LastApplyResult: record.LastApplyResult,
		Backups:         record.Backups,
	}, nil
}

// Status returns a compact lifecycle summary. / Status 返回紧凑的生命周期摘要。
func (s Service) Status() (core.StatusResult, error) {
	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.StatusResult{}, err
	}

	return core.StatusResult{
		CurrentProfile:     record.CurrentProfile,
		LastApplyAt:        record.LastApplyAt,
		LastApplyResult:    record.LastApplyResult,
		LastRollbackAt:     record.LastRollbackAt,
		LastRollbackResult: record.LastRollbackResult,
		BackupCount:        len(record.Backups),
	}, nil
}

// Rollback restores the latest or selected backup. / Rollback 恢复最近一次或指定的备份。
func (s Service) Rollback(backupID string) (core.RollbackResult, error) {
	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.RollbackResult{}, err
	}

	selected, err := selectBackup(record.Backups, backupID)
	if err != nil {
		return core.RollbackResult{}, err
	}

	now := time.Now().UTC()
	currentProfile := record.CurrentProfile
	if strings.TrimSpace(currentProfile) == "" {
		currentProfile = "default"
	}

	preRollbackBackupPath := ""
	if existing, err := os.ReadFile(s.effectiveConfigPath(currentProfile)); err == nil {
		preRollbackBackup, err := s.backupGeneratedConfig(currentProfile, existing)
		if err != nil {
			return core.RollbackResult{}, err
		}
		preRollbackBackupPath = preRollbackBackup.Path
		record.Backups = append(record.Backups, preRollbackBackup)
	}

	data, err := os.ReadFile(selected.Path)
	if err != nil {
		return core.RollbackResult{}, &core.AppError{
			Code:       core.ErrorCodeRollback,
			MessageKey: "error.rollback.failed",
			Cause:      err,
		}
	}

	restoredPath := s.effectiveConfigPath(selected.ProfileName)
	if err := os.MkdirAll(filepath.Dir(restoredPath), 0o755); err != nil {
		return core.RollbackResult{}, &core.AppError{
			Code:       core.ErrorCodeRollback,
			MessageKey: "error.rollback.failed",
			Cause:      err,
		}
	}
	if err := os.WriteFile(restoredPath, data, 0o644); err != nil {
		return core.RollbackResult{}, &core.AppError{
			Code:       core.ErrorCodeRollback,
			MessageKey: "error.rollback.failed",
			Cause:      err,
		}
	}

	record.CurrentProfile = selected.ProfileName
	record.LastRollbackAt = &now
	record.LastRollbackResult = "success"
	if err := state.Save(state.DefaultStatePath(s.RootDir), record); err != nil {
		return core.RollbackResult{}, err
	}

	verifyResultValue, err := s.Verify()
	if err != nil {
		return core.RollbackResult{}, err
	}
	_ = s.logLifecycle("rollback", "rollback.completed", "success", "", selected.ProfileName)

	return core.RollbackResult{
		AppliedAt:             now,
		StatePath:             state.DefaultStatePath(s.RootDir),
		SelectedBackup:        selected,
		RestoredProfile:       selected.ProfileName,
		RestoredPath:          restoredPath,
		PreRollbackBackupPath: preRollbackBackupPath,
		VerifyResult:          verifyResultValue,
	}, nil
}

// Logs returns filtered lifecycle logs and optional bundle metadata. / Logs 返回筛选后的生命周期日志以及可选归档元数据。
func (s Service) Logs(tail int, since *time.Time, bundle bool) (core.LogsResult, error) {
	logPath := s.logPath()
	entries, err := lifecyclelogs.Read(logPath)
	if err != nil {
		return core.LogsResult{}, err
	}

	filtered := filterLogEntries(entries, since)
	if tail > 0 && len(filtered) > tail {
		filtered = filtered[len(filtered)-tail:]
	}

	result := core.LogsResult{
		LogPath: logPath,
		Entries: filtered,
	}
	if bundle {
		metadata, err := s.bundleLogs(logPath, filtered)
		if err != nil {
			return core.LogsResult{}, err
		}
		result.Bundle = &metadata
	}

	return result, nil
}

// Repair plans or executes deterministic repair actions. / Repair 规划或执行确定性的修复动作。
func (s Service) Repair(applySafe bool, yes bool) (core.RepairResult, error) {
	actions := s.repairActions()
	appliedCount := 0

	for index := range actions {
		action := &actions[index]
		switch action.Code {
		case "repair.workspace.init":
			if applySafe {
				if _, err := s.Init(); err != nil {
					return core.RepairResult{}, err
				}
				action.Applied = true
				appliedCount++
				_ = s.logLifecycle("repair", action.Code, "success", "", "")
			}
		case "repair.generated_config.apply":
			if yes {
				if _, err := s.Apply(); err != nil {
					return core.RepairResult{}, err
				}
				action.Applied = true
				appliedCount++
				_ = s.logLifecycle("repair", action.Code, "success", "", "")
			}
		}
	}

	return core.RepairResult{
		Actions:      actions,
		AppliedCount: appliedCount,
	}, nil
}

// ListProfiles enumerates YAML profiles in the workspace. / ListProfiles 列出工作区中的 YAML 配置。
func (s Service) ListProfiles() (core.ProfileListResult, error) {
	record, _ := state.Load(state.DefaultStatePath(s.RootDir))
	entries, err := os.ReadDir(s.profilesPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return core.ProfileListResult{}, nil
		}
		return core.ProfileListResult{}, err
	}

	result := core.ProfileListResult{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(s.profilesPath(), entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return core.ProfileListResult{}, err
		}
		var profile core.Profile
		if err := yaml.Unmarshal(data, &profile); err != nil {
			return core.ProfileListResult{}, err
		}
		result.Profiles = append(result.Profiles, core.ProfileSummary{
			Name:   profile.Name,
			Path:   path,
			Active: record.CurrentProfile == profile.Name,
		})
	}

	sort.Slice(result.Profiles, func(i, j int) bool {
		return result.Profiles[i].Name < result.Profiles[j].Name
	})
	return result, nil
}

// ShowProfile loads one profile by name. / ShowProfile 按名称加载单个配置。
func (s Service) ShowProfile(name string) (core.ProfileShowResult, error) {
	path := filepath.Join(s.profilesPath(), name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return core.ProfileShowResult{}, err
	}

	var profile core.Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return core.ProfileShowResult{}, err
	}

	record, _ := state.Load(state.DefaultStatePath(s.RootDir))
	return core.ProfileShowResult{
		Profile: profile,
		Path:    path,
		Active:  record.CurrentProfile == profile.Name,
	}, nil
}

// CreateProfile creates a profile from provided settings. / CreateProfile 根据提供的设置创建配置。
func (s Service) CreateProfile(profile core.Profile) (core.ProfileShowResult, error) {
	if profile.Version == "" {
		profile.Version = "v1"
	}
	if err := schema.ValidateProfile(mustYAML(profile)); err != nil {
		return core.ProfileShowResult{}, err
	}

	if err := os.MkdirAll(s.profilesPath(), 0o755); err != nil {
		return core.ProfileShowResult{}, err
	}
	path := filepath.Join(s.profilesPath(), profile.Name+".yaml")
	if _, err := os.Stat(path); err == nil {
		return core.ProfileShowResult{}, fmt.Errorf("profile already exists: %s", profile.Name)
	}

	data, err := yaml.Marshal(profile)
	if err != nil {
		return core.ProfileShowResult{}, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return core.ProfileShowResult{}, err
	}

	return core.ProfileShowResult{Profile: profile, Path: path}, nil
}

// ValidateProfile validates a profile file by name. / ValidateProfile 按名称校验配置文件。
func (s Service) ValidateProfile(name string) (core.ProfileShowResult, error) {
	result, err := s.ShowProfile(name)
	if err != nil {
		return core.ProfileShowResult{}, err
	}
	if err := schema.ValidateProfile(mustYAML(result.Profile)); err != nil {
		return core.ProfileShowResult{}, err
	}
	return result, nil
}

// UseProfile marks one profile as active in state. / UseProfile 在状态中标记活动配置。
func (s Service) UseProfile(name string) (core.ProfileUseResult, error) {
	show, err := s.ShowProfile(name)
	if err != nil {
		return core.ProfileUseResult{}, err
	}

	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.ProfileUseResult{}, err
	}
	record.Version = "v1"
	record.CurrentProfile = show.Profile.Name
	if err := state.Save(state.DefaultStatePath(s.RootDir), record); err != nil {
		return core.ProfileUseResult{}, err
	}

	return core.ProfileUseResult{Name: show.Profile.Name, Path: show.Path}, nil
}

func (s Service) workspacePath() string {
	return filepath.Join(s.RootDir, workspaceDirName)
}

func (s Service) profilesPath() string {
	return filepath.Join(s.workspacePath(), profilesDirName)
}

func (s Service) ensureDefaultProfile() error {
	path := filepath.Join(s.profilesPath(), "default.yaml")
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	profile := core.Profile{
		Version: "v1",
		Name:    "default",
		Target:  core.ProfileTarget{Kind: "local"},
	}

	data, err := yaml.Marshal(profile)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func (s Service) ensureState() error {
	statePath := state.DefaultStatePath(s.RootDir)
	if _, err := os.Stat(statePath); err == nil {
		return nil
	}

	return state.Save(statePath, core.StateRecord{
		Version:        "v1",
		CurrentProfile: "default",
	})
}

func (s Service) activeProfile() (core.Profile, string, error) {
	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return core.Profile{}, "", err
	}

	name := record.CurrentProfile
	if strings.TrimSpace(name) == "" {
		name = "default"
	}

	show, err := s.ShowProfile(name)
	if err != nil {
		return core.Profile{}, "", err
	}
	return show.Profile, show.Path, nil
}

func (s Service) effectiveConfigPath(profileName string) string {
	return filepath.Join(s.workspacePath(), "cache", "effective-"+profileName+".yaml")
}

func (s Service) logPath() string {
	return lifecyclelogs.DefaultLogPath(s.RootDir)
}

func (s Service) planChanges(profile core.Profile) []core.PlanChange {
	changes := []core.PlanChange{
		{Path: s.effectiveConfigPath(profile.Name), Action: "create"},
		{Path: filepath.Join(s.workspacePath(), "state", "last-plan.json"), Action: "update"},
		{Path: state.DefaultStatePath(s.RootDir), Action: "update"},
	}

	configData, err := yaml.Marshal(profile)
	if err != nil {
		return changes
	}

	if existing, err := os.ReadFile(s.effectiveConfigPath(profile.Name)); err == nil {
		if bytes.Equal(existing, configData) {
			changes[0].Action = "noop"
		} else {
			changes[0].Action = "update"
		}
	}

	return changes
}

func (s Service) planContentDiffs(profile core.Profile) []core.ContentDiff {
	targetData, err := yaml.Marshal(profile)
	if err != nil {
		return nil
	}

	before := map[string]string{}
	if existing, err := os.ReadFile(s.effectiveConfigPath(profile.Name)); err == nil {
		before = flattenYAML(existing)
	}
	after := flattenYAML(targetData)

	fields := append(sortedStringKeys(before), sortedStringKeys(after)...)
	seen := map[string]struct{}{}
	diffs := make([]core.ContentDiff, 0, len(fields))

	for _, field := range fields {
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}

		beforeValue, beforeOK := before[field]
		afterValue, afterOK := after[field]
		switch {
		case beforeOK && afterOK && beforeValue == afterValue:
			continue
		case !beforeOK && afterOK:
			diffs = append(diffs, core.ContentDiff{Field: field, Action: "add", After: afterValue})
		case beforeOK && !afterOK:
			diffs = append(diffs, core.ContentDiff{Field: field, Action: "remove", Before: beforeValue})
		default:
			diffs = append(diffs, core.ContentDiff{Field: field, Action: "update", Before: beforeValue, After: afterValue})
		}
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Field < diffs[j].Field
	})
	return diffs
}

func (s Service) backupGeneratedConfig(profileName string, data []byte) (core.BackupRecord, error) {
	backupDir := filepath.Join(s.workspacePath(), "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return core.BackupRecord{}, err
	}

	createdAt := time.Now().UTC()
	name := fmt.Sprintf("%s-%s.yaml.bak", profileName, createdAt.Format("20060102T150405Z"))
	path := filepath.Join(backupDir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return core.BackupRecord{}, err
	}
	return core.BackupRecord{
		ID:          name,
		ProfileName: profileName,
		CreatedAt:   createdAt,
		Path:        path,
	}, nil
}

func (s Service) backupPreviousConfig(currentProfile string) (core.BackupRecord, error) {
	cacheDir := filepath.Join(s.workspacePath(), "cache")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return core.BackupRecord{}, nil
		}
		return core.BackupRecord{}, err
	}

	prefix := "effective-"
	suffix := ".yaml"
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), prefix), suffix)
		if name == currentProfile {
			continue
		}
		path := filepath.Join(cacheDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return core.BackupRecord{}, err
		}
		return s.backupGeneratedConfig(name, data)
	}

	return core.BackupRecord{}, nil
}

func (s Service) checkDirectoryWritable(path string, code string, hintKey string) core.DoctorFinding {
	testPath := filepath.Join(path, ".clawtool-write-check")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return core.DoctorFinding{
			Code:               code,
			Severity:           core.SeverityFail,
			MessageKey:         "doctor.finding.directory.not_writable",
			RemediationHintKey: hintKey,
		}
	}

	if err := os.WriteFile(testPath, []byte("ok"), 0o644); err != nil {
		return core.DoctorFinding{
			Code:               code,
			Severity:           core.SeverityFail,
			MessageKey:         "doctor.finding.directory.not_writable",
			RemediationHintKey: hintKey,
		}
	}
	_ = os.Remove(testPath)

	return core.DoctorFinding{
		Code:               code,
		Severity:           core.SeverityPass,
		MessageKey:         "doctor.finding.directory.writable",
		RemediationHintKey: "doctor.hint.none",
	}
}

func detectShell() string {
	for _, key := range []string{"SHELL", "COMSPEC"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func mustYAML(value any) []byte {
	data, _ := yaml.Marshal(value)
	return data
}

func ternarySeverity(condition bool, a core.Severity, b core.Severity) core.Severity {
	if condition {
		return a
	}
	return b
}

func ternaryKey(condition bool, a string, b string) string {
	if condition {
		return a
	}
	return b
}

func sortedStringKeys[V any](value map[string]V) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func flattenYAML(data []byte) map[string]string {
	var value any
	if err := yaml.Unmarshal(data, &value); err != nil {
		return map[string]string{}
	}

	flattened := map[string]string{}
	flattenValue("", normalizeYAMLValue(value), flattened)
	return flattened
}

func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		normalized := map[string]any{}
		for key, item := range typed {
			normalized[key] = normalizeYAMLValue(item)
		}
		return normalized
	case map[any]any:
		normalized := map[string]any{}
		for key, item := range typed {
			normalized[fmt.Sprint(key)] = normalizeYAMLValue(item)
		}
		return normalized
	case []any:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, normalizeYAMLValue(item))
		}
		return items
	default:
		return typed
	}
}

func flattenValue(prefix string, value any, result map[string]string) {
	switch typed := value.(type) {
	case map[string]any:
		keys := sortedStringKeys(typed)
		for _, key := range keys {
			next := key
			if prefix != "" {
				next = prefix + "." + key
			}
			flattenValue(next, typed[key], result)
		}
	case []any:
		for index, item := range typed {
			next := fmt.Sprintf("%s[%d]", prefix, index)
			flattenValue(next, item, result)
		}
	default:
		result[prefix] = fmt.Sprint(typed)
	}
}

func (s Service) logLifecycle(command string, step string, result string, errorCode string, profile string) error {
	return lifecyclelogs.Append(s.logPath(), core.LifecycleLogEntry{
		Timestamp: time.Now().UTC(),
		Command:   command,
		Step:      step,
		Result:    result,
		ErrorCode: errorCode,
		Profile:   profile,
	})
}

func filterLogEntries(entries []core.LifecycleLogEntry, since *time.Time) []core.LifecycleLogEntry {
	if since == nil {
		return entries
	}

	filtered := make([]core.LifecycleLogEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Timestamp.Before(*since) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func (s Service) repairActions() []core.RepairAction {
	actions := make([]core.RepairAction, 0, 2)

	if _, err := os.Stat(s.workspacePath()); err != nil {
		actions = append(actions, core.RepairAction{
			Code:       "repair.workspace.init",
			MessageKey: "repair.action.workspace.init",
			Risky:      false,
		})
		return actions
	}

	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return actions
	}

	profileName := record.CurrentProfile
	if strings.TrimSpace(profileName) == "" {
		profileName = "default"
	}
	if _, err := os.Stat(s.effectiveConfigPath(profileName)); err != nil {
		actions = append(actions, core.RepairAction{
			Code:         "repair.generated_config.apply",
			MessageKey:   "repair.action.generated_config.apply",
			Risky:        true,
			RequiresFlag: "--yes",
		})
	}

	return actions
}

func selectBackup(backups []core.BackupRecord, backupID string) (core.BackupRecord, error) {
	if len(backups) == 0 {
		return core.BackupRecord{}, &core.AppError{
			Code:       core.ErrorCodeBackupNotFound,
			MessageKey: "error.backup.not_found",
		}
	}

	if strings.TrimSpace(backupID) != "" {
		for _, backup := range backups {
			if backup.ID == backupID {
				return backup, nil
			}
		}
		return core.BackupRecord{}, &core.AppError{
			Code:       core.ErrorCodeBackupNotFound,
			MessageKey: "error.backup.not_found",
		}
	}

	selected := backups[0]
	for _, backup := range backups[1:] {
		if backup.CreatedAt.After(selected.CreatedAt) {
			selected = backup
		}
	}
	return selected, nil
}

func (s Service) bundleLogs(logPath string, entries []core.LifecycleLogEntry) (core.LogBundleMetadata, error) {
	bundleDir := filepath.Join(s.workspacePath(), "logs")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		return core.LogBundleMetadata{}, &core.AppError{
			Code:       core.ErrorCodeLogBundle,
			MessageKey: "error.logs.bundle",
			Cause:      err,
		}
	}

	createdAt := time.Now().UTC()
	bundlePath := filepath.Join(bundleDir, "bundle-"+createdAt.Format("20060102T150405Z")+".zip")
	file, err := os.Create(bundlePath)
	if err != nil {
		return core.LogBundleMetadata{}, &core.AppError{
			Code:       core.ErrorCodeLogBundle,
			MessageKey: "error.logs.bundle",
			Cause:      err,
		}
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	if data, err := os.ReadFile(logPath); err == nil {
		if err := writeZipFile(writer, "logs/lifecycle.log", data); err != nil {
			return core.LogBundleMetadata{}, err
		}
	}

	if data, err := os.ReadFile(state.DefaultStatePath(s.RootDir)); err == nil {
		if err := writeZipFile(writer, "state/state.json", data); err != nil {
			return core.LogBundleMetadata{}, err
		}
	}

	if inspectResultValue, err := s.Inspect(); err == nil {
		if data, err := json.MarshalIndent(inspectResultValue, "", "  "); err == nil {
			if err := writeZipFile(writer, "state/inspect-summary.json", data); err != nil {
				return core.LogBundleMetadata{}, err
			}
		}
	}

	failureSummary := map[string]any{}
	if record, err := state.Load(state.DefaultStatePath(s.RootDir)); err == nil {
		failureSummary["last_apply_result"] = record.LastApplyResult
		failureSummary["last_rollback_result"] = record.LastRollbackResult
		if len(entries) > 0 {
			failureSummary["latest_log_result"] = entries[len(entries)-1].Result
		}
	}
	if data, err := json.MarshalIndent(failureSummary, "", "  "); err == nil {
		if err := writeZipFile(writer, "state/recent-failure.json", data); err != nil {
			return core.LogBundleMetadata{}, err
		}
	}

	return core.LogBundleMetadata{
		Version:    "v1",
		CreatedAt:  createdAt,
		BundlePath: bundlePath,
	}, nil
}

func writeZipFile(writer *zip.Writer, name string, data []byte) error {
	file, err := writer.Create(name)
	if err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeLogBundle,
			MessageKey: "error.logs.bundle",
			Cause:      err,
		}
	}
	if _, err := io.Copy(file, bytes.NewReader(data)); err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeLogBundle,
			MessageKey: "error.logs.bundle",
			Cause:      err,
		}
	}
	return nil
}

func verifyResult(findings []core.VerifyFinding) string {
	result := "success"
	for _, finding := range findings {
		if finding.Severity == core.SeverityFail {
			return "fail"
		}
		if finding.Severity == core.SeverityWarn {
			result = "warn"
		}
	}
	return result
}
