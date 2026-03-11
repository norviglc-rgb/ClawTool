package remote

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openclaw/clawtool/internal/core"
)

func TestLoadProfilePlanVerifyAndExec(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	profilesDir := filepath.Join(rootDir, ".clawtool", "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("mkdir profiles: %v", err)
	}

	data := []byte("version: v1\nname: remote\ntarget:\n  kind: ssh\n  address: ssh.example.internal\n  user: deploy\n  port: 2222\n  key_path: C:/keys/demo\n  host_key_strategy: insecure\n")
	if err := os.WriteFile(filepath.Join(profilesDir, "remote.yaml"), data, 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	service := NewServiceWithExecutor(rootDir, fakeExecutor{
		output: ExecOutput{
			Stdout:   "hello",
			ExitCode: 0,
			Duration: 150 * time.Millisecond,
		},
	})

	profile, path, err := service.LoadProfile("remote")
	if err != nil {
		t.Fatalf("load profile: %v", err)
	}
	if filepath.Base(path) != "remote.yaml" {
		t.Fatalf("unexpected path: %s", path)
	}

	plan := service.Plan(profile)
	if len(plan.Steps) != 4 {
		t.Fatalf("unexpected steps: %+v", plan.Steps)
	}
	if len(plan.Changes) != 2 {
		t.Fatalf("unexpected changes: %+v", plan.Changes)
	}

	verifyResult := service.Verify(profile)
	if verifyResult.Profile != "remote" {
		t.Fatalf("unexpected verify profile: %+v", verifyResult)
	}
	assertRemoteFinding(t, verifyResult.Findings, "remote.verify.target_user", core.SeverityPass)
	assertRemoteFinding(t, verifyResult.Findings, "remote.verify.key_path", core.SeverityPass)
	assertRemoteFinding(t, verifyResult.Findings, "remote.verify.host_key", core.SeverityWarn)

	execResult, err := service.Exec(context.Background(), profile, "echo hello")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if execResult.User != "deploy" || execResult.Port != 2222 {
		t.Fatalf("unexpected exec result: %+v", execResult)
	}
	if execResult.Stdout != "hello" || execResult.ExitCode != 0 {
		t.Fatalf("unexpected exec output: %+v", execResult)
	}
}

func TestLoadProfileRejectsNonSSHProfile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	profilesDir := filepath.Join(rootDir, ".clawtool", "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("mkdir profiles: %v", err)
	}

	data := []byte("version: v1\nname: local\ntarget:\n  kind: local\n")
	if err := os.WriteFile(filepath.Join(profilesDir, "local.yaml"), data, 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	service := NewService(rootDir)
	if _, _, err := service.LoadProfile("local"); err == nil {
		t.Fatal("expected non-ssh profile to be rejected")
	}
}

type fakeExecutor struct {
	output ExecOutput
	err    error
}

func (f fakeExecutor) Execute(_ context.Context, _ ConnectionOptions, _ string) (ExecOutput, error) {
	return f.output, f.err
}

func assertRemoteFinding(t *testing.T, findings []core.VerifyFinding, code string, severity core.Severity) {
	t.Helper()

	for _, finding := range findings {
		if finding.Code == code {
			if finding.Severity != severity {
				t.Fatalf("unexpected severity for %s: %+v", code, finding)
			}
			return
		}
	}

	t.Fatalf("missing remote finding %s in %+v", code, findings)
}
