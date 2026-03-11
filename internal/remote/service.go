package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/openclaw/clawtool/internal/core"
	lifecyclelogs "github.com/openclaw/clawtool/internal/logs"
	platformcommon "github.com/openclaw/clawtool/internal/platform/common"
	"github.com/openclaw/clawtool/internal/schema"
	"github.com/openclaw/clawtool/internal/state"
	"gopkg.in/yaml.v3"
)

// Service prepares remote lifecycle operations from SSH profiles. / Service 基于 SSH 配置准备远程生命周期操作。
type Service struct {
	RootDir  string
	Executor Executor
}

// NewService creates a remote preparation service. / NewService 创建远程准备服务。
func NewService(rootDir string) Service {
	return NewServiceWithExecutor(rootDir, NewSSHExecutor())
}

// NewServiceWithExecutor creates a remote service with an explicit executor. / NewServiceWithExecutor 使用显式执行器创建远程服务。
func NewServiceWithExecutor(rootDir string, executor Executor) Service {
	if executor == nil {
		executor = NewSSHExecutor()
	}
	return Service{
		RootDir:  rootDir,
		Executor: executor,
	}
}

// LoadProfile loads and validates a remote SSH profile. / LoadProfile 加载并校验远程 SSH 配置。
func (s Service) LoadProfile(name string) (core.Profile, string, error) {
	path := filepath.Join(s.RootDir, ".clawtool", "profiles", name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return core.Profile{}, "", err
	}

	var profile core.Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return core.Profile{}, "", err
	}
	if err := schema.ValidateProfile(data); err != nil {
		return core.Profile{}, "", err
	}
	if profile.Target.Kind != "ssh" {
		return core.Profile{}, "", fmt.Errorf("profile %s is not an ssh target", name)
	}
	if strings.TrimSpace(profile.Target.Address) == "" {
		return core.Profile{}, "", fmt.Errorf("profile %s is missing ssh address", name)
	}

	return profile, path, nil
}

// Plan reuses the shared plan model for a remote target. / Plan 为远程目标复用共享计划模型。
func (s Service) Plan(profile core.Profile) core.Plan {
	steps := []core.PlanStep{
		{
			ID:         "remote.plan.validate_profile",
			Kind:       "remote",
			MessageKey: "remote.step.validate_profile",
			TemplateData: map[string]string{
				"Profile": profile.Name,
			},
		},
		{
			ID:         "remote.plan.prepare_connection",
			Kind:       "remote",
			MessageKey: "remote.step.prepare_connection",
			TemplateData: map[string]string{
				"Address": profile.Target.Address,
			},
		},
		{
			ID:         "remote.plan.sync_config",
			Kind:       "remote",
			MessageKey: "remote.step.sync_config",
			TemplateData: map[string]string{
				"Profile": profile.Name,
			},
		},
		{
			ID:         "remote.plan.verify_target",
			Kind:       "remote",
			MessageKey: "remote.step.verify_target",
			TemplateData: map[string]string{
				"Address": profile.Target.Address,
			},
		},
	}

	return core.Plan{
		Version:         "v1",
		Profile:         profile.Name,
		GeneratedConfig: filepath.Join(s.RootDir, ".clawtool", "cache", "effective-"+profile.Name+".yaml"),
		StatePath:       filepath.Join(s.RootDir, ".clawtool", "state", "state.json"),
		PlanRecordPath:  filepath.Join(s.RootDir, ".clawtool", "state", "last-plan.json"),
		RequiresBackup:  true,
		VerificationSteps: []string{
			"remote.verify.profile",
			"remote.verify.target_address",
			"remote.verify.target_user",
			"remote.verify.key_path",
			"remote.verify.host_key",
		},
		Changes: []core.PlanChange{
			{Path: "remote:ssh://" + profile.Target.Address + "/etc/openclaw/config.yaml", Action: "update"},
			{Path: "remote:ssh://" + profile.Target.Address + "/usr/local/bin/openclaw", Action: "verify"},
		},
		Steps: steps,
	}
}

