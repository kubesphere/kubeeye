// Code generated from EventRule.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // EventRule

import "github.com/antlr/antlr4/runtime/Go/antlr"

// A complete Visitor for a parse tree produced by EventRuleParser.
type EventRuleVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by EventRuleParser#start.
	VisitStart(ctx *StartContext) interface{}

	// Visit a parse tree produced by EventRuleParser#InOrNot.
	VisitInOrNot(ctx *InOrNotContext) interface{}

	// Visit a parse tree produced by EventRuleParser#RegexOrNot.
	VisitRegexOrNot(ctx *RegexOrNotContext) interface{}

	// Visit a parse tree produced by EventRuleParser#Not.
	VisitNot(ctx *NotContext) interface{}

	// Visit a parse tree produced by EventRuleParser#Parenthesis.
	VisitParenthesis(ctx *ParenthesisContext) interface{}

	// Visit a parse tree produced by EventRuleParser#BoolCompare.
	VisitBoolCompare(ctx *BoolCompareContext) interface{}

	// Visit a parse tree produced by EventRuleParser#Variable.
	VisitVariable(ctx *VariableContext) interface{}

	// Visit a parse tree produced by EventRuleParser#NotVariable.
	VisitNotVariable(ctx *NotVariableContext) interface{}

	// Visit a parse tree produced by EventRuleParser#Compare.
	VisitCompare(ctx *CompareContext) interface{}

	// Visit a parse tree produced by EventRuleParser#ExistsOrNot.
	VisitExistsOrNot(ctx *ExistsOrNotContext) interface{}

	// Visit a parse tree produced by EventRuleParser#AndOr.
	VisitAndOr(ctx *AndOrContext) interface{}

	// Visit a parse tree produced by EventRuleParser#ContainsOrNot.
	VisitContainsOrNot(ctx *ContainsOrNotContext) interface{}
}
