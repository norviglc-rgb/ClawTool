package schema

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	rootassets "github.com/openclaw/clawtool"
	"github.com/openclaw/clawtool/internal/core"
	"gopkg.in/yaml.v3"
)

// Load returns an embedded schema file. / Load 返回嵌入的 schema 文件。
func Load(name string) ([]byte, error) {
	return fs.ReadFile(rootassets.EmbeddedFiles, "schemas/"+name)
}

// ValidateProfile applies a minimal explicit validation rule-set. / ValidateProfile 应用最小显式校验规则。
func ValidateProfile(data []byte) error {
	var profile core.Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeSchemaValidation,
			MessageKey: "error.schema.profile",
			Cause:      err,
		}
	}

	if profile.Version == "" || profile.Name == "" || profile.Target.Kind == "" {
		return &core.AppError{
			Code:       core.ErrorCodeSchemaValidation,
			MessageKey: "error.schema.profile",
			Cause:      fmt.Errorf("missing required fields"),
		}
	}

	switch profile.Target.Kind {
	case "local":
	case "ssh":
		if strings.TrimSpace(profile.Target.Address) == "" {
			return &core.AppError{
				Code:       core.ErrorCodeSchemaValidation,
				MessageKey: "error.schema.profile",
				Cause:      fmt.Errorf("ssh target requires address"),
			}
		}
	default:
		return &core.AppError{
			Code:       core.ErrorCodeSchemaValidation,
			MessageKey: "error.schema.profile",
			Cause:      fmt.Errorf("unsupported target kind: %s", profile.Target.Kind),
		}
	}

	return nil
}

// ValidateManifest applies a minimal explicit validation rule-set. / ValidateManifest 应用最小显式校验规则。
func ValidateManifest(data []byte) error {
	var manifest core.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return &core.AppError{
			Code:       core.ErrorCodeSchemaValidation,
			MessageKey: "error.schema.manifest",
			Cause:      err,
		}
	}

	if manifest.Version == "" || manifest.Profile == "" {
		return &core.AppError{
			Code:       core.ErrorCodeSchemaValidation,
			MessageKey: "error.schema.manifest",
			Cause:      fmt.Errorf("missing required fields"),
		}
	}

	return nil
}

// CatalogKeys exposes raw locale keys for completeness checks. / CatalogKeys 暴露原始语言键以进行完整性检查。
func CatalogKeys(name string) (map[string]struct{}, error) {
	data, err := fs.ReadFile(rootassets.EmbeddedFiles, "locales/"+name)
	if err != nil {
		return nil, err
	}

	var messages []struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(messages))
	for _, message := range messages {
		keys[message.ID] = struct{}{}
	}

	return keys, nil
}