// Verify performs deterministic preflight checks for remote execution. / Verify 为远程执行执行确定性的预检查。
func (s Service) Verify(profile core.Profile) core.VerifyResult {
	options, optionsErr := resolveConnectionOptions(profile)
	findings := []core.VerifyFinding{
		{Code: "remote.verify.profile", Severity: core.SeverityPass, MessageKey: "remote.verify.finding.profile.valid"},
		{Code: "remote.verify.target_kind", Severity: core.SeverityPass, MessageKey: "remote.verify.finding.target_kind.valid"},
	}

	if strings.TrimSpace(profile.Target.Address) == "" {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.target_address",
			Severity:   core.SeverityFail,
			MessageKey: "remote.verify.finding.target_address.missing",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.target_address",
			Severity:   core.SeverityPass,
			MessageKey: "remote.verify.finding.target_address.valid",
		})
	}

	if optionsErr != nil {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.target_user",
			Severity:   core.SeverityFail,
			MessageKey: "remote.verify.finding.target_user.missing",
		})
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.key_path",
			Severity:   core.SeverityWarn,
			MessageKey: "remote.verify.finding.key_path.missing",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.target_user",
			Severity:   core.SeverityPass,
			MessageKey: "remote.verify.finding.target_user.valid",
		})

		if strings.TrimSpace(options.KeyPath) == "" {
			findings = append(findings, core.VerifyFinding{
				Code:       "remote.verify.key_path",
				Severity:   core.SeverityWarn,
				MessageKey: "remote.verify.finding.key_path.missing",
			})
		} else {
			findings = append(findings, core.VerifyFinding{
				Code:       "remote.verify.key_path",
				Severity:   core.SeverityPass,
				MessageKey: "remote.verify.finding.key_path.valid",
			})
		}

		switch options.HostKeyStrategy {
		case "insecure":
			findings = append(findings, core.VerifyFinding{
				Code:       "remote.verify.host_key",
				Severity:   core.SeverityWarn,
				MessageKey: "remote.verify.finding.host_key.insecure",
			})
		default:
			if _, err := os.Stat(options.KnownHostsPath); err == nil {
				findings = append(findings, core.VerifyFinding{
					Code:       "remote.verify.host_key",
					Severity:   core.SeverityPass,
					MessageKey: "remote.verify.finding.host_key.known_hosts",
				})
			} else {
				findings = append(findings, core.VerifyFinding{
					Code:       "remote.verify.host_key",
					Severity:   core.SeverityWarn,
					MessageKey: "remote.verify.finding.host_key.missing_known_hosts",
				})
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Code < findings[j].Code
	})

	return core.VerifyResult{
		Profile:  profile.Name,
		Findings: findings,
	}
}

// Exec runs one command against the resolved SSH target. / Exec 对解析后的 SSH 目标执行一条命令。
func (s Service) Exec(ctx context.Context, profile core.Profile, command string) (core.RemoteExecResult, error) {
	options, err := resolveConnectionOptions(profile)
	if err != nil {
		return core.RemoteExecResult{}, err
	}
	if s.Executor == nil {
		s.Executor = NewSSHExecutor()
	}

	output, err := s.Executor.Execute(ctx, options, command)
	if err != nil {
		return core.RemoteExecResult{}, err
	}

	return core.RemoteExecResult{
		Profile:         profile.Name,
		Address:         options.OriginalAddress,
		User:            options.User,
		Port:            options.Port,
		HostKeyStrategy: options.HostKeyStrategy,
		Command:         command,
		Stdout:          output.Stdout,
		Stderr:          output.Stderr,
		ExitCode:        output.ExitCode,
		Duration:        output.Duration,
	}, nil
}

