package regorules

import (
	"embed"

	"github.com/kubesphere/kubeeye/pkg/register"
)

//go:embed rules
var DefaultEmbRegoRules embed.FS

func init() {
	register.RegoRuleRegistry(DefaultEmbRegoRules)
}
