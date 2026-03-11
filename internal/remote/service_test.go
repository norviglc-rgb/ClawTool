package remote

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openclaw/clawtool/internal/core"
)

func TestLoadProfilePlanVerifyExecAndApply(t *testing.T) {
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

	executor := &fakeExecutor{
		executeResults: []ExecOutput{
			{Stdout: "hello", ExitCode: 0, Duration: 150 * time.Millisecond},
			{ExitCode: 0},
			{ExitCode: 0},
			{ExitCode: 0},
			{ExitCode: 0},
		},
	}
	service := NewServiceWithExecutor(rootDir, executor)

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

	applyResult, err := service.Apply(context.Background(), profile)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if applyResult.RemoteConfigPath != "/etc/openclaw/config.yaml" {
		t.Fatalf("unexpected remote config path: %+v", applyResult)
	}
	if len(executor.writeRequests) != 1 {
		t.Fatalf("expected one remote write, got %d", len(executor.writeRequests))
	}
	if !bytes.Contains(executor.writeRequests[0].data, []byte("name: remote")) {
		t.Fatalf("unexpected remote write payload: %s", string(executor.writeRequests[0].data))
	}
	assertRemoteFinding(t, applyResult.VerifyResult.Findings, "remote.verify.remote_config", core.SeverityPass)
	assertRemoteFinding(t, applyResult.VerifyResult.Findings, "remote.verify.openclaw", core.SeverityPass)
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
	executeResults []ExecOutput
	executeErr     error
	writeErr       error
	writeRequests  []fakeWriteRequest
}

func (f *fakeExecutor) Execute(_ context.Context, _ ConnectionOptions, _ string) (ExecOutput, error) {
	if f.executeErr != nil {
		return ExecOutput{}, f.executeErr
	}
	if len(f.executeResults) == 0 {
		return ExecOutput{}, errors.New("unexpected execute call")
	}
	result := f.executeResults[0]
	f.executeResults = f.executeResults[1:]
	return result, nil
}

func (f *fakeExecutor) WriteFile(_ context.Context, _ ConnectionOptions, path string, data []byte, mode string) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.writeRequests = append(f.writeRequests, fakeWriteRequest{
		path: path,
		data: append([]byte(nil), data...),
		mode: mode,
	})
	return nil
}

type fakeWriteRequest struct {
	path string
	data []byte
	mode string
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
