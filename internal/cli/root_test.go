package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
}
