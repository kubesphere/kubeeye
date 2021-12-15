package regorules

import (
	"embed"
)

//go:embed rules
var DefaultEmbRegoRules embed.FS
