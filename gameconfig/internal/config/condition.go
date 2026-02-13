package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Condition 条件表达式接口
type Condition interface {
	// Evaluate 评估条件是否满足
	Evaluate(ctx *EvalContext) (bool, error)
	// String 返回条件表达式的字符串表示
	String() string
	// DependentFields 返回依赖的字段名
	DependentFields() []string
}

// EvalContext 评估上下文
type EvalContext struct {
	Values        map[string]interface{} // 当前行的字段值
	ResolvedFields map[string]bool        // 已解析的字段
	Errors        []error               // 错误收集
}

// NewEvalContext 创建评估上下文
func NewEvalContext() *EvalContext {
	return &EvalContext{
		Values:        make(map[string]interface{}),
		ResolvedFields: make(map[string]bool),
		Errors:        make([]error, 0),
	}
}

// SetValue 设置字段值
func (c *EvalContext) SetValue(name string, value interface{}) {
	c.Values[name] = value
}

// GetValue 获取字段值
func (c *EvalContext) GetValue(name string) (interface{}, bool) {
	val, ok := c.Values[name]
	return val, ok
}

// MarkResolved 标记字段已解析
func (c *EvalContext) MarkResolved(name string) {
	c.ResolvedFields[name] = true
}

// IsResolved 检查字段是否已解析
func (c *EvalContext) IsResolved(name string) bool {
	return c.ResolvedFields[name]
}

// AddError 添加错误
func (c *EvalContext) AddError(err error) {
	c.Errors = append(c.Errors, err)
}

// FieldRef 字段引用节点
type FieldRef struct {
	FieldName string
}

func (f *FieldRef) Evaluate(ctx *EvalContext) (bool, error) {
	val, ok := ctx.GetValue(f.FieldName)
	if !ok {
		return false, fmt.Errorf("字段 '%s' 不存在", f.FieldName)
	}

	// 将值转换为布尔值
	return toBool(val), nil
}

func (f *FieldRef) String() string {
	return f.FieldName
}

func (f *FieldRef) DependentFields() []string {
	return []string{f.FieldName}
}

// Literal 字面量节点
type Literal struct {
	Value interface{}
}

func (l *Literal) Evaluate(ctx *EvalContext) (bool, error) {
	return toBool(l.Value), nil
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

func (l *Literal) DependentFields() []string {
	return []string{}
}

// BinaryOp 二元操作节点
type BinaryOp struct {
	Left     Condition
	Right    Condition
	Operator string
}

func (b *BinaryOp) Evaluate(ctx *EvalContext) (bool, error) {
	// in 操作符特殊处理
	if b.Operator == "in" {
		return checkIn(ctx, b.Left, b.Right)
	}

	leftVal, leftErr := b.Left.Evaluate(ctx)
	if leftErr != nil {
		return false, leftErr
	}

	// 或操作符的短路优化：如果左侧为 true，直接返回 true
	if b.Operator == "|" || b.Operator == "||" || b.Operator == "or" {
		if leftVal {
			// 短路：左侧为 true，整体为 true，不需要评估右侧
			return true, nil
		}
	}

	// 与操作符的短路优化：如果左侧为 false，直接返回 false
	if b.Operator == "&" || b.Operator == "&&" || b.Operator == "and" {
		if !leftVal {
			// 短路：左侧为 false，整体为 false，不需要评估右侧
			return false, nil
		}
	}

	rightVal, rightErr := b.Right.Evaluate(ctx)
	if rightErr != nil {
		return false, rightErr
	}

	switch b.Operator {
	case "==", "=":
		return compareEqual(ctx, b.Left, b.Right)
	case "!=":
		result, err := compareEqual(ctx, b.Left, b.Right)
		return !result, err
	case ">":
		return compareNumeric(ctx, b.Left, b.Right, func(l, r float64) bool { return l > r })
	case ">=":
		return compareNumeric(ctx, b.Left, b.Right, func(l, r float64) bool { return l >= r })
	case "<":
		return compareNumeric(ctx, b.Left, b.Right, func(l, r float64) bool { return l < r })
	case "<=":
		return compareNumeric(ctx, b.Left, b.Right, func(l, r float64) bool { return l <= r })
	case "&", "&&", "and":
		return leftVal && rightVal, nil
	case "|", "||", "or":
		return leftVal || rightVal, nil
	case "in":
		return checkIn(ctx, b.Left, b.Right)
	default:
		return false, fmt.Errorf("不支持的操作符: %s", b.Operator)
	}
}

func (b *BinaryOp) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}

