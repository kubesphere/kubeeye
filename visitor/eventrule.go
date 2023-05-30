/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package visitor

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/golang/glog"
	"github.com/kubesphere/kubeeye/visitor/parser"
	"regexp"
	"strconv"
	"strings"
)

const (
	LevelInfo = 6
)

const (
	ArrayOperatorAny = 1
	ArrayOperatorAll = 2
)

type Visitor struct {
	parser.BaseEventRuleVisitor
	valueStack []bool
	m          map[string]interface{}
}

func NewVisitor(m map[string]interface{}) *Visitor {
	return &Visitor{
		m: m,
	}
}

func (v *Visitor) pushValue(i bool) {
	v.valueStack = append(v.valueStack, i)
}

func (v *Visitor) popValue() bool {
	if len(v.valueStack) < 1 {
		panic("valueStack is empty unable to pop")
	}

	// Get the last value from the stack.
	result := v.valueStack[len(v.valueStack)-1]

	// Remove the last element from the stack.
	v.valueStack = v.valueStack[:len(v.valueStack)-1]

	return result
}

func (v *Visitor) visitRule(node antlr.RuleNode) interface{} {
	node.Accept(v)
	return nil
}

func (v *Visitor) VisitStart(ctx *parser.StartContext) interface{} {
	return v.visitRule(ctx.Expression())
}

func (v *Visitor) VisitAndOr(ctx *parser.AndOrContext) interface{} {

	//push expression result to stack
	v.visitRule(ctx.Expression(0))
	v.visitRule(ctx.Expression(1))

	//push result to stack
	t := ctx.GetOp()
	right := v.popValue()
	left := v.popValue()
	switch t.GetTokenType() {
	case parser.EventRuleParserAND:
		v.pushValue(left && right)
	case parser.EventRuleParserOR:
		v.pushValue(left || right)
	default:
		panic("should not happen")
	}

	return nil
}

func (v *Visitor) VisitNot(ctx *parser.NotContext) interface{} {

	v.visitRule(ctx.Expression())

	value := v.popValue()
	v.pushValue(!value)

	return nil
}

func (v *Visitor) VisitCompare(ctx *parser.CompareContext) interface{} {

	varName := ctx.VAR().GetText()
	if !strings.Contains(varName, "[") {
		if v.m[varName] == nil {
			v.pushValue(false)
			return nil
		}

		v.pushValue(compare(varName, v.m[varName], ctx))
		return nil
	}

	v.pushValue(arrayOperator(v, varName, ctx.GetOp().GetTokenType(), func(value interface{}) bool {
		return compare(varName, value, ctx)
	}))

	return nil
}

func compare(name string, value interface{}, ctx *parser.CompareContext) bool {

	if value == nil {
		return false
	}

	result := false
	if ctx.STRING() != nil {
		strValue := ctx.STRING().GetText()
		strValue = strings.TrimLeft(strValue, `"`)
		strValue = strings.TrimRight(strValue, `"`)

		switch ctx.GetOp().GetTokenType() {
		case parser.EventRuleParserEQU:
			result = fmt.Sprint(value) == strValue
		case parser.EventRuleParserNEQ:
			result = fmt.Sprint(value) != strValue
		case parser.EventRuleParserGT:
			result = fmt.Sprint(value) > strValue
		case parser.EventRuleParserLT:
			result = fmt.Sprint(value) < strValue
		case parser.EventRuleParserGTE:
			result = fmt.Sprint(value) >= strValue
		case parser.EventRuleParserLTE:
			result = fmt.Sprint(value) <= strValue
		}

		glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", name, value, ctx.GetOp().GetText(), strValue, result)
	} else {
		numValue, err := strconv.ParseFloat(ctx.NUMBER().GetText(), 64)
		if err != nil {
			panic(fmt.Errorf("%s is not number", ctx.NUMBER().GetText()))
		}

		num, err := strconv.ParseFloat(fmt.Sprint(value), 64)
		if err != nil {
			panic(fmt.Errorf("%s is not number", value))
		}

		switch ctx.GetOp().GetTokenType() {
		case parser.EventRuleParserEQU:
			result = num == numValue
		case parser.EventRuleParserNEQ:
			result = num != numValue
		case parser.EventRuleParserGT:
			result = num > numValue
		case parser.EventRuleParserLT:
			result = num < numValue
		case parser.EventRuleParserGTE:
			result = num >= numValue
		case parser.EventRuleParserLTE:
			result = num <= numValue
		}

		glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", name, value, ctx.GetOp().GetText(), numValue, result)
	}

	return result
}

