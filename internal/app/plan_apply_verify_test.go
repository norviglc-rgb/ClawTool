package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/openclaw/clawtool/internal/state"
)

func TestPlanApplyVerifyLifecycle(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	if _, err := service.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	plan, err := service.Plan()
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if plan.Profile != "default" {
		t.Fatalf("unexpected profile: %+v", plan)
	}
	if filepath.Base(plan.GeneratedConfig) != "effective-default.yaml" {
		t.Fatalf("unexpected generated config path: %s", plan.GeneratedConfig)
	}
	if len(plan.Changes) == 0 || plan.Changes[0].Action != "create" {
		t.Fatalf("unexpected plan changes: %+v", plan.Changes)
	}
	if len(plan.ContentDiffs) == 0 {
		t.Fatal("expected content diffs")
	}

	applyResult, err := service.Apply()
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !applyResult.Changed {
		t.Fatal("expected first apply to write config")
	}
	if _, err := os.Stat(applyResult.GeneratedConfig); err != nil {
		t.Fatalf("generated config missing: %v", err)
	}
	if _, err := os.Stat(applyResult.PlanRecordPath); err != nil {
		t.Fatalf("plan record missing: %v", err)
	}

	verifyResult, err := service.Verify()
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if len(verifyResult.Findings) == 0 {
		t.Fatal("expected verify findings")
	}

	secondPlan, err := service.Plan()
	if err != nil {
		t.Fatalf("second plan: %v", err)
	}
	if secondPlan.Changes[0].Action != "noop" {
		t.Fatalf("expected noop change, got %+v", secondPlan.Changes)
	}
	if len(secondPlan.ContentDiffs) != 0 {
		t.Fatalf("expected no content diffs on noop plan, got %+v", secondPlan.ContentDiffs)
	}

	secondApply, err := service.Apply()
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if secondApply.Changed {
		t.Fatal("expected second apply to be idempotent")
	}
	if secondApply.BackupPath != "" {
		t.Fatalf("expected no backup on noop apply, got %s", secondApply.BackupPath)
	}
}

func TestApplyCreatesBackupWhenConfigChanges(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	if _, err := service.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := service.Apply(); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if _, err := service.CreateProfile(core.Profile{
		Version: "v1",
		Name:    "remote",
		Target: core.ProfileTarget{
			Kind:    "ssh",
			Address: "127.0.0.1",
		},
	}); err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if _, err := service.UseProfile("remote"); err != nil {
		t.Fatalf("use profile: %v", err)
	}

	result, err := service.Apply()
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if !result.Changed {
		t.Fatal("expected apply to rewrite config after profile change")
	}
	if result.BackupPath == "" {
		t.Fatal("expected backup path")
	}
	if _, err := os.Stat(result.BackupPath); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}

func TestInspectAndStatus(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	if _, err := service.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := service.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	inspectResult, err := service.Inspect()
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if inspectResult.CurrentProfile != "default" {
		t.Fatalf("unexpected inspect profile: %+v", inspectResult)
	}
	if inspectResult.PlanRecordPath == "" || inspectResult.StatePath == "" {
		t.Fatalf("unexpected inspect paths: %+v", inspectResult)
	}

	statusResult, err := service.Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if statusResult.CurrentProfile != "default" {
		t.Fatalf("unexpected status: %+v", statusResult)
	}
	if statusResult.LastApplyResult != "success" {
		t.Fatalf("unexpected last apply result: %+v", statusResult)
	}
}

func TestRollbackRestoresLatestBackup(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	if _, err := service.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := service.Apply(); err != nil {
		t.Fatalf("apply default: %v", err)
	}
	if _, err := service.CreateProfile(core.Profile{
		Version: "v1",
		Name:    "remote",
		Target: core.ProfileTarget{
			Kind:    "ssh",
			Address: "127.0.0.1",
		},
	}); err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if _, err := service.UseProfile("remote"); err != nil {
		t.Fatalf("use remote: %v", err)
	}
	if _, err := service.Apply(); err != nil {
		t.Fatalf("apply remote: %v", err)
	}

	result, err := service.Rollback("")
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if result.RestoredProfile != "default" {
		t.Fatalf("unexpected restored profile: %+v", result)
	}
	if result.SelectedBackup.ProfileName != "default" {
		t.Fatalf("unexpected selected backup: %+v", result.SelectedBackup)
	}
	if result.PreRollbackBackupPath == "" {
		t.Fatal("expected pre-rollback backup path")
	}
	if _, err := os.Stat(result.RestoredPath); err != nil {
		t.Fatalf("restored path missing: %v", err)
	}

	record, err := state.Load(state.DefaultStatePath(rootDir))
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if record.CurrentProfile != "default" {
		t.Fatalf("unexpected current profile after rollback: %+v", record)
	}
	if record.LastRollbackResult != "success" || record.LastRollbackAt == nil {
		t.Fatalf("unexpected rollback state: %+v", record)
	}
}

func TestRollbackWithoutBackupFails(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	if _, err := service.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := service.Rollback("")
	if err == nil {
		t.Fatal("expected rollback to fail without backup")
	}

	var appErr *core.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %T", err)
	}
	if appErr.Code != core.ErrorCodeBackupNotFound {
		t.Fatalf("unexpected error code: %v", appErr.Code)
	}
}

func TestLogsSupportTailSinceAndBundle(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	if _, err := service.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := service.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	since := time.Now().UTC().Add(-1 * time.Hour)
	result, err := service.Logs(2, &since, true)
	if err != nil {
		t.Fatalf("logs: %v", err)
	}
	if len(result.Entries) == 0 {
		t.Fatal("expected log entries")
	}
	if len(result.Entries) > 2 {
		t.Fatalf("unexpected entry count: %d", len(result.Entries))
	}
	if result.Bundle == nil || result.Bundle.BundlePath == "" {
		t.Fatalf("expected bundle metadata: %+v", result.Bundle)
	}
	if _, err := os.Stat(result.Bundle.BundlePath); err != nil {
		t.Fatalf("bundle missing: %v", err)
	}
}

func TestRepairPlansAndExecutesDeterministicActions(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	planned, err := service.Repair(false, false)
	if err != nil {
		t.Fatalf("repair plan: %v", err)
	}
	if len(planned.Actions) != 1 || planned.Actions[0].Code != "repair.workspace.init" {
		t.Fatalf("unexpected repair plan: %+v", planned)
	}
	if planned.AppliedCount != 0 {
		t.Fatalf("unexpected applied count: %+v", planned)
	}

	appliedSafe, err := service.Repair(true, false)
	if err != nil {
		t.Fatalf("repair apply safe: %v", err)
	}
	if appliedSafe.AppliedCount != 1 || !appliedSafe.Actions[0].Applied {
		t.Fatalf("expected safe repair to apply: %+v", appliedSafe)
	}

	riskyPlan, err := service.Repair(false, false)
	if err != nil {
		t.Fatalf("repair risky plan: %v", err)
	}
	if len(riskyPlan.Actions) != 1 || riskyPlan.Actions[0].Code != "repair.generated_config.apply" {
		t.Fatalf("unexpected risky repair plan: %+v", riskyPlan)
	}
	if riskyPlan.Actions[0].Applied {
		t.Fatalf("risky action should not auto-apply: %+v", riskyPlan)
	}

	riskyApply, err := service.Repair(false, true)
	if err != nil {
		t.Fatalf("repair risky apply: %v", err)
	}
	if riskyApply.AppliedCount != 1 || !riskyApply.Actions[0].Applied {
		t.Fatalf("expected risky repair to apply with yes: %+v", riskyApply)
	}
}