func (b *BinaryOp) DependentFields() []string {
	left := b.Left.DependentFields()
	right := b.Right.DependentFields()
	return append(left, right...)
}

// UnaryOp 一元操作节点
type UnaryOp struct {
	Operand  Condition
	Operator string
}

func (u *UnaryOp) Evaluate(ctx *EvalContext) (bool, error) {
	val, err := u.Operand.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	switch u.Operator {
	case "!", "not":
		return !val, nil
	default:
		return false, fmt.Errorf("不支持的一元操作符: %s", u.Operator)
	}
}

func (u *UnaryOp) String() string {
	return fmt.Sprintf("%s%s", u.Operator, u.Operand.String())
}

func (u *UnaryOp) DependentFields() []string {
	return u.Operand.DependentFields()
}

// ListLiteral 列表字面量节点（用于 in 操作符）
type ListLiteral struct {
	Values []interface{}
}

func (l *ListLiteral) Evaluate(ctx *EvalContext) (bool, error) {
	return false, fmt.Errorf("列表字面量不能直接评估")
}

func (l *ListLiteral) String() string {
	strs := make([]string, len(l.Values))
	for i, v := range l.Values {
		strs[i] = fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

func (l *ListLiteral) DependentFields() []string {
	return []string{}
}

// RangeLiteral 范围字面量节点（用于 between 操作符）
type RangeLiteral struct {
	Min interface{}
	Max interface{}
}

func (r *RangeLiteral) Evaluate(ctx *EvalContext) (bool, error) {
	return false, fmt.Errorf("范围字面量不能直接评估")
}

func (r *RangeLiteral) String() string {
	return fmt.Sprintf("%v..%v", r.Min, r.Max)
}

func (r *RangeLiteral) DependentFields() []string {
	return []string{}
}

// BetweenOp 范围操作节点（between 操作符）
type BetweenOp struct {
	Field Condition
	Range *RangeLiteral
}

func (b *BetweenOp) Evaluate(ctx *EvalContext) (bool, error) {
	fieldRef, ok := b.Field.(*FieldRef)
	if !ok {
		return false, fmt.Errorf("between 操作符左侧必须是字段引用")
	}

	val, ok := ctx.GetValue(fieldRef.FieldName)
	if !ok {
		return false, fmt.Errorf("字段 '%s' 不存在", fieldRef.FieldName)
	}

	valFloat, err := toFloat64(val)
	if err != nil {
		return false, fmt.Errorf("字段 '%s' 的值 %v 不是数值类型", fieldRef.FieldName, val)
	}

	minFloat, err := toFloat64(b.Range.Min)
	if err != nil {
		return false, fmt.Errorf("范围最小值 %v 不是数值类型", b.Range.Min)
	}

	maxFloat, err := toFloat64(b.Range.Max)
	if err != nil {
		return false, fmt.Errorf("范围最大值 %v 不是数值类型", b.Range.Max)
	}

	return valFloat >= minFloat && valFloat <= maxFloat, nil
}

func (b *BetweenOp) String() string {
	return fmt.Sprintf("%s between %s", b.Field.String(), b.Range.String())
}

func (b *BetweenOp) DependentFields() []string {
	return b.Field.DependentFields()
}

// ConditionParser 条件解析器
type ConditionParser struct {
	tokens []token
	pos    int
}

type token struct {
	Type    TokenType
	Literal string
}

type TokenType int

const (
	TokenField TokenType = iota
	TokenNumber
	TokenString
	TokenOperator
	TokenComma
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenKeyword
	TokenEOF
)

// ParseCondition 解析条件表达式
func ParseCondition(expr string) (Condition, error) {
	parser := &ConditionParser{
		tokens: tokenize(expr),
		pos:    0,
	}

	cond, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	if parser.pos < len(parser.tokens) {
		return nil, fmt.Errorf("条件表达式末尾有多余的标记")
	}

	return cond, nil
}

// parseExpression 解析表达式（处理逻辑操作符）
func (p *ConditionParser) parseExpression() (Condition, error) {
	return p.parseLogicalOr()
}

// parseLogicalOr 解析逻辑或表达式
func (p *ConditionParser) parseLogicalOr() (Condition, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == TokenOperator && (p.peek().Literal == "|" || p.peek().Literal == "||" || p.peek().Literal == "or") {
		op := p.next()
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Left: left, Right: right, Operator: op.Literal}
	}

	return left, nil
}

// parseLogicalAnd 解析逻辑与表达式
func (p *ConditionParser) parseLogicalAnd() (Condition, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == TokenOperator && (p.peek().Literal == "&" || p.peek().Literal == "&&" || p.peek().Literal == "and") {
		op := p.next()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Left: left, Right: right, Operator: op.Literal}
	}

	return left, nil
}

