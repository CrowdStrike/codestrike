package config

import "embed"

// embeddedAssets contains the default configuration distributed in the
// codestrike binary.
//
//go:embed assets/default.yaml assets/prompts/*.md assets/tones/*.md
var embeddedAssets embed.FS
