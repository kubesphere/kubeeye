// Code generated from EventRule.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // EventRule

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 29, 56, 4,
	2, 9, 2, 4, 3, 9, 3, 3, 2, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 7, 3, 32, 10, 3, 12, 3, 14, 3, 35, 11, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 5, 3, 46, 10, 3, 3,
	3, 3, 3, 3, 3, 7, 3, 51, 10, 3, 12, 3, 14, 3, 54, 11, 3, 3, 3, 2, 3, 4,
	4, 2, 4, 2, 10, 3, 2, 8, 13, 4, 2, 25, 25, 27, 27, 3, 2, 8, 9, 3, 2, 14,
	15, 3, 2, 16, 17, 3, 2, 18, 21, 3, 2, 22, 23, 3, 2, 5, 6, 2, 64, 2, 6,
	3, 2, 2, 2, 4, 45, 3, 2, 2, 2, 6, 7, 5, 4, 3, 2, 7, 8, 7, 2, 2, 3, 8, 3,
	3, 2, 2, 2, 9, 10, 8, 3, 1, 2, 10, 11, 7, 7, 2, 2, 11, 46, 5, 4, 3, 12,
	12, 13, 7, 3, 2, 2, 13, 14, 5, 4, 3, 2, 14, 15, 7, 4, 2, 2, 15, 46, 3,
	2, 2, 2, 16, 17, 7, 28, 2, 2, 17, 18, 9, 2, 2, 2, 18, 46, 9, 3, 2, 2, 19,
	20, 7, 28, 2, 2, 20, 21, 9, 4, 2, 2, 21, 46, 7, 26, 2, 2, 22, 23, 7, 28,
	2, 2, 23, 24, 9, 5, 2, 2, 24, 46, 9, 3, 2, 2, 25, 26, 7, 28, 2, 2, 26,
	27, 9, 6, 2, 2, 27, 28, 7, 3, 2, 2, 28, 33, 9, 3, 2, 2, 29, 30, 7, 24,
	2, 2, 30, 32, 9, 3, 2, 2, 31, 29, 3, 2, 2, 2, 32, 35, 3, 2, 2, 2, 33, 31,
	3, 2, 2, 2, 33, 34, 3, 2, 2, 2, 34, 36, 3, 2, 2, 2, 35, 33, 3, 2, 2, 2,
	36, 46, 7, 4, 2, 2, 37, 38, 7, 28, 2, 2, 38, 39, 9, 7, 2, 2, 39, 46, 7,
	27, 2, 2, 40, 41, 7, 28, 2, 2, 41, 46, 9, 8, 2, 2, 42, 46, 7, 28, 2, 2,
	43, 44, 7, 7, 2, 2, 44, 46, 7, 28, 2, 2, 45, 9, 3, 2, 2, 2, 45, 12, 3,
	2, 2, 2, 45, 16, 3, 2, 2, 2, 45, 19, 3, 2, 2, 2, 45, 22, 3, 2, 2, 2, 45,
	25, 3, 2, 2, 2, 45, 37, 3, 2, 2, 2, 45, 40, 3, 2, 2, 2, 45, 42, 3, 2, 2,
	2, 45, 43, 3, 2, 2, 2, 46, 52, 3, 2, 2, 2, 47, 48, 12, 13, 2, 2, 48, 49,
	9, 9, 2, 2, 49, 51, 5, 4, 3, 14, 50, 47, 3, 2, 2, 2, 51, 54, 3, 2, 2, 2,
	52, 50, 3, 2, 2, 2, 52, 53, 3, 2, 2, 2, 53, 5, 3, 2, 2, 2, 54, 52, 3, 2,
	2, 2, 5, 33, 45, 52,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'('", "')'", "'and'", "'or'", "", "'='", "'!='", "'>'", "'<'", "'>='",
	"'<='", "'contains'", "'not contains'", "'in'", "'not in'", "'like'", "'not like'",
	"'regex'", "'not regex'", "'exists'", "'not exists'", "','",
}
var symbolicNames = []string{
	"", "", "", "AND", "OR", "NOT", "EQU", "NEQ", "GT", "LT", "GTE", "LTE",
	"CONTAINS", "NOTCONTAINS", "IN", "NOTIN", "LIKE", "NOTLIKE", "REGEX", "NOTREGEX",
	"EXISTS", "NOTEXISTS", "COMMA", "NUMBER", "BOOLEAN", "STRING", "VAR", "WHITESPACE",
}