// parseComparison 解析比较表达式
func (p *ConditionParser) parseComparison() (Condition, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	// 检查是否是比较操作符
	if p.peek().Type == TokenOperator && isComparisonOperator(p.peek().Literal) {
		op := p.next()

		// 特殊处理 between 操作符
		if op.Literal == "between" {
			min, err := p.parseOperand()
			if err != nil {
				return nil, fmt.Errorf("between 操作符需要范围最小值: %w", err)
			}

			if p.peek().Type != TokenComma {
				return nil, fmt.Errorf("between 操作符需要逗号分隔最小值和最大值")
			}
			p.next() // 消耗逗号

			max, err := p.parseOperand()
			if err != nil {
				return nil, fmt.Errorf("between 操作符需要范围最大值: %w", err)
			}

			rang := &RangeLiteral{Min: getLiteralValue(min), Max: getLiteralValue(max)}
			return &BetweenOp{Field: left, Range: rang}, nil
		}

		right, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		return &BinaryOp{Left: left, Right: right, Operator: op.Literal}, nil
	}

	return left, nil
}

// parseUnary 解析一元操作符
func (p *ConditionParser) parseUnary() (Condition, error) {
	if p.peek().Type == TokenOperator && (p.peek().Literal == "!" || p.peek().Literal == "not") {
		op := p.next()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOp{Operand: operand, Operator: op.Literal}, nil
	}
	return p.parseOperand()
}

// parseOperand 解析操作数
func (p *ConditionParser) parseOperand() (Condition, error) {
	tok := p.peek()

	switch tok.Type {
	case TokenEOF:
		return nil, fmt.Errorf("意外的表达式结束")

	case TokenField:
		p.next()
		return &FieldRef{FieldName: tok.Literal}, nil

	case TokenNumber:
		p.next()
		num, _ := strconv.ParseFloat(tok.Literal, 64)
		return &Literal{Value: num}, nil

	case TokenString:
		p.next()
		return &Literal{Value: tok.Literal}, nil

	case TokenKeyword:
		p.next()
		// 处理布尔关键字
		switch strings.ToLower(tok.Literal) {
		case "true":
			return &Literal{Value: true}, nil
		case "false":
			return &Literal{Value: false}, nil
		default:
			return nil, fmt.Errorf("不支持的关键字: %s", tok.Literal)
		}

	case TokenLParen:
		p.next()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != TokenRParen {
			return nil, fmt.Errorf("缺少右括号")
		}
		p.next()
		return expr, nil

	case TokenLBracket:
		p.next()
		// 解析列表 [1,2,3]
		var values []interface{}
		for p.peek().Type != TokenRBracket {
			operand, err := p.parseOperand()
			if err != nil {
				return nil, fmt.Errorf("解析列表元素失败: %w", err)
			}
			values = append(values, getLiteralValue(operand))

			if p.peek().Type == TokenComma {
				p.next()
			}
		}
		if p.peek().Type != TokenRBracket {
			return nil, fmt.Errorf("列表缺少右括号")
		}
		p.next()
		return &ListLiteral{Values: values}, nil

	default:
		return nil, fmt.Errorf("意外的标记: %s", tok.Literal)
	}
}

