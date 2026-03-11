package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openclaw/clawtool/internal/app"
	"github.com/openclaw/clawtool/internal/core"
	"github.com/openclaw/clawtool/internal/remote"
	"github.com/openclaw/clawtool/internal/render"
)

func TestRootCommandSmokeJSON(t *testing.T) {
	t.Parallel()

	cmd := NewRootCommand()
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"detect", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "\"command\": \"detect\"") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "\"openclaw_installed\"") {
		t.Fatalf("missing detect detail: %s", output)
	}
}

func TestDetectLocalizedHumanOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		lang         string
		summary      string
		detailLabel  string
		installLabel string
	}{
		{name: "en", lang: "en", summary: "Local environment detection completed.", detailLabel: "Architecture", installLabel: "OpenClaw Installed"},
		{name: "zh-CN", lang: "zh-CN", summary: "本地环境检测已完成。", detailLabel: "架构", installLabel: "OpenClaw 已安装"},
		{name: "ja", lang: "ja", summary: "ローカル環境の検出が完了しました。", detailLabel: "アーキテクチャ", installLabel: "OpenClaw 導入済み"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewRootCommand()
			buffer := &bytes.Buffer{}
			cmd.SetOut(buffer)
			cmd.SetErr(buffer)
			cmd.SetArgs([]string{"--lang", tc.lang, "detect"})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute: %v", err)
			}

			output := strings.ReplaceAll(buffer.String(), "\r\n", "\n")
			if !strings.Contains(output, tc.summary) {
				t.Fatalf("missing summary %q in output:\n%s", tc.summary, output)
			}
			if !strings.Contains(output, tc.detailLabel) {
				t.Fatalf("missing detail label %q in output:\n%s", tc.detailLabel, output)
			}
			if !strings.Contains(output, tc.installLabel) {
				t.Fatalf("missing install label %q in output:\n%s", tc.installLabel, output)
			}
		})
	}
}