var ruleNames = []string{
	"start", "expression",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type EventRuleParser struct {
	*antlr.BaseParser
}

func NewEventRuleParser(input antlr.TokenStream) *EventRuleParser {
	this := new(EventRuleParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "EventRule.g4"

	return this
}

// EventRuleParser tokens.
const (
	EventRuleParserEOF         = antlr.TokenEOF
	EventRuleParserT__0        = 1
	EventRuleParserT__1        = 2
	EventRuleParserAND         = 3
	EventRuleParserOR          = 4
	EventRuleParserNOT         = 5
	EventRuleParserEQU         = 6
	EventRuleParserNEQ         = 7
	EventRuleParserGT          = 8
	EventRuleParserLT          = 9
	EventRuleParserGTE         = 10
	EventRuleParserLTE         = 11
	EventRuleParserCONTAINS    = 12
	EventRuleParserNOTCONTAINS = 13
	EventRuleParserIN          = 14
	EventRuleParserNOTIN       = 15
	EventRuleParserLIKE        = 16
	EventRuleParserNOTLIKE     = 17
	EventRuleParserREGEX       = 18
	EventRuleParserNOTREGEX    = 19
	EventRuleParserEXISTS      = 20
	EventRuleParserNOTEXISTS   = 21
	EventRuleParserCOMMA       = 22
	EventRuleParserNUMBER      = 23
	EventRuleParserBOOLEAN     = 24
	EventRuleParserSTRING      = 25
	EventRuleParserVAR         = 26
	EventRuleParserWHITESPACE  = 27
)

// EventRuleParser rules.
const (
	EventRuleParserRULE_start      = 0
	EventRuleParserRULE_expression = 1
)

// IStartContext is an interface to support dynamic dispatch.
type IStartContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsStartContext differentiates from other interfaces.
	IsStartContext()
}

type StartContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStartContext() *StartContext {
	var p = new(StartContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = EventRuleParserRULE_start
	return p
}

func (*StartContext) IsStartContext() {}

func NewStartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StartContext {
	var p = new(StartContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = EventRuleParserRULE_start

	return p
}

func (s *StartContext) GetParser() antlr.Parser { return s.parser }

func (s *StartContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *StartContext) EOF() antlr.TerminalNode {
	return s.GetToken(EventRuleParserEOF, 0)
}

func (s *StartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StartContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StartContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitStart(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *EventRuleParser) Start() (localctx IStartContext) {
	localctx = NewStartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, EventRuleParserRULE_start)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(4)
		p.expression(0)
	}
	{
		p.SetState(5)
		p.Match(EventRuleParserEOF)
	}

	return localctx
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = EventRuleParserRULE_expression
	return p
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = EventRuleParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) CopyFrom(ctx *ExpressionContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type InOrNotContext struct {
	*ExpressionContext
	op antlr.Token
}

func NewInOrNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InOrNotContext {
	var p = new(InOrNotContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *InOrNotContext) GetOp() antlr.Token { return s.op }

func (s *InOrNotContext) SetOp(v antlr.Token) { s.op = v }

func (s *InOrNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InOrNotContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *InOrNotContext) AllNUMBER() []antlr.TerminalNode {
	return s.GetTokens(EventRuleParserNUMBER)
}

func (s *InOrNotContext) NUMBER(i int) antlr.TerminalNode {
	return s.GetToken(EventRuleParserNUMBER, i)
}

func (s *InOrNotContext) AllSTRING() []antlr.TerminalNode {
	return s.GetTokens(EventRuleParserSTRING)
}

func (s *InOrNotContext) STRING(i int) antlr.TerminalNode {
	return s.GetToken(EventRuleParserSTRING, i)
}

func (s *InOrNotContext) IN() antlr.TerminalNode {
	return s.GetToken(EventRuleParserIN, 0)
}

func (s *InOrNotContext) NOTIN() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNOTIN, 0)
}

func (s *InOrNotContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(EventRuleParserCOMMA)
}

func (s *InOrNotContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(EventRuleParserCOMMA, i)
}

func (s *InOrNotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitInOrNot(s)

	default:
		return t.VisitChildren(s)
	}
}

type RegexOrNotContext struct {
	*ExpressionContext
	op antlr.Token
}

func NewRegexOrNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *RegexOrNotContext {
	var p = new(RegexOrNotContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *RegexOrNotContext) GetOp() antlr.Token { return s.op }

func (s *RegexOrNotContext) SetOp(v antlr.Token) { s.op = v }

func (s *RegexOrNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RegexOrNotContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *RegexOrNotContext) STRING() antlr.TerminalNode {
	return s.GetToken(EventRuleParserSTRING, 0)
}

func (s *RegexOrNotContext) REGEX() antlr.TerminalNode {
	return s.GetToken(EventRuleParserREGEX, 0)
}

func (s *RegexOrNotContext) NOTREGEX() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNOTREGEX, 0)
}

