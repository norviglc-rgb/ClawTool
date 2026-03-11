package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/openclaw/clawtool/internal/core"
)

// DefaultStatePath returns the default state file path. / DefaultStatePath 返回默认状态文件路径。
func DefaultStatePath(workDir string) string {
	return filepath.Join(workDir, ".clawtool", "state", "state.json")
}

// Load reads persisted state or returns a default value. / Load 读取持久化状态，若不存在则返回默认值。
func Load(path string) (core.StateRecord, error) {
	var record core.StateRecord

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return core.StateRecord{Version: "v1"}, nil
		}

		return record, &core.AppError{
			Code:       core.ErrorCodeStateRead,
			MessageKey: "error.state.read",
			Cause:      err,
		}
	}

	if err := json.Unmarshal(data, &record); err != nil {
		return record, &core.AppError{
			Code:       core.ErrorCodeStateRead,
			MessageKey: "error.state.read",
			Cause:      err,
		}
	}

	return record, nil
}

// Save persists state atomically enough for local CLI usage. / Save 以适合本地 CLI 的方式持久化状态。
func Save(path string, record core.StateRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeStateWrite,
			MessageKey: "error.state.write",
			Cause:      err,
		}
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeStateWrite,
			MessageKey: "error.state.write",
			Cause:      err,
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeStateWrite,
			MessageKey: "error.state.write",
			Cause:      err,
		}
	}

	return nil
}
