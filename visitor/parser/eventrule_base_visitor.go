// Code generated from EventRule.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // EventRule

import "github.com/antlr/antlr4/runtime/Go/antlr"

type BaseEventRuleVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseEventRuleVisitor) VisitStart(ctx *StartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitInOrNot(ctx *InOrNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitRegexOrNot(ctx *RegexOrNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitNot(ctx *NotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitParenthesis(ctx *ParenthesisContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitBoolCompare(ctx *BoolCompareContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitVariable(ctx *VariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitNotVariable(ctx *NotVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitCompare(ctx *CompareContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitExistsOrNot(ctx *ExistsOrNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitAndOr(ctx *AndOrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseEventRuleVisitor) VisitContainsOrNot(ctx *ContainsOrNotContext) interface{} {
	return v.VisitChildren(ctx)
}
