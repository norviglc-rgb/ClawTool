package core

import "time"

// Severity classifies findings deterministically. / Severity 以确定性方式对发现项分级。
type Severity string

const (
	SeverityPass Severity = "PASS"
	SeverityWarn Severity = "WARN"
	SeverityFail Severity = "FAIL"
)

// ResultStatus is the high-level command status. / ResultStatus 是命令级别的结果状态。
type ResultStatus string

const (
	ResultStatusOK    ResultStatus = "ok"
	ResultStatusWarn  ResultStatus = "warn"
	ResultStatusError ResultStatus = "error"
)

// DetailItem is a stable key-value output field. / DetailItem 是稳定的键值输出字段。
type DetailItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CommandResult separates command execution from rendering. / CommandResult 将命令执行与渲染分离。
type CommandResult struct {
	Command    string       `json:"command"`
	Status     ResultStatus `json:"status"`
	SummaryKey string       `json:"summary_key"`
	Details    []DetailItem `json:"details"`
}

// Profile describes local or remote execution intent. / Profile 描述本地或远程执行意图。
type Profile struct {
	Version string        `json:"version" yaml:"version"`
	Name    string        `json:"name" yaml:"name"`
	Target  ProfileTarget `json:"target" yaml:"target"`
}

// ProfileTarget keeps transport-specific settings grouped. / ProfileTarget 将传输相关设置集中管理。
type ProfileTarget struct {
	Kind    string `json:"kind" yaml:"kind"`
	Address string `json:"address,omitempty" yaml:"address,omitempty"`
}

// Manifest describes resolved configuration inputs. / Manifest 描述解析后的配置输入。
type Manifest struct {
	Version string `json:"version" yaml:"version"`
	Profile string `json:"profile" yaml:"profile"`
}

// Plan is the deterministic output of plan resolution. / Plan 是规划解析后的确定性结果。
type Plan struct {
	Version           string        `json:"version"`
	Profile           string        `json:"profile"`
	GeneratedConfig   string        `json:"generated_config"`
	StatePath         string        `json:"state_path"`
	PlanRecordPath    string        `json:"plan_record_path"`
	RequiresBackup    bool          `json:"requires_backup"`
	VerificationSteps []string      `json:"verification_steps"`
	Changes           []PlanChange  `json:"changes"`
	ContentDiffs      []ContentDiff `json:"content_diffs"`
	Steps             []PlanStep    `json:"steps"`
}

// PlanStep is an individual planned action. / PlanStep 是单个计划动作。
type PlanStep struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

// PlanChange describes a deterministic file change. / PlanChange 描述确定性的文件变更。
type PlanChange struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}