func (s *RegexOrNotContext) LIKE() antlr.TerminalNode {
	return s.GetToken(EventRuleParserLIKE, 0)
}

func (s *RegexOrNotContext) NOTLIKE() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNOTLIKE, 0)
}

func (s *RegexOrNotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitRegexOrNot(s)

	default:
		return t.VisitChildren(s)
	}
}

type NotContext struct {
	*ExpressionContext
}

func NewNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NotContext {
	var p = new(NotContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *NotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NotContext) NOT() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNOT, 0)
}

func (s *NotContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *NotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitNot(s)

	default:
		return t.VisitChildren(s)
	}
}

type ParenthesisContext struct {
	*ExpressionContext
}

func NewParenthesisContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParenthesisContext {
	var p = new(ParenthesisContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ParenthesisContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParenthesisContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ParenthesisContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitParenthesis(s)

	default:
		return t.VisitChildren(s)
	}
}

type BoolCompareContext struct {
	*ExpressionContext
	op antlr.Token
}

func NewBoolCompareContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BoolCompareContext {
	var p = new(BoolCompareContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *BoolCompareContext) GetOp() antlr.Token { return s.op }

func (s *BoolCompareContext) SetOp(v antlr.Token) { s.op = v }

func (s *BoolCompareContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BoolCompareContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *BoolCompareContext) BOOLEAN() antlr.TerminalNode {
	return s.GetToken(EventRuleParserBOOLEAN, 0)
}

func (s *BoolCompareContext) EQU() antlr.TerminalNode {
	return s.GetToken(EventRuleParserEQU, 0)
}

func (s *BoolCompareContext) NEQ() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNEQ, 0)
}

func (s *BoolCompareContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitBoolCompare(s)

	default:
		return t.VisitChildren(s)
	}
}

type VariableContext struct {
	*ExpressionContext
}

func NewVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VariableContext {
	var p = new(VariableContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *VariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *VariableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

type NotVariableContext struct {
	*ExpressionContext
}

func NewNotVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NotVariableContext {
	var p = new(NotVariableContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *NotVariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NotVariableContext) NOT() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNOT, 0)
}

func (s *NotVariableContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *NotVariableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitNotVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

type CompareContext struct {
	*ExpressionContext
	op antlr.Token
}

func NewCompareContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CompareContext {
	var p = new(CompareContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *CompareContext) GetOp() antlr.Token { return s.op }

func (s *CompareContext) SetOp(v antlr.Token) { s.op = v }

func (s *CompareContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CompareContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *CompareContext) STRING() antlr.TerminalNode {
	return s.GetToken(EventRuleParserSTRING, 0)
}

func (s *CompareContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNUMBER, 0)
}

func (s *CompareContext) EQU() antlr.TerminalNode {
	return s.GetToken(EventRuleParserEQU, 0)
}

func (s *CompareContext) NEQ() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNEQ, 0)
}

func (s *CompareContext) GT() antlr.TerminalNode {
	return s.GetToken(EventRuleParserGT, 0)
}

func (s *CompareContext) LT() antlr.TerminalNode {
	return s.GetToken(EventRuleParserLT, 0)
}

