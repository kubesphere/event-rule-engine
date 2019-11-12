package visitor

import (
	"errors"
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/golang/glog"
	"github.com/wanjunlei/event-rule-engine/visitor/parser"

	"strconv"
	"strings"
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
	glog.Infof("VisitAndOr")

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
	glog.Infof("VisitNot")

	v.visitRule(ctx.Expression())

	value := v.popValue()
	v.pushValue(!value)

	return nil
}

func (v *Visitor) VisitStringEqualContains(ctx *parser.StringEqualContainsContext) interface{} {
	t := ctx.GetOp()
	varName := ctx.VAR().GetText()
	strValue := ctx.STRING().GetText()
	strValue = strings.TrimLeft(strValue, `"`)
	strValue = strings.TrimRight(strValue, `"`)

	glog.Infof("VisitStringEqualContains %s %d %s", varName, t.GetTokenType(), strValue)

	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}

	switch t.GetTokenType() {
	case parser.EventRuleParserEQU:
		v.pushValue(v.m[varName].(string) == strValue)
	case parser.EventRuleParserCONTAINS:
		v.pushValue(strings.Contains(v.m[varName].(string), strValue))
	}

	return nil
}

func (v *Visitor) VisitStringIn(ctx *parser.StringInContext) interface{} {
	varName := ctx.VAR().GetText()
	length := len(ctx.AllSTRING())

	var strValues []string
	for i := 0; i < length; i++ {
		strValue := ctx.STRING(i).GetText()
		strValue = strings.TrimLeft(strValue, `"`)
		strValue = strings.TrimRight(strValue, `"`)
		strValues = append(strValues, strValue)
	}

	glog.Infof("VisitStringIn %s in %v", varName, strValues)

	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}

	varValue := v.m[varName].(string)

	result := false
	for _, strValue := range strValues {
		if varValue == strValue {
			result = true
			break
		}
	}

	v.pushValue(result)

	return nil
}

func (v *Visitor) VisitCompareNumber(ctx *parser.CompareNumberContext) interface{} {
	varName := ctx.VAR().GetText()
	numValue, err := strconv.ParseFloat(ctx.NUMBER().GetText(), 64)
	if err != nil {
		panic(fmt.Sprintf("number err, key: %s, value: %s, err: %s", varName, ctx.NUMBER().GetText(), err.Error()))
	}

	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}

	varValue, err := strconv.ParseFloat(v.m[varName].(string), 64)
	if err != nil {
		panic(fmt.Sprintf("number err, key: %s, value: %s, err: %s", varName, ctx.NUMBER().GetText(), err.Error()))
	}

	t := ctx.GetOp()

	glog.Info("VisitCompareNumber %s %d %v", varName, t.GetTokenType(), numValue)

	switch t.GetTokenType() {
	case parser.EventRuleParserEQU:
		v.pushValue(varValue == numValue)
	case parser.EventRuleParserNEQ:
		v.pushValue(varValue != numValue)
	case parser.EventRuleParserGT:
		v.pushValue(varValue > numValue)
	case parser.EventRuleParserLT:
		v.pushValue(varValue < numValue)
	case parser.EventRuleParserGTE:
		v.pushValue(varValue >= numValue)
	case parser.EventRuleParserLTE:
		v.pushValue(varValue <= numValue)
	}

	return nil
}

func (v *Visitor) VisitVariable(ctx *parser.VariableContext) interface{} {
	varName := ctx.VAR().GetText()
	glog.Infof("VisitVariable %v", varName)

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
	err, _ := EventRuleEvaluate(m, "count=1 and reason=pod and count=1")
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
