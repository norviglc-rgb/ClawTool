package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	rootassets "github.com/openclaw/clawtool"
	"github.com/openclaw/clawtool/internal/core"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Loader loads embedded message catalogs. / Loader 加载嵌入式消息目录。
type Loader struct{}

// NewBundle creates an i18n bundle from embedded locale files. / NewBundle 从嵌入式语言文件创建 i18n bundle。
func (Loader) NewBundle() (*goi18n.Bundle, error) {
	bundle := goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := fs.Glob(rootassets.EmbeddedFiles, "locales/*.json")
	if err != nil {
		return nil, &core.AppError{
			Code:       core.ErrorCodeI18NLoad,
			MessageKey: "error.i18n.load",
			Cause:      err,
		}
	}

	for _, entry := range entries {
		if _, err := bundle.LoadMessageFileFS(rootassets.EmbeddedFiles, entry); err != nil {
			return nil, &core.AppError{
				Code:       core.ErrorCodeI18NLoad,
				MessageKey: "error.i18n.load",
				Cause:      fmt.Errorf("%s: %w", entry, err),
			}
		}
	}

	return bundle, nil
}

// NewLocalizer creates a localizer for a resolved language. / NewLocalizer 为解析后的语言创建本地化器。
func (Loader) NewLocalizer(bundle *goi18n.Bundle, lang string) *goi18n.Localizer {
	if strings.TrimSpace(lang) == "" {
		lang = "en"
	}
	return goi18n.NewLocalizer(bundle, lang, "en")
}
