package cli

import (
	"context"
	"io"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/openclaw/clawtool/internal/app"
	"github.com/openclaw/clawtool/internal/config"
	"github.com/openclaw/clawtool/internal/core"
	"github.com/openclaw/clawtool/internal/remote"
	"github.com/openclaw/clawtool/internal/render"
)

type runtimeKey struct{}

// Runtime provides resolved execution dependencies. / Runtime 提供解析后的执行依赖。
type Runtime struct {
	Config   config.RuntimeConfig
	Localize func(key string, data map[string]any) string
	Renderer render.Renderer
	Service  app.Service
	Remote   remote.Service
}

func withRuntime(ctx context.Context, value Runtime) context.Context {
	return context.WithValue(ctx, runtimeKey{}, value)
}

func runtimeFromContext(ctx context.Context) Runtime {
	runtime, _ := ctx.Value(runtimeKey{}).(Runtime)
	return runtime
}

func localizeFunc(localizer *goi18n.Localizer) func(string, map[string]any) string {
	return func(key string, data map[string]any) string {
		message, err := localizer.Localize(&goi18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: data,
		})
		if err != nil {
			return key
		}
		return message
	}
}

func renderResult(cmdContext context.Context, writer io.Writer, result core.CommandResult) error {
	runtime := runtimeFromContext(cmdContext)
	return runtime.Renderer.Render(writer, result)
}

func severityRank(severity core.Severity) int {
	switch severity {
	case core.SeverityFail:
		return 3
	case core.SeverityWarn:
		return 2
	default:
		return 1
	}
}

func verifyResultValue(findings []core.VerifyFinding) string {
	result := "success"
	for _, finding := range findings {
		switch finding.Severity {
		case core.SeverityFail:
			return "fail"
		case core.SeverityWarn:
			result = "warn"
		}
	}
	return result
}

func findingString(localize func(string, map[string]any) string, finding core.VerifyFinding) string {
	text := finding.MessageKey
	if localize != nil {
		text = localize(finding.MessageKey, nil)
	}
	return string(finding.Severity) + " " + text
}
