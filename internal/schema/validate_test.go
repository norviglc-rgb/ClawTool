package schema

import "testing"

func TestCatalogCompleteness(t *testing.T) {
	t.Parallel()

	base, err := CatalogKeys("en.json")
	if err != nil {
		t.Fatalf("load en: %v", err)
	}

	for _, locale := range []string{"zh-CN.json", "ja.json"} {
		keys, err := CatalogKeys(locale)
		if err != nil {
			t.Fatalf("load %s: %v", locale, err)
		}

		for key := range base {
			if _, ok := keys[key]; !ok {
				t.Fatalf("missing key %s in %s", key, locale)
			}
		}
	}
}

func TestValidateProfile(t *testing.T) {
	t.Parallel()

	profile := []byte("version: v1\nname: test\ntarget:\n  kind: local\n")
	if err := ValidateProfile(profile); err != nil {
		t.Fatalf("validate profile: %v", err)
	}
}

func TestValidateManifest(t *testing.T) {
	t.Parallel()

	manifest := []byte("version: v1\nprofile: default\n")
	if err := ValidateManifest(manifest); err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
}