func (v *Visitor) VisitBoolCompare(ctx *parser.BoolCompareContext) interface{} {

	varName := ctx.VAR().GetText()
	if !strings.Contains(varName, "[") {
		v.pushValue(boolCompare(varName, v.m[varName], ctx))
		return nil
	}

	v.pushValue(arrayOperator(v, varName, ctx.GetOp().GetTokenType(), func(value interface{}) bool {
		return boolCompare(varName, value, ctx)
	}))

	return nil
}

func boolCompare(name string, value interface{}, ctx *parser.BoolCompareContext) bool {

	if value == nil {
		return false
	}

	boolValue, err := strconv.ParseBool(ctx.BOOLEAN().GetText())
	if err != nil {
		panic(fmt.Errorf("%s is not bool", ctx.BOOLEAN().GetText()))
	}

	bv, err := strconv.ParseBool(fmt.Sprint(value))
	if err != nil {
		panic(fmt.Errorf("%s is not bool", value))
	}

	result := boolValue == bv
	if ctx.GetOp().GetTokenType() == parser.EventRuleLexerNEQ {
		result = !result
	}

	glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", name, value, ctx.GetOp().GetText(), boolValue, result)
	return result
}

func (v *Visitor) VisitContainsOrNot(ctx *parser.ContainsOrNotContext) interface{} {

	varName := ctx.VAR().GetText()
	if !strings.Contains(varName, "[") {
		v.pushValue(containsOrNot(varName, v.m[varName], ctx))
		return nil
	}

	v.pushValue(arrayOperator(v, varName, ctx.GetOp().GetTokenType(), func(value interface{}) bool {
		return containsOrNot(varName, fmt.Sprint(value), ctx)
	}))

	return nil
}

func containsOrNot(name string, value interface{}, ctx *parser.ContainsOrNotContext) bool {

	if value == nil {
		return false
	}

	node := ctx.STRING()
	var strValue string
	if node != nil {
		strValue = node.GetText()
		strValue = strings.TrimLeft(strValue, `"`)
		strValue = strings.TrimRight(strValue, `"`)
	}
	if node == nil {
		node = ctx.NUMBER()
		strValue = node.GetText()
	}

	result := strings.Contains(fmt.Sprint(value), strValue)
	if ctx.GetOp().GetTokenType() == parser.EventRuleParserNOTCONTAINS {
		result = !result
	}

	glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", name, value, ctx.GetOp().GetText(), strValue, result)
	return result
}

func (v *Visitor) VisitInOrNot(ctx *parser.InOrNotContext) interface{} {

	varName := ctx.VAR().GetText()
	if !strings.Contains(varName, "[") {
		v.pushValue(inOrNot(varName, v.m[varName], ctx))
		return nil
	}

	v.pushValue(arrayOperator(v, varName, ctx.GetOp().GetTokenType(), func(value interface{}) bool {
		return inOrNot(varName, value, ctx)
	}))
	return nil
}

func inOrNot(name string, value interface{}, ctx *parser.InOrNotContext) bool {

	if value == nil {
		return false
	}

	var strValues []string
	for _, p := range ctx.AllNUMBER() {
		strValue := p.GetText()
		strValues = append(strValues, strValue)
	}

	for _, p := range ctx.AllSTRING() {
		strValue := p.GetText()
		strValue = strings.TrimLeft(strValue, `"`)
		strValue = strings.TrimRight(strValue, `"`)
		strValues = append(strValues, strValue)
	}

	result := false
	for _, strValue := range strValues {
		if fmt.Sprint(value) == strValue {
			result = true
			break
		}
	}

	if ctx.GetOp().GetTokenType() == parser.EventRuleParserNOTIN {
		result = !result
	}

	glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", name, value, ctx.GetOp().GetText(), strValues, result)
	return result
}

