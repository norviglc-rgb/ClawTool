package clawtoolassets

import "embed"

// EmbeddedFiles embeds bundled runtime assets. / EmbeddedFiles 嵌入运行时资源。
//
//go:embed locales/*.json schemas/*.json templates/*.yaml
var EmbeddedFiles embed.FS
