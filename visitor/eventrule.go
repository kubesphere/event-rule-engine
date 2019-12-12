package visitor

import (
	"errors"
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/golang/glog"
	"github.com/wanjunlei/event-rule-engine/visitor/parser"
	"regexp"
	"strings"
)

const (
	LevelInfo = 6
	//LevelWarning = 5
	//LevelError = 4
	//LevelFatal = 3
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
	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}
	varValue := fmt.Sprint(v.m[varName])

	node := ctx.STRING()
	var strValue string
	if node != nil {
		strValue = node.GetText()
		strValue = strings.TrimLeft(strValue, `"`)
		strValue = strings.TrimRight(strValue, `"`)
	} else {
		node = ctx.NUMBER()
		strValue = node.GetText()
	}

	result := false
	switch ctx.GetOp().GetTokenType() {
	case parser.EventRuleParserEQU:
		result = varValue == strValue
	case parser.EventRuleParserNEQ:
		result = varValue != strValue
	case parser.EventRuleParserGT:
		result = varValue > strValue
	case parser.EventRuleParserLT:
		result = varValue < strValue
	case parser.EventRuleParserGTE:
		result = varValue >= strValue
	case parser.EventRuleParserLTE:
		result = varValue <= strValue
	}

	v.pushValue(result)
	glog.V(LevelInfo).Info("visit %s(%s) %s %s, %s", varName, varValue, ctx.GetOp().GetText(), strValue, result)

	return nil
}

func (v *Visitor) VisitContainsOrNot(ctx *parser.ContainsOrNotContext) interface{} {

	varName := ctx.VAR().GetText()
	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}
	varValue := fmt.Sprint(v.m[varName])

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

	result := strings.Contains(varValue, strValue)
	if ctx.GetOp().GetTokenType() == parser.EventRuleParserNOTCONTAINS {
		result = !result
	}
	v.pushValue(result)
	glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", varName, varValue, ctx.GetOp().GetText(), strValue, result)

	return nil
}

func (v *Visitor) VisitInOrNot(ctx *parser.InOrNotContext) interface{} {

	varName := ctx.VAR().GetText()
	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}
	varValue := fmt.Sprint(v.m[varName])

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
		if varValue == strValue {
			result = true
			break
		}
	}

	if ctx.GetOp().GetTokenType() == parser.EventRuleParserNOTIN {
		result = !result
	}

	v.pushValue(result)
	glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", varName, varValue, ctx.GetOp().GetText(), strValues, result)

	return nil
}

func (v *Visitor) VisitRegexpOrNot(ctx *parser.RegexpOrNotContext) interface{} {

	varName := ctx.VAR().GetText()
	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}
	varValue := fmt.Sprint(v.m[varName])

	strValue := ctx.STRING().GetText()
	strValue = strings.TrimLeft(strValue, `"`)
	strValue = strings.TrimRight(strValue, `"`)

	pattern := strValue
	if ctx.GetOp().GetTokenType() == parser.EventRuleLexerLIKE ||  ctx.GetOp().GetTokenType() == parser.EventRuleLexerNOTLIKE{

		pattern = strings.ReplaceAll(pattern, "?", ".")

		rege, err := regexp.Compile("(\\*)+")
		if err != nil {
			panic(err)
		}
		pattern = rege.ReplaceAllString(pattern, "(.*)")
	}

	result, err := regexp.Match(pattern, []byte(varValue))
	if err!= nil {
		panic(err)
	}
	if ctx.GetOp().GetTokenType() == parser.EventRuleLexerNOTLIKE || ctx.GetOp().GetTokenType() == parser.EventRuleLexerNOTREGEXP {
		result = !result
	}
	v.pushValue(result)
	glog.V(LevelInfo).Infof("visit %s(%s) %s %s, %s", varName, varValue, ctx.GetOp().GetText(), strValue, result)

	return nil
}

func (v *Visitor) VisitVariable(ctx *parser.VariableContext) interface{} {
	varName := ctx.VAR().GetText()

	if v.m[varName] == nil {
		v.pushValue(false)
	} else {
		v.pushValue(v.m[varName].(bool))
	}

	return nil
}

func (v *Visitor) VisitParenthesis(ctx *parser.ParenthesisContext) interface{} {
	v.visitRule(ctx.Expression())
	return nil
}

func CheckRule(expression string) bool {

	m := make(map[string]interface{})
	err, _ := EventRuleEvaluate(m, expression)
	if err != nil {
		return false
	}

	return true
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