// peek 查看当前标记
func (p *ConditionParser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

// next 消耗并返回当前标记
func (p *ConditionParser) next() token {
	tok := p.peek()
	p.pos++
	return tok
}

// tokenize 将表达式字符串转换为标记流
func tokenize(expr string) []token {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return []token{}
	}

	var tokens []token
	i := 0

	for i < len(expr) {
		ch := expr[i]

		switch {
		case isSpace(ch):
			i++

		case ch == '(':
			tokens = append(tokens, token{Type: TokenLParen, Literal: "("})
			i++

		case ch == ')':
			tokens = append(tokens, token{Type: TokenRParen, Literal: ")"})
			i++

		case ch == '[':
			tokens = append(tokens, token{Type: TokenLBracket, Literal: "["})
			i++

		case ch == ']':
			tokens = append(tokens, token{Type: TokenRBracket, Literal: "]"})
			i++

		case ch == ',':
			tokens = append(tokens, token{Type: TokenComma, Literal: ","})
			i++

		case ch == '!':
			if i+1 < len(expr) && expr[i+1] == '=' {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "!="})
				i += 2
			} else {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "!"})
				i++
			}

		case ch == '&':
			if i+1 < len(expr) && expr[i+1] == '&' {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "&&"})
				i += 2
			} else {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "&"})
				i++
			}

		case ch == '|':
			if i+1 < len(expr) && expr[i+1] == '|' {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "||"})
				i += 2
			} else {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "|"})
				i++
			}

		case ch == '=':
			if i+1 < len(expr) && expr[i+1] == '=' {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "=="})
				i += 2
			} else {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "="})
				i++
			}

		case ch == '>':
			if i+1 < len(expr) && expr[i+1] == '=' {
				tokens = append(tokens, token{Type: TokenOperator, Literal: ">="})
				i += 2
			} else {
				tokens = append(tokens, token{Type: TokenOperator, Literal: ">"})
				i++
			}

		case ch == '<':
			if i+1 < len(expr) && expr[i+1] == '=' {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "<="})
				i += 2
			} else {
				tokens = append(tokens, token{Type: TokenOperator, Literal: "<"})
				i++
			}

		case isDigit(ch) || ch == '-':
			start := i
			if ch == '-' {
				i++
			}
			for i < len(expr) && (isDigit(expr[i]) || expr[i] == '.') {
				i++
			}
			tokens = append(tokens, token{Type: TokenNumber, Literal: expr[start:i]})

		case isLetter(ch):
			start := i
			for i < len(expr) && (isLetter(expr[i]) || isDigit(expr[i])) {
				i++
			}
			word := expr[start:i]

			// 检查是否是关键字
			if strings.ToLower(word) == "true" || strings.ToLower(word) == "false" {
				tokens = append(tokens, token{Type: TokenKeyword, Literal: word})
			} else if strings.ToLower(word) == "and" || strings.ToLower(word) == "or" || strings.ToLower(word) == "not" || strings.ToLower(word) == "in" || strings.ToLower(word) == "between" {
				tokens = append(tokens, token{Type: TokenOperator, Literal: strings.ToLower(word)})
			} else {
				tokens = append(tokens, token{Type: TokenField, Literal: word})
			}

		case ch == '"':
			start := i
			i++
			for i < len(expr) && expr[i] != '"' {
				i++
			}
			if i < len(expr) {
				i++
			}
			tokens = append(tokens, token{Type: TokenString, Literal: expr[start+1 : i-1]})

		default:
			// 跳过未知字符
			i++
		}
	}

	return tokens
}

// 辅助函数

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isComparisonOperator(op string) bool {
	switch op {
	case "==", "=", "!=", ">", ">=", "<", "<=", "in", "between":
		return true
	default:
		return false
	}
}

func toBool(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return false
	}
}

func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("无法转换为数值: %T", v)
	}
}

func compareEqual(ctx *EvalContext, left, right Condition) (bool, error) {
	leftVal := getLiteralValue(left)
	rightVal := getLiteralValue(right)

	// 处理字段引用
	if fieldRef, ok := left.(*FieldRef); ok {
		val, exists := ctx.GetValue(fieldRef.FieldName)
		if !exists {
			return false, fmt.Errorf("字段 '%s' 不存在", fieldRef.FieldName)
		}
		leftVal = val
	}

	if fieldRef, ok := right.(*FieldRef); ok {
		val, exists := ctx.GetValue(fieldRef.FieldName)
		if !exists {
			return false, fmt.Errorf("字段 '%s' 不存在", fieldRef.FieldName)
		}
		rightVal = val
	}

	return compareValues(leftVal, rightVal) == 0, nil
}

func compareNumeric(ctx *EvalContext, left, right Condition, compare func(float64, float64) bool) (bool, error) {
	leftVal := getLiteralValue(left)
	rightVal := getLiteralValue(right)

	// 处理字段引用
	if fieldRef, ok := left.(*FieldRef); ok {
		val, exists := ctx.GetValue(fieldRef.FieldName)
		if !exists {
			return false, fmt.Errorf("字段 '%s' 不存在", fieldRef.FieldName)
		}
		leftVal = val
	}

	if fieldRef, ok := right.(*FieldRef); ok {
		val, exists := ctx.GetValue(fieldRef.FieldName)
		if !exists {
			return false, fmt.Errorf("字段 '%s' 不存在", fieldRef.FieldName)
		}
		rightVal = val
	}

	leftFloat, err := toFloat64(leftVal)
	if err != nil {
		return false, fmt.Errorf("左侧值 %v 不是数值类型", leftVal)
	}

	rightFloat, err := toFloat64(rightVal)
	if err != nil {
		return false, fmt.Errorf("右侧值 %v 不是数值类型", rightVal)
	}

	return compare(leftFloat, rightFloat), nil
}