func TestPlanOutAndShow(t *testing.T) {
	t.Parallel()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	planPath := filepath.Join(tempDir, "plan.json")

	initCmd := NewRootCommand()
	initBuffer := &bytes.Buffer{}
	initCmd.SetOut(initBuffer)
	initCmd.SetErr(initBuffer)
	initCmd.SetArgs([]string{"init"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init execute: %v", err)
	}

	planCmd := NewRootCommand()
	planBuffer := &bytes.Buffer{}
	planCmd.SetOut(planBuffer)
	planCmd.SetErr(planBuffer)
	planCmd.SetArgs([]string{"plan", "--out", planPath})
	if err := planCmd.Execute(); err != nil {
		t.Fatalf("plan execute: %v", err)
	}
	if _, err := os.Stat(planPath); err != nil {
		t.Fatalf("plan file missing: %v", err)
	}

	showCmd := NewRootCommand()
	showBuffer := &bytes.Buffer{}
	showCmd.SetOut(showBuffer)
	showCmd.SetErr(showBuffer)
	showCmd.SetArgs([]string{"show", planPath})
	if err := showCmd.Execute(); err != nil {
		t.Fatalf("show execute: %v", err)
	}

	output := strings.ReplaceAll(showBuffer.String(), "\r\n", "\n")
	if !strings.Contains(output, "Saved plan loaded successfully.") {
		t.Fatalf("missing show summary in output:\n%s", output)
	}
	if !strings.Contains(output, "Plan File") {
		t.Fatalf("missing plan file detail in output:\n%s", output)
	}
	if !strings.Contains(output, "Ensure workspace directories exist.") {
		t.Fatalf("missing localized plan step detail in output:\n%s", output)
	}
}

func TestVerifyReturnsSilentExitErrorOnFailure(t *testing.T) {
	t.Parallel()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	initCmd := NewRootCommand()
	initBuffer := &bytes.Buffer{}
	initCmd.SetOut(initBuffer)
	initCmd.SetErr(initBuffer)
	initCmd.SetArgs([]string{"init"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init execute: %v", err)
	}

	verifyCmd := NewRootCommand()
	verifyBuffer := &bytes.Buffer{}
	verifyCmd.SetOut(verifyBuffer)
	verifyCmd.SetErr(verifyBuffer)
	verifyCmd.SetArgs([]string{"verify"})
	err = verifyCmd.Execute()
	if err == nil {
		t.Fatal("expected verify to return exit error")
	}

	var exitErr *core.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exit error, got %T", err)
	}
	if exitErr.Code != 1 || !exitErr.Silent {
		t.Fatalf("unexpected exit error: %+v", exitErr)
	}

	output := strings.ReplaceAll(verifyBuffer.String(), "\r\n", "\n")
	if !strings.Contains(output, "Verification completed.") {
		t.Fatalf("missing verify summary in output:\n%s", output)
	}
}

func TestRemotePlanAndVerifySmoke(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	service := app.NewService(tempDir)
	if _, err := service.Init(); err != nil {
		t.Fatalf("init service: %v", err)
	}
	if _, err := service.CreateProfile(core.Profile{
		Version: "v1",
		Name:    "remote",
		Target: core.ProfileTarget{
			Kind:            "ssh",
			Address:         "ssh.example.internal",
			User:            "deploy",
			KeyPath:         filepath.Join(tempDir, "id_test"),
			HostKeyStrategy: "insecure",
		},
	}); err != nil {
		t.Fatalf("create profile: %v", err)
	}

	localize := func(key string, data map[string]any) string {
		switch key {
		case "cmd.remote.short":
			return "Run remote lifecycle preparation commands."
		case "cmd.remote.plan.short":
			return "Create a deterministic remote execution plan."
		case "cmd.remote.verify.short":
			return "Run deterministic remote verification preflight."
		case "cmd.remote.exec.short":
			return "Run one remote command through SSH."
		case "remote.plan.summary.ready":
			return "Remote execution plan generated successfully."
		case "remote.verify.summary.ready":
			return "Remote verification preflight completed."
		case "remote.exec.summary.ready":
			return "Remote command execution completed."
		case "detail.profile":
			return "Profile"
		case "detail.target_address":
			return "Target Address"
		case "detail.target_user":
			return "Target User"
		case "detail.target_port":
			return "Target Port"
		case "detail.target_key_path":
			return "Target Key Path"
		case "detail.host_key_strategy":
			return "Host Key Strategy"
		case "detail.requires_backup":
			return "Requires Backup"
		case "detail.verification_steps":
			return "Verification Steps"
		case "detail.plan_steps":
			return "Plan Steps"
		case "detail.plan_step_details":
			return "Plan Step Details"
		case "detail.changes":
			return "Changes"
		case "detail.command":
			return "Command"
		case "detail.exit_code":
			return "Exit Code"
		case "detail.stdout":
			return "Stdout"
		case "detail.stderr":
			return "Stderr"
		case "remote.step.validate_profile":
			return "Validate remote profile remote."
		case "remote.step.prepare_connection":
			return "Prepare SSH connection settings for ssh.example.internal."
		case "remote.step.sync_config":
			return "Plan remote config sync for profile remote."
		case "remote.step.verify_target":
			return "Plan verification checks for remote target ssh.example.internal."
		case "remote.verify.finding.profile.valid":
			return "Remote profile passed validation."
		case "remote.verify.finding.target_kind.valid":
			return "Remote target kind is ssh."
		case "remote.verify.finding.target_address.valid":
			return "Remote target address is configured."
		case "remote.verify.finding.target_user.valid":
			return "Remote target user is configured."
		case "remote.verify.finding.key_path.valid":
			return "Remote private key path is configured."
		case "remote.verify.finding.host_key.insecure":
			return "Remote host key verification uses insecure mode."
		default:
			return key
		}
	}

	remoteService := remote.NewServiceWithExecutor(tempDir, fakeRemoteExecutor{
		output: remote.ExecOutput{Stdout: "hello", ExitCode: 0},
	})

	planCmd := newRemoteCommand(localize)
	planBuffer := &bytes.Buffer{}
	planCmd.SetOut(planBuffer)
	planCmd.SetErr(planBuffer)
	planCmd.SetContext(withRuntime(context.Background(), Runtime{
		Localize: localize,
		Renderer: render.HumanRenderer{Localize: localize},
		Service:  service,
		Remote:   remoteService,
	}))
	planCmd.SetArgs([]string{"plan", "remote"})
	if err := planCmd.Execute(); err != nil {
		t.Fatalf("remote plan execute: %v", err)
	}
	planOutput := strings.ReplaceAll(planBuffer.String(), "\r\n", "\n")
	if !strings.Contains(planOutput, "Remote execution plan generated successfully.") {
		t.Fatalf("missing remote plan summary:\n%s", planOutput)
	}
	if !strings.Contains(planOutput, "ssh.example.internal") {
		t.Fatalf("missing remote target address:\n%s", planOutput)
	}

	verifyCmd := newRemoteCommand(localize)
	verifyBuffer := &bytes.Buffer{}
	verifyCmd.SetOut(verifyBuffer)
	verifyCmd.SetErr(verifyBuffer)
	verifyCmd.SetContext(withRuntime(context.Background(), Runtime{
		Localize: localize,
		Renderer: render.HumanRenderer{Localize: localize},
		Service:  service,
		Remote:   remoteService,
	}))
	verifyCmd.SetArgs([]string{"verify", "remote"})
	if err := verifyCmd.Execute(); err != nil {
		t.Fatalf("remote verify execute: %v", err)
	}
	verifyOutput := strings.ReplaceAll(verifyBuffer.String(), "\r\n", "\n")
	if !strings.Contains(verifyOutput, "Remote verification preflight completed.") {
		t.Fatalf("missing remote verify summary:\n%s", verifyOutput)
	}

	execCmd := newRemoteCommand(localize)
	execBuffer := &bytes.Buffer{}
	execCmd.SetOut(execBuffer)
	execCmd.SetErr(execBuffer)
	execCmd.SetContext(withRuntime(context.Background(), Runtime{
		Localize: localize,
		Renderer: render.HumanRenderer{Localize: localize},
		Service:  service,
		Remote:   remoteService,
	}))
	execCmd.SetArgs([]string{"exec", "remote", "echo", "hello"})
	if err := execCmd.Execute(); err != nil {
		t.Fatalf("remote exec execute: %v", err)
	}
	execOutput := strings.ReplaceAll(execBuffer.String(), "\r\n", "\n")
	if !strings.Contains(execOutput, "Remote command execution completed.") {
		t.Fatalf("missing remote exec summary:\n%s", execOutput)
	}
	if !strings.Contains(execOutput, "hello") {
		t.Fatalf("missing remote exec stdout:\n%s", execOutput)
	}
}

type fakeRemoteExecutor struct {
	output remote.ExecOutput
	err    error
}

func (f fakeRemoteExecutor) Execute(context.Context, remote.ConnectionOptions, string) (remote.ExecOutput, error) {
	return f.output, f.err
}
