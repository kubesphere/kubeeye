package register

import (
	"embed"

	"github.com/kubesphere/kubeeye/pkg/funcrules"
	"github.com/kubesphere/kubeeye/pkg/regorules"
)

var RegoRulesListChan = make(chan string)
var FuncRulesListchan = make(chan funcrules.FuncRule)

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

func init() {
	FuncRuleRegistry(funcrules.CertExpireRule{})
	RegoRuleRegistry(regorules.DefaultEmbRegoRules)
}

type funcRulesBuilder []funcrules.FuncRule

func newFuncRulesBuilder() *funcRulesBuilder {
	var er funcRulesBuilder
	return &er
}

func (er *funcRulesBuilder) Registry(e funcrules.FuncRule) error {
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
