package regorules

import (
	"embed"

	"github.com/leonharetd/kubeeye/pkg/register"
)

//go:embed rules
var DefaultEmbRegoRules embed.FS

func init() {
	register.RegoRuleRegistry(DefaultEmbRegoRules)
}