func (s *CompareContext) GTE() antlr.TerminalNode {
	return s.GetToken(EventRuleParserGTE, 0)
}

func (s *CompareContext) LTE() antlr.TerminalNode {
	return s.GetToken(EventRuleParserLTE, 0)
}

func (s *CompareContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitCompare(s)

	default:
		return t.VisitChildren(s)
	}
}

type ExistsOrNotContext struct {
	*ExpressionContext
	op antlr.Token
}

func NewExistsOrNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ExistsOrNotContext {
	var p = new(ExistsOrNotContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ExistsOrNotContext) GetOp() antlr.Token { return s.op }

func (s *ExistsOrNotContext) SetOp(v antlr.Token) { s.op = v }

func (s *ExistsOrNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExistsOrNotContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *ExistsOrNotContext) EXISTS() antlr.TerminalNode {
	return s.GetToken(EventRuleParserEXISTS, 0)
}

func (s *ExistsOrNotContext) NOTEXISTS() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNOTEXISTS, 0)
}

func (s *ExistsOrNotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitExistsOrNot(s)

	default:
		return t.VisitChildren(s)
	}
}

type AndOrContext struct {
	*ExpressionContext
	op antlr.Token
}

func NewAndOrContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AndOrContext {
	var p = new(AndOrContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *AndOrContext) GetOp() antlr.Token { return s.op }

func (s *AndOrContext) SetOp(v antlr.Token) { s.op = v }

func (s *AndOrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AndOrContext) AllExpression() []IExpressionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpressionContext)(nil)).Elem())
	var tst = make([]IExpressionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpressionContext)
		}
	}

	return tst
}

func (s *AndOrContext) Expression(i int) IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *AndOrContext) AND() antlr.TerminalNode {
	return s.GetToken(EventRuleParserAND, 0)
}

func (s *AndOrContext) OR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserOR, 0)
}

func (s *AndOrContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitAndOr(s)

	default:
		return t.VisitChildren(s)
	}
}

type ContainsOrNotContext struct {
	*ExpressionContext
	op antlr.Token
}

func NewContainsOrNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ContainsOrNotContext {
	var p = new(ContainsOrNotContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ContainsOrNotContext) GetOp() antlr.Token { return s.op }

func (s *ContainsOrNotContext) SetOp(v antlr.Token) { s.op = v }

func (s *ContainsOrNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ContainsOrNotContext) VAR() antlr.TerminalNode {
	return s.GetToken(EventRuleParserVAR, 0)
}

func (s *ContainsOrNotContext) STRING() antlr.TerminalNode {
	return s.GetToken(EventRuleParserSTRING, 0)
}

func (s *ContainsOrNotContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNUMBER, 0)
}

func (s *ContainsOrNotContext) CONTAINS() antlr.TerminalNode {
	return s.GetToken(EventRuleParserCONTAINS, 0)
}

func (s *ContainsOrNotContext) NOTCONTAINS() antlr.TerminalNode {
	return s.GetToken(EventRuleParserNOTCONTAINS, 0)
}

