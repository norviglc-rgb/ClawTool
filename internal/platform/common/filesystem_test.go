package common

import (
	"path/filepath"
	"testing"
)

func TestEnsureDirectoryAndDirectoryExists(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "workspace")
	if DirectoryExists(path) {
		t.Fatal("directory should not exist yet")
	}

	created, err := EnsureDirectory(path)
	if err != nil {
		t.Fatalf("ensure directory: %v", err)
	}
	if !created {
		t.Fatal("expected first ensure to create directory")
	}
	if !DirectoryExists(path) {
		t.Fatal("directory should exist after ensure")
	}

	created, err = EnsureDirectory(path)
	if err != nil {
		t.Fatalf("ensure directory second time: %v", err)
	}
	if created {
		t.Fatal("expected second ensure to report existing directory")
	}
}

func TestDirectoryWritable(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "writable")
	if err := DirectoryWritable(path); err != nil {
		t.Fatalf("directory writable: %v", err)
	}
	if !DirectoryExists(path) {
		t.Fatal("directory should exist after writable check")
	}
}
