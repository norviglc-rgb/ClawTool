package i18n

import (
	"testing"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

func TestLoaderLocalizesPlaceholders(t *testing.T) {
	t.Parallel()

	loader := Loader{}
	bundle, err := loader.NewBundle()
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}

	localizer := loader.NewLocalizer(bundle, "en")
	text, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:    "command.stub.summary",
		TemplateData: map[string]any{"Command": "detect"},
	})
	if err != nil {
		t.Fatalf("localize: %v", err)
	}
	if text != "detect is scaffolded and ready for implementation." {
		t.Fatalf("unexpected text: %s", text)
	}
}
