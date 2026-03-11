package windows

import (
	"runtime"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/openclaw/clawtool/internal/platform/common"
	"github.com/openclaw/clawtool/internal/state"
)

// Adapter collects Windows-specific facts. / Adapter 收集 Windows 平台信息。
type Adapter struct{}

// Detect collects local Windows environment facts. / Detect 收集本地 Windows 环境事实。
func (Adapter) Detect(rootDir string) (core.DetectResult, error) {
	facts, err := common.DetectFacts(rootDir)
	if err != nil {
		return core.DetectResult{}, err
	}

	return core.DetectResult{
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		WorkingDir:        facts.WorkingDir,
		HomeDir:           facts.HomeDir,
		Shell:             facts.Shell,
		ExecutablePath:    facts.ExecutablePath,
		OpenClawPath:      facts.OpenClawPath,
		OpenClawInstalled: facts.OpenClawPath != "",
		WorkspacePath:     common.WorkspacePath(rootDir),
		ProfilesPath:      common.ProfilesPath(rootDir),
		StatePath:         state.DefaultStatePath(rootDir),
	}, nil
}
