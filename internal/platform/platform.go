package platform

import (
	"runtime"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/openclaw/clawtool/internal/platform/darwin"
	"github.com/openclaw/clawtool/internal/platform/linux"
	"github.com/openclaw/clawtool/internal/platform/windows"
)

// Adapter hides platform-specific fact collection. / Adapter 隐藏平台相关的探测实现。
type Adapter interface {
	Detect(rootDir string) (core.DetectResult, error)
}

// Current returns the adapter for the current OS. / Current 返回当前操作系统对应的适配器。
func Current() Adapter {
	switch runtime.GOOS {
	case "darwin":
		return darwin.Adapter{}
	case "linux":
		return linux.Adapter{}
	case "windows":
		return windows.Adapter{}
	default:
		return linux.Adapter{}
	}
}
