package render

import (
	"encoding/json"
	"io"

	"github.com/openclaw/clawtool/internal/core"
)

// JSONRenderer writes stable locale-neutral JSON. / JSONRenderer 输出稳定且与语言无关的 JSON。
type JSONRenderer struct{}

// Render writes indented JSON with deterministic field names. / Render 输出带缩进且字段稳定的 JSON。
func (JSONRenderer) Render(w io.Writer, result core.CommandResult) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
