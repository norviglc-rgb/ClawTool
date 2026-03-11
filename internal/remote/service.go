package remote

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/openclaw/clawtool/internal/schema"
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
