package i18n

import (
	"os"
	"strings"

	"github.com/openclaw/clawtool/internal/config"
	"golang.org/x/text/language"
)

// Resolver resolves locale in the required precedence order. / Resolver 按要求的优先级解析语言。
type Resolver struct{}

// Resolve resolves locale from flag, env, config, and system settings. / Resolve 从 flag、环境、配置和系统设置解析语言。
func (Resolver) Resolve(flagValue string, cfg config.RuntimeConfig) string {
	if value := normalizeLanguage(flagValue); value != "" {
		return value
	}

	if value := normalizeLanguage(os.Getenv("CLAWTOOL_LANG")); value != "" {
		return value
	}

	if value := normalizeLanguage(cfg.Language); value != "" {
		return value
	}

	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := normalizeLanguage(os.Getenv(key)); value != "" {
			return value
		}
	}

	return "en"
}

func normalizeLanguage(raw string) string {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return ""
	}

	cleaned = strings.ReplaceAll(cleaned, "_", "-")
	cleaned = strings.Split(cleaned, ".")[0]

	tag, err := language.Parse(cleaned)
	if err != nil {
		return ""
	}

	base, _ := tag.Base()
	region, _ := tag.Region()
	switch {
	case base.String() == "zh" && region.String() == "CN":
		return "zh-CN"
	case base.String() == "ja":
		return "ja"
	case base.String() == "en":
		return "en"
	case base.String() == "zh":
		return "zh-CN"
	default:
		return "en"
	}
}
