package register

import (
	"embed"

	"github.com/leonharetd/kubeeye/pkg/kube"
)

// Provide for user registration
var (
	funcRules        = newFuncRulesBuilder()
	FuncRuleRegistry = funcRules.Registry
	FuncRuleList     = funcRules.List

	// Provide for user registration
	regoRules        = newRegoRulesBuilder()
	RegoRuleRegistry = regoRules.Registry
	RegoRuleList     = regoRules.List
)

type funcRulesBuilder []kube.FuncRule

func newFuncRulesBuilder() *funcRulesBuilder {
	var er funcRulesBuilder
	return &er
}

func (er *funcRulesBuilder) Registry(e kube.FuncRule) error {
	*er = append(*er, e)
	return nil
}

func (er *funcRulesBuilder) List() *funcRulesBuilder {
	return er
}

type regoRulesBuilder []embed.FS

func newRegoRulesBuilder() *regoRulesBuilder {
	var rr regoRulesBuilder
	return &rr
}

func (rr *regoRulesBuilder) Registry(r embed.FS) error {
	*rr = append(*rr, r)
	return nil
}

func (rr *regoRulesBuilder) List() *regoRulesBuilder {
	return rr
}
