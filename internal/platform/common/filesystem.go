package common

import (
	"errors"
	"os"
	"path/filepath"
)

// EnsureDirectory creates a directory if needed and reports whether it was created. / EnsureDirectory 在需要时创建目录并返回是否新建。
func EnsureDirectory(path string) (bool, error) {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return false, nil
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, err
	}
	return true, nil
}

// DirectoryExists reports whether a directory exists. / DirectoryExists 返回目录是否存在。
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// DirectoryWritable verifies that a directory can be created and written. / DirectoryWritable 校验目录是否可创建且可写。
func DirectoryWritable(path string) error {
	if _, err := EnsureDirectory(path); err != nil {
		return err
	}

	testPath := filepath.Join(path, ".clawtool-write-check")
	if err := os.WriteFile(testPath, []byte("ok"), 0o644); err != nil {
		return err
	}
	if err := os.Remove(testPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