func (s *ContainsOrNotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case EventRuleVisitor:
		return t.VisitContainsOrNot(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *EventRuleParser) Expression() (localctx IExpressionContext) {
	return p.expression(0)
}

func (p *EventRuleParser) expression(_p int) (localctx IExpressionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()
	_parentState := p.GetState()
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpressionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 2
	p.EnterRecursionRule(localctx, 2, EventRuleParserRULE_expression, _p)
	var _la int

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(43)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext()) {
	case 1:
		localctx = NewNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(8)
			p.Match(EventRuleParserNOT)
		}
		{
			p.SetState(9)
			p.expression(10)
		}

	case 2:
		localctx = NewParenthesisContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(10)
			p.Match(EventRuleParserT__0)
		}
		{
			p.SetState(11)
			p.expression(0)
		}
		{
			p.SetState(12)
			p.Match(EventRuleParserT__1)
		}

	case 3:
		localctx = NewCompareContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(14)
			p.Match(EventRuleParserVAR)
		}
		{
			p.SetState(15)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*CompareContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<EventRuleParserEQU)|(1<<EventRuleParserNEQ)|(1<<EventRuleParserGT)|(1<<EventRuleParserLT)|(1<<EventRuleParserGTE)|(1<<EventRuleParserLTE))) != 0) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*CompareContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(16)
			_la = p.GetTokenStream().LA(1)

			if !(_la == EventRuleParserNUMBER || _la == EventRuleParserSTRING) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	case 4:
		localctx = NewBoolCompareContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(17)
			p.Match(EventRuleParserVAR)
		}
		{
			p.SetState(18)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*BoolCompareContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == EventRuleParserEQU || _la == EventRuleParserNEQ) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*BoolCompareContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(19)
			p.Match(EventRuleParserBOOLEAN)
		}

	case 5:
		localctx = NewContainsOrNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(20)
			p.Match(EventRuleParserVAR)
		}
		{
			p.SetState(21)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*ContainsOrNotContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == EventRuleParserCONTAINS || _la == EventRuleParserNOTCONTAINS) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*ContainsOrNotContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(22)
			_la = p.GetTokenStream().LA(1)

			if !(_la == EventRuleParserNUMBER || _la == EventRuleParserSTRING) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	case 6:
		localctx = NewInOrNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(23)
			p.Match(EventRuleParserVAR)
		}
		{
			p.SetState(24)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*InOrNotContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == EventRuleParserIN || _la == EventRuleParserNOTIN) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*InOrNotContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(25)
			p.Match(EventRuleParserT__0)
		}
		{
			p.SetState(26)
			_la = p.GetTokenStream().LA(1)

			if !(_la == EventRuleParserNUMBER || _la == EventRuleParserSTRING) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		p.SetState(31)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for _la == EventRuleParserCOMMA {
			{
				p.SetState(27)
				p.Match(EventRuleParserCOMMA)
			}
			{
				p.SetState(28)
				_la = p.GetTokenStream().LA(1)

				if !(_la == EventRuleParserNUMBER || _la == EventRuleParserSTRING) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}

			p.SetState(33)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(34)
			p.Match(EventRuleParserT__1)
		}

	case 7:
		localctx = NewRegexOrNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(35)
			p.Match(EventRuleParserVAR)
		}
		{
			p.SetState(36)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*RegexOrNotContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<EventRuleParserLIKE)|(1<<EventRuleParserNOTLIKE)|(1<<EventRuleParserREGEX)|(1<<EventRuleParserNOTREGEX))) != 0) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*RegexOrNotContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(37)
			p.Match(EventRuleParserSTRING)
		}

	case 8:
		localctx = NewExistsOrNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(38)
			p.Match(EventRuleParserVAR)
		}
		{
			p.SetState(39)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*ExistsOrNotContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == EventRuleParserEXISTS || _la == EventRuleParserNOTEXISTS) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*ExistsOrNotContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	case 9:
		localctx = NewVariableContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(40)
			p.Match(EventRuleParserVAR)
		}

	case 10:
		localctx = NewNotVariableContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(41)
			p.Match(EventRuleParserNOT)
		}
		{
			p.SetState(42)
			p.Match(EventRuleParserVAR)
		}

	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			localctx = NewAndOrContext(p, NewExpressionContext(p, _parentctx, _parentState))
			p.PushNewRecursionContext(localctx, _startState, EventRuleParserRULE_expression)
			p.SetState(45)

			if !(p.Precpred(p.GetParserRuleContext(), 11)) {
				panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 11)", ""))
			}
			{
				p.SetState(46)

				var _lt = p.GetTokenStream().LT(1)

				localctx.(*AndOrContext).op = _lt

				_la = p.GetTokenStream().LA(1)

				if !(_la == EventRuleParserAND || _la == EventRuleParserOR) {
					var _ri = p.GetErrorHandler().RecoverInline(p)

					localctx.(*AndOrContext).op = _ri
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}
			{
				p.SetState(47)
				p.expression(12)
			}

		}
		p.SetState(52)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())
	}

	return localctx
}

func (p *EventRuleParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 1:
		var t *ExpressionContext = nil
		if localctx != nil {
			t = localctx.(*ExpressionContext)
		}
		return p.Expression_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *EventRuleParser) Expression_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 11)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
