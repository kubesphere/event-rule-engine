package visitor

import (
	"github.com/kubesphere/event-rule-engine/visitor/parser"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/golang/glog"

	"strconv"
	"strings"
)

type Visitor struct {
	parser.BaseEventRuleVisitor
	valueStack []bool
	m map[string]interface{}
}

func NewVisitor(m map[string]interface{}) *Visitor {
	return &Visitor{
		m: m,
	}
}

func (l *Visitor) pushValue(i bool) {
	l.valueStack = append(l.valueStack, i)
}

func (l *Visitor) popValue() bool {
	if len(l.valueStack) < 1 {
		glog.Errorf("valueStack is empty unable to pop")
		return false
	}

	// Get the last value from the stack.
	result := l.valueStack[len(l.valueStack)-1]

	// Remove the last element from the stack.
	l.valueStack = l.valueStack[:len(l.valueStack)-1]

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
	glog.Info("VisitAndOr")

	//push expression result to stack
	v.visitRule(ctx.Expression(0))
	v.visitRule(ctx.Expression(1))

	//push result to stack
	var t antlr.Token = ctx.GetOp()
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
	glog.Info("VisitNot")

	v.visitRule(ctx.Expression())

	value := v.popValue()
	v.pushValue(!value)

	return nil
}

func (v *Visitor) VisitStringEqualContains(ctx *parser.StringEqualContainsContext) interface{} {
	varName := ctx.VAR().GetText()
	strValue := ctx.STRING().GetText()
	var t antlr.Token = ctx.GetOp()

	strValue = strings.TrimLeft(strValue, `"`)
	strValue = strings.TrimRight(strValue, `"`)

	glog.Info("VisitStringEqualContains %s %d %s", varName, t.GetTokenType(), strValue)

	switch t.GetTokenType() {
	case parser.EventRuleParserEQU:
		if v.m[varName] == nil {
			v.pushValue(false)
		} else {
			v.pushValue(v.m[varName].(string) == strValue)
		}
	case parser.EventRuleParserCONTAINS:
		if v.m[varName] == nil {
			v.pushValue(false)
		} else {
			v.pushValue(strings.Contains(v.m[varName].(string), strValue))
		}
	}

	return nil
}

func (v *Visitor) VisitStringIn(ctx *parser.StringInContext) interface{} {
	varName := ctx.VAR().GetText()
	length := len(ctx.AllSTRING())

	strValues := []string{}
	for i := 0; i<length; i++ {
		strValue := ctx.STRING(i).GetText()
		strValue = strings.TrimLeft(strValue, `"`)
		strValue = strings.TrimRight(strValue, `"`)
		strValues = append(strValues, strValue)
	}

	glog.Info("VisitStringIn %s in %v", varName, strValues)

	if v.m[varName] == nil {
		v.pushValue(false)

		return nil
	}

	varValue := v.m[varName].(string)

	result := false
	for _, strValue := range(strValues) {
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
		glog.Errorf("number err, key: %s, value: %s, err: %s", varName, ctx.NUMBER().GetText(), err.Error())
		return nil
	}
	var t antlr.Token = ctx.GetOp()

	if v.m[varName] == nil {
		v.pushValue(false)
		return nil
	}

	varValue := v.m[varName].(float64)

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
	glog.Info("VisitVariable %v", varName)

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

func EventRuleEvaluate(m map[string]interface{}, expression string) bool {
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
}