// Apply executes a minimal remote lifecycle on the SSH target. / Apply 在 SSH 目标上执行最小远程生命周期。
func (s Service) Apply(ctx context.Context, profile core.Profile) (core.RemoteApplyResult, error) {
	if err := s.ensureWorkspace(); err != nil {
		return core.RemoteApplyResult{}, err
	}

	precheck := s.Verify(profile)
	if hasRemoteFailure(precheck.Findings) {
		_ = s.recordApplyState(profile.Name, "fail", core.BackupRecord{})
		_ = s.logLifecycle("remote apply", "remote.apply.precheck", "fail", string(core.ErrorCodeApplyPrecheck), profile.Name)
		return core.RemoteApplyResult{}, &core.AppError{
			Code:       core.ErrorCodeApplyPrecheck,
			MessageKey: "error.apply.precheck",
			Cause:      fmt.Errorf("remote pre-check reported failing findings"),
		}
	}

	plan := s.Plan(profile)
	plan.PlanRecordPath = s.remotePlanRecordPath(profile.Name)
	plan.GeneratedConfig = s.effectiveConfigPath(profile.Name)

	configData, err := yaml.Marshal(profile)
	if err != nil {
		return core.RemoteApplyResult{}, err
	}

	changed := true
	if existing, err := os.ReadFile(plan.GeneratedConfig); err == nil && string(existing) == string(configData) {
		changed = false
	}
	if err := os.WriteFile(plan.GeneratedConfig, configData, 0o644); err != nil {
		return core.RemoteApplyResult{}, err
	}

	planData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return core.RemoteApplyResult{}, err
	}
	if err := os.WriteFile(plan.PlanRecordPath, planData, 0o644); err != nil {
		return core.RemoteApplyResult{}, err
	}

	options, err := resolveConnectionOptions(profile)
	if err != nil {
		return core.RemoteApplyResult{}, err
	}
	if s.Executor == nil {
		s.Executor = NewSSHExecutor()
	}

	backupRecord := core.BackupRecord{}
	remoteConfigPath := "/etc/openclaw/config.yaml"
	if output, err := s.Executor.Execute(ctx, options, "sh -lc "+shellQuote("test -f "+shellQuote(remoteConfigPath))); err == nil && output.ExitCode == 0 {
		backupPath := remoteConfigPath + ".bak." + time.Now().UTC().Format("20060102T150405Z")
		copyCommand := "sh -lc " + shellQuote("cp "+shellQuote(remoteConfigPath)+" "+shellQuote(backupPath))
		copyResult, execErr := s.Executor.Execute(ctx, options, copyCommand)
		if execErr != nil {
			_ = s.recordApplyState(profile.Name, "fail", backupRecord)
			return core.RemoteApplyResult{}, execErr
		}
		if copyResult.ExitCode != 0 {
			_ = s.recordApplyState(profile.Name, "fail", backupRecord)
			return core.RemoteApplyResult{}, &core.AppError{
				Code:       core.ErrorCodeRemoteExec,
				MessageKey: "error.remote.exec",
				Cause:      fmt.Errorf("remote backup command failed with exit code %d", copyResult.ExitCode),
			}
		}
		backupRecord = core.BackupRecord{
			ID:          filepath.Base(backupPath),
			ProfileName: profile.Name,
			CreatedAt:   time.Now().UTC(),
			Path:        "remote:ssh://" + profile.Target.Address + backupPath,
		}
	}

	if err := s.Executor.WriteFile(ctx, options, remoteConfigPath, configData, "0644"); err != nil {
		_ = s.recordApplyState(profile.Name, "fail", backupRecord)
		_ = s.logLifecycle("remote apply", "remote.apply.write_config", "fail", string(core.ErrorCodeRemoteExec), profile.Name)
		return core.RemoteApplyResult{}, err
	}

	verifyResultValue, err := s.verifyRemoteState(ctx, profile, options, remoteConfigPath)
	if err != nil {
		_ = s.recordApplyState(profile.Name, "fail", backupRecord)
		_ = s.logLifecycle("remote apply", "remote.apply.verify", "fail", string(core.ErrorCodeApplyVerify), profile.Name)
		return core.RemoteApplyResult{}, err
	}

	applyStatus := remoteApplyStateResult(verifyResultValue.Findings)
	if err := s.recordApplyState(profile.Name, applyStatus, backupRecord); err != nil {
		return core.RemoteApplyResult{}, err
	}
	if applyStatus == "fail" {
		_ = s.logLifecycle("remote apply", "remote.apply.verify", "fail", string(core.ErrorCodeApplyVerify), profile.Name)
		return core.RemoteApplyResult{}, &core.AppError{
			Code:       core.ErrorCodeApplyVerify,
			MessageKey: "error.apply.verify",
			Cause:      fmt.Errorf("remote post-apply verification reported failing findings"),
		}
	}

	now := time.Now().UTC()
	_ = s.logLifecycle("remote apply", "remote.apply.completed", applyStatus, "", profile.Name)
	return core.RemoteApplyResult{
		Plan:             plan,
		AppliedAt:        now,
		StatePath:        state.DefaultStatePath(s.RootDir),
		GeneratedConfig:  plan.GeneratedConfig,
		PlanRecordPath:   plan.PlanRecordPath,
		RemoteConfigPath: remoteConfigPath,
		BackupPath:       backupRecord.Path,
		VerifyResult:     verifyResultValue,
		Changed:          changed,
	}, nil
}

