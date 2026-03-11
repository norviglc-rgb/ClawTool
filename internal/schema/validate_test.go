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

func TestValidateProfileSSH(t *testing.T) {
	t.Parallel()

	profile := []byte("version: v1\nname: remote\ntarget:\n  kind: ssh\n  address: ssh.example.internal\n  user: deploy\n  port: 22\n  host_key_strategy: known_hosts\n")
	if err := ValidateProfile(profile); err != nil {
		t.Fatalf("validate ssh profile: %v", err)
	}
}

func TestValidateProfileRejectsBadHostKeyStrategy(t *testing.T) {
	t.Parallel()

	profile := []byte("version: v1\nname: remote\ntarget:\n  kind: ssh\n  address: ssh.example.internal\n  host_key_strategy: broken\n")
	if err := ValidateProfile(profile); err == nil {
		t.Fatal("expected bad host key strategy to be rejected")
	}
}

func TestValidateManifest(t *testing.T) {
	t.Parallel()

	manifest := []byte("version: v1\nprofile: default\n")
	if err := ValidateManifest(manifest); err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
}