func checkIn(ctx *EvalContext, left, right Condition) (bool, error) {
	list, ok := right.(*ListLiteral)
	if !ok {
		return false, fmt.Errorf("in 操作符右侧必须是列表")
	}

	// 获取左侧的值
	var leftVal interface{}
	if fieldRef, ok := left.(*FieldRef); ok {
		val, exists := ctx.GetValue(fieldRef.FieldName)
		if !exists {
			return false, fmt.Errorf("字段 '%s' 不存在", fieldRef.FieldName)
		}
		leftVal = val
	} else {
		leftVal = getLiteralValue(left)
	}

	// 检查是否在列表中
	for _, item := range list.Values {
		if compareValues(leftVal, item) == 0 {
			return true, nil
		}
	}
	return false, nil
}

func getLiteralValue(cond Condition) interface{} {
	switch c := cond.(type) {
	case *Literal:
		return c.Value
	case *ListLiteral:
		return c.Values
	case *RangeLiteral:
		return []interface{}{c.Min, c.Max}
	default:
		return nil
	}
}

func compareValues(a, b interface{}) int {
	// 简单的值比较，返回 -1, 0, 1
	switch va := a.(type) {
	case int:
		switch vb := b.(type) {
		case int:
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			}
			return 0
		case int64:
			// int 与 int64 比较
			ia := int64(va)
			if ia < vb {
				return -1
			} else if ia > vb {
				return 1
			}
			return 0
		case float64:
			fa := float64(va)
			if fa < vb {
				return -1
			} else if fa > vb {
				return 1
			}
			return 0
		case string:
			// int 与 string 比较（处理 CSV 字符串）
			ib, err := strconv.ParseInt(vb, 10, 64)
			if err != nil {
				return 0 // 类型不匹配
			}
			ia := int64(va)
			if ia < ib {
				return -1
			} else if ia > ib {
				return 1
			}
			return 0
		}
	case int64:
		switch vb := b.(type) {
		case int:
			ib := int64(vb)
			if va < ib {
				return -1
			} else if va > ib {
				return 1
			}
			return 0
		case int64:
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			}
			return 0
		case float64:
			fa := float64(va)
			if fa < vb {
				return -1
			} else if fa > vb {
				return 1
			}
			return 0
		case string:
			ib, err := strconv.ParseInt(vb, 10, 64)
			if err != nil {
				return 0
			}
			if va < ib {
				return -1
			} else if va > ib {
				return 1
			}
			return 0
		}
	case float64:
		switch vb := b.(type) {
		case int:
			fb := float64(vb)
			if va < fb {
				return -1
			} else if va > fb {
				return 1
			}
			return 0
		case int64:
			fb := float64(vb)
			if va < fb {
				return -1
			} else if va > fb {
				return 1
			}
			return 0
		case float64:
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			}
			return 0
		case string:
			fb, err := strconv.ParseFloat(vb, 64)
			if err != nil {
				return 0
			}
			if va < fb {
				return -1
			} else if va > fb {
				return 1
			}
			return 0
		}
	case string:
		if vb, ok := b.(string); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			}
			return 0
		}
		// 处理 string 与数值类型的比较（如从 CSV 读取的字符串 "1" 与数值 1 比较）
		switch vb := b.(type) {
		case int:
			// 尝试将 string 转换为 int 比较
			ia, err := strconv.ParseInt(va, 10, 64)
			if err != nil {
				return 0 // 类型不匹配，不相等
			}
			ib := int64(vb)
			if ia < ib {
				return -1
			} else if ia > ib {
				return 1
			}
			return 0
		case int64:
			ia, err := strconv.ParseInt(va, 10, 64)
			if err != nil {
				return 0
			}
			if ia < vb {
				return -1
			} else if ia > vb {
				return 1
			}
			return 0
		case float64:
			fa, err := strconv.ParseFloat(va, 64)
			if err != nil {
				return 0
			}
			if fa < vb {
				return -1
			} else if fa > vb {
				return 1
			}
			return 0
		}
	}
	return 0
}
