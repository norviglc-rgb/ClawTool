package i18n

import (
	"testing"

	"github.com/openclaw/clawtool/internal/config"
)

func TestResolverResolvePrecedence(t *testing.T) {
	t.Setenv("CLAWTOOL_LANG", "ja")
	t.Setenv("LANG", "zh_CN.UTF-8")

	resolver := Resolver{}
	got := resolver.Resolve("en", config.RuntimeConfig{Language: "zh-CN"})
	if got != "en" {
		t.Fatalf("expected en, got %s", got)
	}
}

func TestResolverFallback(t *testing.T) {
	t.Setenv("CLAWTOOL_LANG", "")
	t.Setenv("LANG", "")

	resolver := Resolver{}
	got := resolver.Resolve("", config.RuntimeConfig{})
	if got != "en" {
		t.Fatalf("expected en, got %s", got)
	}
}

func TestResolverSystemLocale(t *testing.T) {
	t.Setenv("CLAWTOOL_LANG", "")
	t.Setenv("LANG", "zh_CN.UTF-8")

	resolver := Resolver{}
	got := resolver.Resolve("", config.RuntimeConfig{})
	if got != "zh-CN" {
		t.Fatalf("expected zh-CN, got %s", got)
	}
}
