package config

// RuntimeConfig stores resolved runtime options. / RuntimeConfig 保存解析后的运行时选项。
type RuntimeConfig struct {
	Language string `json:"language" yaml:"language"`
}