// ContentDiff describes a field-level configuration diff. / ContentDiff 描述字段级配置差异。
type ContentDiff struct {
	Field  string `json:"field"`
	Action string `json:"action"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// BackupRecord tracks a single backup artifact. / BackupRecord 跟踪单个备份工件。
type BackupRecord struct {
	ID          string    `json:"id"`
	ProfileName string    `json:"profile_name"`
	CreatedAt   time.Time `json:"created_at"`
	Path        string    `json:"path"`
}

// StateRecord stores durable lifecycle state. / StateRecord 存储持久的生命周期状态。
type StateRecord struct {
	Version            string         `json:"version"`
	CurrentProfile     string         `json:"current_profile"`
	LastApplyAt        *time.Time     `json:"last_apply_at,omitempty"`
	LastApplyResult    string         `json:"last_apply_result"`
	LastRollbackAt     *time.Time     `json:"last_rollback_at,omitempty"`
	LastRollbackResult string         `json:"last_rollback_result,omitempty"`
	Backups            []BackupRecord `json:"backups"`
}

// DoctorFinding describes a health-check result. / DoctorFinding 描述健康检查结果。
type DoctorFinding struct {
	Code               string   `json:"code"`
	Severity           Severity `json:"severity"`
	MessageKey         string   `json:"message_key"`
	RemediationHintKey string   `json:"remediation_hint_key"`
}

// VerifyFinding describes a verification result. / VerifyFinding 描述验证结果。
type VerifyFinding struct {
	Code       string   `json:"code"`
	Severity   Severity `json:"severity"`
	MessageKey string   `json:"message_key"`
}

// ApplyResult captures apply execution outcomes. / ApplyResult 描述 apply 执行结果。
type ApplyResult struct {
	Plan            Plan      `json:"plan"`
	AppliedAt       time.Time `json:"applied_at"`
	StatePath       string    `json:"state_path"`
	GeneratedConfig string    `json:"generated_config"`
	PlanRecordPath  string    `json:"plan_record_path"`
	BackupPath      string    `json:"backup_path,omitempty"`
	Changed         bool      `json:"changed"`
}

// RollbackResult captures rollback execution outcomes. / RollbackResult 描述 rollback 执行结果。
type RollbackResult struct {
	AppliedAt             time.Time    `json:"applied_at"`
	StatePath             string       `json:"state_path"`
	SelectedBackup        BackupRecord `json:"selected_backup"`
	RestoredProfile       string       `json:"restored_profile"`
	RestoredPath          string       `json:"restored_path"`
	PreRollbackBackupPath string       `json:"pre_rollback_backup_path,omitempty"`
	VerifyResult          VerifyResult `json:"verify_result"`
}

// VerifyResult captures deterministic verification findings. / VerifyResult 描述确定性的验证结果。
type VerifyResult struct {
	Profile  string          `json:"profile"`
	Findings []VerifyFinding `json:"findings"`
}

// InspectResult shows detailed lifecycle state. / InspectResult 展示详细生命周期状态。
type InspectResult struct {
	CurrentProfile  string           `json:"current_profile"`
	Profiles        []ProfileSummary `json:"profiles"`
	InstallPath     string           `json:"install_path,omitempty"`
	ConfigPath      string           `json:"config_path,omitempty"`
	StatePath       string           `json:"state_path"`
	PlanRecordPath  string           `json:"plan_record_path"`
	LastApplyAt     *time.Time       `json:"last_apply_at,omitempty"`
	LastApplyResult string           `json:"last_apply_result"`
	Backups         []BackupRecord   `json:"backups"`
}

// StatusResult shows a compact lifecycle summary. / StatusResult 展示紧凑的生命周期摘要。
type StatusResult struct {
	CurrentProfile     string     `json:"current_profile"`
	LastApplyAt        *time.Time `json:"last_apply_at,omitempty"`
	LastApplyResult    string     `json:"last_apply_result"`
	LastRollbackAt     *time.Time `json:"last_rollback_at,omitempty"`
	LastRollbackResult string     `json:"last_rollback_result,omitempty"`
	BackupCount        int        `json:"backup_count"`
}

// RepairAction describes a deterministic remediation rule. / RepairAction 描述确定性的修复规则。
type RepairAction struct {
	Code         string `json:"code"`
	MessageKey   string `json:"message_key"`
	Risky        bool   `json:"risky"`
	RequiresFlag string `json:"requires_flag,omitempty"`
	Applied      bool   `json:"applied"`
}

// LogBundleMetadata describes a generated support bundle. / LogBundleMetadata 描述生成的日志包。
type LogBundleMetadata struct {
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	BundlePath  string    `json:"bundle_path"`
	FailureCode string    `json:"failure_code,omitempty"`
}

// LifecycleLogEntry describes a structured lifecycle log event. / LifecycleLogEntry 描述结构化生命周期日志事件。
type LifecycleLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command"`
	Step      string    `json:"step"`
	Result    string    `json:"result"`
	ErrorCode string    `json:"error_code,omitempty"`
	Profile   string    `json:"profile,omitempty"`
}

// LogsResult describes filtered log output and optional bundle metadata. / LogsResult 描述筛选后的日志输出以及可选归档元数据。
type LogsResult struct {
	LogPath string              `json:"log_path"`
	Entries []LifecycleLogEntry `json:"entries"`
	Bundle  *LogBundleMetadata  `json:"bundle,omitempty"`
}

// RepairResult describes deterministic repair planning and execution. / RepairResult 描述确定性修复计划与执行结果。
type RepairResult struct {
	Actions      []RepairAction `json:"actions"`
	AppliedCount int            `json:"applied_count"`
}

// DetectResult contains local environment facts. / DetectResult 包含本地环境事实信息。
type DetectResult struct {
	OS                string `json:"os"`
	Arch              string `json:"arch"`
	WorkingDir        string `json:"working_dir"`
	HomeDir           string `json:"home_dir"`
	Shell             string `json:"shell"`
	ExecutablePath    string `json:"executable_path"`
	OpenClawPath      string `json:"openclaw_path,omitempty"`
	OpenClawInstalled bool   `json:"openclaw_installed"`
	WorkspacePath     string `json:"workspace_path"`
	ProfilesPath      string `json:"profiles_path"`
	StatePath         string `json:"state_path"`
}

// DoctorResult contains deterministic health-check findings. / DoctorResult 包含确定性的健康检查结果。
type DoctorResult struct {
	Findings []DoctorFinding `json:"findings"`
}

// InitResult contains workspace initialization facts. / InitResult 包含工作区初始化结果。
type InitResult struct {
	WorkspacePath  string   `json:"workspace_path"`
	CreatedPaths   []string `json:"created_paths"`
	ExistingPaths  []string `json:"existing_paths"`
	DefaultProfile string   `json:"default_profile"`
	StatePath      string   `json:"state_path"`
}

// ProfileSummary is a compact profile listing row. / ProfileSummary 是简洁的配置概要行。
type ProfileSummary struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Active bool   `json:"active"`
}

// ProfileListResult describes available profiles. / ProfileListResult 描述可用配置列表。
type ProfileListResult struct {
	Profiles []ProfileSummary `json:"profiles"`
}

// ProfileShowResult describes one profile and its metadata. / ProfileShowResult 描述单个配置及其元数据。
type ProfileShowResult struct {
	Profile Profile `json:"profile"`
	Path    string  `json:"path"`
	Active  bool    `json:"active"`
}

// ProfileUseResult describes an updated active profile. / ProfileUseResult 描述已更新的活动配置。
type ProfileUseResult struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
