package platform

import (
	"runtime"
	"testing"
)

func TestCurrentReturnsAdapterForRuntime(t *testing.T) {
	t.Parallel()

	adapter := Current()
	if adapter == nil {
		t.Fatal("expected platform adapter")
	}

	result, err := adapter.Detect(t.TempDir())
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if result.OS != runtime.GOOS {
		t.Fatalf("unexpected os: %+v", result)
	}
	if result.WorkingDir == "" || result.WorkspacePath == "" || result.ProfilesPath == "" {
		t.Fatalf("missing platform paths: %+v", result)
	}
}
