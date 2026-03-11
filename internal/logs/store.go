package logs

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/openclaw/clawtool/internal/core"
)

// DefaultLogPath returns the lifecycle log path. / DefaultLogPath 返回生命周期日志路径。
func DefaultLogPath(workDir string) string {
	return filepath.Join(workDir, ".clawtool", "logs", "lifecycle.log")
}

// Append writes one structured lifecycle entry. / Append 写入一条结构化生命周期日志。
func Append(path string, entry core.LifecycleLogEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeLogWrite,
			MessageKey: "error.logs.write",
			Cause:      err,
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeLogWrite,
			MessageKey: "error.logs.write",
			Cause:      err,
		}
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeLogWrite,
			MessageKey: "error.logs.write",
			Cause:      err,
		}
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeLogWrite,
			MessageKey: "error.logs.write",
			Cause:      err,
		}
	}

	return nil
}

// Read loads structured lifecycle entries in file order. / Read 按文件顺序读取结构化生命周期日志。
func Read(path string) ([]core.LifecycleLogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, &core.AppError{
			Code:       core.ErrorCodeLogRead,
			MessageKey: "error.logs.read",
			Cause:      err,
		}
	}
	defer file.Close()

	entries := make([]core.LifecycleLogEntry, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry core.LifecycleLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, &core.AppError{
				Code:       core.ErrorCodeLogRead,
				MessageKey: "error.logs.read",
				Cause:      err,
			}
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, &core.AppError{
			Code:       core.ErrorCodeLogRead,
			MessageKey: "error.logs.read",
			Cause:      err,
		}
	}

	return entries, nil
}
