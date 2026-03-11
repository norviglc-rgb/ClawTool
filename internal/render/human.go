package render

import (
	"fmt"
	"io"

	"github.com/openclaw/clawtool/internal/core"
)

// LocalizeFunc resolves a message key to human text. / LocalizeFunc 将消息键解析为可读文本。
type LocalizeFunc func(key string, data map[string]any) string

// Renderer renders a command result. / Renderer 渲染命令结果。
type Renderer interface {
	Render(w io.Writer, result core.CommandResult) error
}

// HumanRenderer prints localized human-readable output. / HumanRenderer 输出本地化的人类可读文本。
type HumanRenderer struct {
	Localize LocalizeFunc
}

// Render writes a compact summary and detail list. / Render 输出简洁摘要和明细列表。
func (r HumanRenderer) Render(w io.Writer, result core.CommandResult) error {
	summary := result.SummaryKey
	if r.Localize != nil {
		summary = r.Localize(result.SummaryKey, map[string]any{"Command": result.Command})
	}

	if _, err := fmt.Fprintln(w, summary); err != nil {
		return err
	}

	for _, detail := range result.Details {
		label := detail.Key
		if r.Localize != nil {
			label = r.Localize("detail."+detail.Key, nil)
		}

		if _, err := fmt.Fprintf(w, "- %s: %s\n", label, detail.Value); err != nil {
			return err
		}
	}

	return nil
}