func (v *Visitor) VisitRegexOrNot(ctx *parser.RegexOrNotContext) interface{} {

	varName := ctx.VAR().GetText()
	if !strings.Contains(varName, "[") {
		v.pushValue(regexOrNot(varName, v.m[varName], ctx))
		return nil
	}

	v.pushValue(arrayOperator(v, varName, ctx.GetOp().GetTokenType(), func(value interface{}) bool {
		return regexOrNot(varName, value, ctx)
	}))
	return nil
}

func regexOrNot(name string, value interface{}, ctx *parser.RegexOrNotContext) bool {

	if value == nil {
		return false
	}

	strValue := ctx.STRING().GetText()
	strValue = strings.TrimLeft(strValue, `"`)
	strValue = strings.TrimRight(strValue, `"`)

	pattern := strValue
	if ctx.GetOp().GetTokenType() == parser.EventRuleLexerLIKE || ctx.GetOp().GetTokenType() == parser.EventRuleLexerNOTLIKE {

		pattern = strings.ReplaceAll(pattern, "?", ".")

		regex, err := regexp.Compile("(\\*)+")
		if err != nil {
			panic(err)
		}
		pattern = regex.ReplaceAllString(pattern, "(.*)")
	}

	result, err := regexp.Match(pattern, []byte(fmt.Sprint(value)))
	if err != nil {
		panic(err)
	}
	if ctx.GetOp().GetTokenType() == parser.EventRuleLexerNOTLIKE || ctx.GetOp().GetTokenType() == parser.EventRuleLexerNOTREGEX {
		result = !result
	}

	glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", name, value, ctx.GetOp().GetText(), strValue, result)
	return result
}

func (v *Visitor) VisitVariable(ctx *parser.VariableContext) interface{} {
	return visitVariable(ctx.VAR().GetText(), v, true)
}

func (v *Visitor) VisitNotVariable(ctx *parser.NotVariableContext) interface{} {
	return visitVariable(ctx.VAR().GetText(), v, false)
}

func visitVariable(varName string, v *Visitor, flag bool) error {
	if !strings.Contains(varName, "[") {
		v.pushValue(variable(varName, v, flag))
		return nil
	}

	v.pushValue(arrayOperator(v, varName, -1, func(value interface{}) bool {
		return variable(varName, v, flag)
	}))

	return nil
}

func variable(varName string, v *Visitor, flag bool) bool {

	if v.m[varName] == nil {
		return false
	}

	bv, err := strconv.ParseBool(fmt.Sprint(v.m[varName]))
	if err != nil {
		panic(fmt.Errorf("%s is not bool", v.m[varName]))
	}
	return bv == flag

}

func (v *Visitor) VisitParenthesis(ctx *parser.ParenthesisContext) interface{} {
	v.visitRule(ctx.Expression())
	return nil
}

func (v *Visitor) VisitExistsOrNot(ctx *parser.ExistsOrNotContext) interface{} {
	varName := ctx.VAR().GetText()
	if !strings.Contains(varName, "[") {
		v.pushValue(existsOrNot(varName, v.m[varName], ctx.GetOp().GetTokenType(), v, true))
		return nil
	}

	v.pushValue(arrayOperator(v, varName, ctx.GetOp().GetTokenType(), func(value interface{}) bool {
		return existsOrNot(varName, value, ctx.GetOp().GetTokenType(), v, false)
	}))
	return nil
}

func existsOrNot(name string, value interface{}, tokenType int, v *Visitor, flag bool) bool {
	result := true
	if value == nil {
		result = false
	}

	if !result && flag {
		for k, v := range v.m {
			if strings.HasPrefix(k, name+".") && v != nil {
				result = true
				break
			}
		}
	}

	if tokenType == parser.EventRuleParserNOTEXISTS {
		result = !result
	}

	glog.V(LevelInfo).Infof("visit %s %s, %s", name, tokenType, result)
	return result
}

func CheckRule(expression string) (bool, error) {

	m := make(map[string]interface{})
	err, _ := EventRuleEvaluate(m, expression)
	if err != nil {
		return false, err
	}

	return true, nil
}

func EventRuleEvaluate(m map[string]interface{}, expression string) (error, bool) {

	var err error
	res := func() bool {
		defer func() {
			if i := recover(); i != nil {
				err = errors.New(i.(string))
			}
		}()

		is := antlr.NewInputStream(expression)
		// Create the Lexer
		lexer := parser.NewEventRuleLexer(is)
		tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		// Create the Parser
		p := parser.NewEventRuleParser(tokens)
		v := NewVisitor(m)
		//Start is rule name of EventRule.g4
		p.Start().Accept(v)

		return v.popValue()
	}()

	if err != nil {
		return err, false
	}

	return nil, res
}

