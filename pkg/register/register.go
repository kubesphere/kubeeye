package register

import (
	"embed"

	"github.com/leonharetd/kubeeye/pkg/kube"
)

// Provide for user registration
var (
	execRules        = newExecRulesBuilder()
	ExecRuleRegistry = execRules.Registry
	ExecRuleList     = execRules.List

	// Provide for user registration
	regoRules        = newRegoRulesBuilder()
	RegoRuleRegistry = regoRules.Registry
	RegoRuleList     = regoRules.List
)

type EXECRule interface {
	Exec() []kube.ResultReceiver
}

type execRulesBuilder []EXECRule

func newExecRulesBuilder() *execRulesBuilder {
	var er execRulesBuilder
	return &er
}

func (er *execRulesBuilder) Registry(e EXECRule) error{
	*er = append(*er, e)
	return nil
}

func (er *execRulesBuilder) List() *execRulesBuilder {
	return er
}

type regoRulesBuilder []embed.FS

func newRegoRulesBuilder() *regoRulesBuilder {
	var rr regoRulesBuilder
	return &rr
}

func (rr *regoRulesBuilder) Registry(r embed.FS) error{
	*rr = append(*rr, r)
	return nil
}

func (rr *regoRulesBuilder) List() *regoRulesBuilder {
	return rr
}
