package app

import (
	"path/filepath"
	"testing"
)

func TestWorkspaceInitAndProfiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service := NewService(rootDir)

	initResult, err := service.Init()
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if initResult.WorkspacePath == "" {
		t.Fatal("expected workspace path")
	}

	listResult, err := service.ListProfiles()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listResult.Profiles) != 1 || listResult.Profiles[0].Name != "default" {
		t.Fatalf("unexpected profiles: %+v", listResult.Profiles)
	}

	showResult, err := service.ShowProfile("default")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if filepath.Base(showResult.Path) != "default.yaml" {
		t.Fatalf("unexpected profile path: %s", showResult.Path)
	}

	useResult, err := service.UseProfile("default")
	if err != nil {
		t.Fatalf("use: %v", err)
	}
	if useResult.Name != "default" {
		t.Fatalf("unexpected active profile: %+v", useResult)
	}
}