func arrayOperator(v *Visitor, varName string, tokenType int, match func(value interface{}) bool) bool {
	if strings.HasSuffix(varName, "]") &&
		tokenType != parser.EventRuleParserNOTCONTAINS &&
		tokenType != parser.EventRuleParserCONTAINS {
		panic("array only support contains or not contains method")
	}

	ss := strings.Split(varName, ".")
	buf := bytes.Buffer{}
	for i := 0; i < len(ss); i++ {
		s := ss[i]
		if !strings.Contains(s, "[") {
			buf.WriteString(s)
			buf.WriteString(".")
			continue
		}

		buf.WriteString(s[0:strings.Index(s, "[")])
		ss = ss[i:]
		break
	}

	if v.m[buf.String()] == nil {
		return false
	}

	varValue := v.m[buf.String()]
	if !isArray(varValue) {
		panic(fmt.Errorf("%s is not array", buf.String()))
	}

	value := varValue.([]interface{})
	if len(value) == 0 {
		return false
	}

	childValue, op, err := getChildValue(ss[0], value)
	if err != nil {
		panic(err)
	}

	if childValue == nil || len(childValue) == 0 {
		return false
	}

	return arrayMatch(op, childValue, ss[1:], match)
}

func arrayMatch(op int, value []interface{}, array []string, match func(value interface{}) bool) bool {

	b := false
	result := false
	for _, v := range value {
		if len(array) == 0 {
			b = match(v)
		} else {
			if !isMap(v) {
				panic("the value not a map")
			}

			s := array[0]
			key := s
			if strings.Contains(s, "[") {
				key = key[0:strings.Index(key, "[")]
			}

			subValue, ok := v.(map[string]interface{})[key]
			if !ok && op == ArrayOperatorAll {
				return false
			}

			cop := 0
			var childValue []interface{}
			if strings.Contains(s, "[") {
				if !isArray(subValue) {
					panic("the value not a array")
				}

				var err error
				childValue, cop, err = getChildValue(s, subValue.([]interface{}))
				if err != nil {
					panic(err)
				}

				if childValue == nil || len(childValue) == 0 {
					return false
				}
			} else {
				childValue = append(childValue, subValue)
				cop = ArrayOperatorAny
			}

			b = arrayMatch(cop, childValue, array[1:], match)
		}

		switch op {
		case ArrayOperatorAny:
			if b {
				return true
			}
		case ArrayOperatorAll:
			if !b {
				return false
			}
			result = true
		}
	}

	return result
}

func getChildValue(param string, value []interface{}) ([]interface{}, int, error) {

	s := param[strings.Index(param, "["):]
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")

	switch s {
	case "*":
		return value, ArrayOperatorAny, nil
	default:
		if strings.Contains(s, ":") {
			ns := strings.Split(s, ":")
			start := 0
			if len(ns[0]) > 0 {
				var err error
				start, err = strconv.Atoi(ns[0])
				if err != nil {
					return nil, 0, err
				}

				if start < 0 {
					start = 0
				}
				if start > len(value) {
					start = len(value)
				}
			}
			end := len(value)
			if len(ns[1]) > 0 {
				var err error
				end, err = strconv.Atoi(ns[1])
				if err != nil {
					return nil, 0, err
				}

				if end < 0 {
					return nil, 0, fmt.Errorf("array out of bound end %d", end)
				}

				if end > len(value) {
					end = len(value)
				}
			}
			if start > end {
				return nil, 0, fmt.Errorf("wrong array range start %d, end %d", start, end)
			}

			return value[start:end], ArrayOperatorAll, nil
		} else {
			index, err := strconv.Atoi(s)
			if err != nil {
				return nil, 0, err
			}

			var cv []interface{}
			if index >= 0 && index < len(value) {
				cv = append(cv, value[index])
			}

			return cv, ArrayOperatorAll, nil
		}
	}
}

func isArray(v interface{}) bool {
	switch v.(type) {
	case []interface{}:
		return true
	default:
		return false
	}
}

func isMap(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}:
		return true
	default:
		return false
	}
}