func (s Service) verifyRemoteState(ctx context.Context, profile core.Profile, options ConnectionOptions, remoteConfigPath string) (core.VerifyResult, error) {
	findings := append([]core.VerifyFinding{}, s.Verify(profile).Findings...)

	configCheck, err := s.Executor.Execute(ctx, options, "sh -lc "+shellQuote("test -s "+shellQuote(remoteConfigPath)))
	if err != nil {
		return core.VerifyResult{}, err
	}
	if configCheck.ExitCode == 0 {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.remote_config",
			Severity:   core.SeverityPass,
			MessageKey: "remote.verify.finding.remote_config.present",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.remote_config",
			Severity:   core.SeverityFail,
			MessageKey: "remote.verify.finding.remote_config.missing",
		})
	}

	openclawCheck, err := s.Executor.Execute(ctx, options, "sh -lc "+shellQuote("test -x /usr/local/bin/openclaw || command -v openclaw >/dev/null"))
	if err != nil {
		return core.VerifyResult{}, err
	}
	if openclawCheck.ExitCode == 0 {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.openclaw",
			Severity:   core.SeverityPass,
			MessageKey: "remote.verify.finding.openclaw.available",
		})
	} else {
		findings = append(findings, core.VerifyFinding{
			Code:       "remote.verify.openclaw",
			Severity:   core.SeverityFail,
			MessageKey: "remote.verify.finding.openclaw.missing",
		})
	}

	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Code < findings[j].Code
	})
	return core.VerifyResult{
		Profile:  profile.Name,
		Findings: findings,
	}, nil
}

func (s Service) ensureWorkspace() error {
	paths := []string{
		s.workspacePath(),
		filepath.Join(s.workspacePath(), "cache"),
		filepath.Join(s.workspacePath(), "state"),
		filepath.Join(s.workspacePath(), "logs"),
	}
	for _, path := range paths {
		if _, err := platformcommon.EnsureDirectory(path); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) workspacePath() string {
	return platformcommon.WorkspacePath(s.RootDir)
}

func (s Service) effectiveConfigPath(profileName string) string {
	return filepath.Join(s.workspacePath(), "cache", "effective-"+profileName+".yaml")
}

func (s Service) remotePlanRecordPath(profileName string) string {
	return filepath.Join(s.workspacePath(), "state", "remote-last-plan-"+profileName+".json")
}

func (s Service) logPath() string {
	return lifecyclelogs.DefaultLogPath(s.RootDir)
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

func (s Service) recordApplyState(profileName string, result string, backupRecord core.BackupRecord) error {
	record, err := state.Load(state.DefaultStatePath(s.RootDir))
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	record.Version = "v1"
	record.CurrentProfile = profileName
	record.LastApplyAt = &now
	record.LastApplyResult = result
	if backupRecord.Path != "" {
		record.Backups = append(record.Backups, backupRecord)
	}
	return state.Save(state.DefaultStatePath(s.RootDir), record)
}

func hasRemoteFailure(findings []core.VerifyFinding) bool {
	for _, finding := range findings {
		if finding.Severity == core.SeverityFail {
			return true
		}
	}
	return false
}

func remoteApplyStateResult(findings []core.VerifyFinding) string {
	for _, finding := range findings {
		if finding.Severity == core.SeverityFail {
			return "fail"
		}
	}
	return "success"
}
