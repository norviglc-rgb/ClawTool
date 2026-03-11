package common

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	workspaceDirName = ".clawtool"
	profilesDirName  = "profiles"
)

// Facts holds shared platform detection results. / Facts 保存共享的平台探测结果。
type Facts struct {
	WorkingDir     string
	HomeDir        string
	ExecutablePath string
	Shell          string
	OpenClawPath   string
}

// DetectFacts collects portable runtime facts. / DetectFacts 收集可移植的运行时信息。
func DetectFacts(rootDir string) (Facts, error) {
	workingDir, err := filepath.Abs(rootDir)
	if err != nil {
		return Facts{}, err
	}

	homeDir, _ := os.UserHomeDir()
	executablePath, _ := os.Executable()
	openClawPath, _ := exec.LookPath("openclaw")

	return Facts{
		WorkingDir:     workingDir,
		HomeDir:        homeDir,
		ExecutablePath: executablePath,
		Shell:          detectShell(),
		OpenClawPath:   openClawPath,
	}, nil
}

// WorkspacePath returns the managed workspace path. / WorkspacePath 返回托管工作区路径。
func WorkspacePath(rootDir string) string {
	return filepath.Join(rootDir, workspaceDirName)
}

// ProfilesPath returns the managed profiles directory path. / ProfilesPath 返回托管配置目录路径。
func ProfilesPath(rootDir string) string {
	return filepath.Join(WorkspacePath(rootDir), profilesDirName)
}

func detectShell() string {
	for _, key := range []string{"SHELL", "COMSPEC"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
