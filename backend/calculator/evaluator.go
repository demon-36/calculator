package calculator

import (
	"math/big"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	endToken      rune = -1
	numberToken   rune = -2
	maxNesting         = 32
	maxOperations      = 128
	maxBodyBytes       = 16 << 10
)

var decimalLiteral = regexp.MustCompile(`^([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?`)

type exprNode struct {
	op          rune
	literal     string
	left, right *exprNode
}

type exprParser struct {
	remaining string
	token     rune
	literal   string
	ops       int
	depth     int
}

func Evaluate(input string) (float64, *Problem) {
	if len(input) > maxBodyBytes {
		return 0, SyntaxProblem("Expression is too long.")
	}
	parser := exprParser{remaining: input}
	parser.next()
	root, p := parser.parse(1)
	if p != nil {
		return 0, p
	}
	if parser.token != endToken {
		return 0, SyntaxProblem("Unexpected input after the expression.")
	}

	// Validate the whole expression before performing any arithmetic.
	budget := powerBudget(maxExactPowerBits)
	value, p := root.evaluate(&budget)
	if p != nil {
		return 0, p
	}
	return roundedResult(value)
}

func EvaluateExpression(input string) (float64, *Problem) {
	return Evaluate(input)
}

func (p *exprParser) next() {
	p.remaining = strings.TrimLeftFunc(p.remaining, unicode.IsSpace)
	if p.remaining == "" {
		p.token = endToken
		return
	}
	if literal := decimalLiteral.FindString(p.remaining); literal != "" {
		p.token, p.literal = numberToken, literal
		p.remaining = p.remaining[len(literal):]
		return
	}
	if after, found := strings.CutPrefix(p.remaining, "sqrt"); found {
		p.token, p.remaining = '√', after
		return
	}
	r, size := utf8.DecodeRuneInString(p.remaining)
	p.remaining = p.remaining[size:]
	switch r {
	case '×':
		r = '*'
	case '÷':
		r = '/'
	case '−':
		r = '-'
	}
	p.token = r
}

func (p *exprParser) takeOperator() *Problem {
	p.ops++
	if p.ops > maxOperations {
		return SyntaxProblem("Use at most 128 operations.")
	}
	p.next()
	return nil
}

const powerPrecedence = 3

func precedence(operator rune) int {
	switch operator {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	case '^':
		return powerPrecedence
	case '%':
		return 4
	default:
		return 0
	}
}

// The right operand of a power may contain another power. Every other
// binary operator stops before the next operator of the same precedence.
func (p *exprParser) parse(minPrecedence int) (*exprNode, *Problem) {
	left, err := p.prefix()
	if err != nil {
		return nil, err
	}
	for precedence(p.token) >= minPrecedence {
		op := p.token
		if err := p.takeOperator(); err != nil {
			return nil, err
		}
		if op == '%' {
			left = &exprNode{op: op, left: left}
			continue
		}
		nextMin := precedence(op) + 1
		if op == '^' {
			nextMin = powerPrecedence
		}
		right, err := p.parse(nextMin)
		if err != nil {
			return nil, err
		}
		left = &exprNode{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *exprParser) prefix() (*exprNode, *Problem) {
	op := p.token
	switch op {
	case numberToken:
		node := &exprNode{op: op, literal: p.literal}
		p.next()
		return node, nil
	case '-', '+':
		if err := p.takeOperator(); err != nil {
			return nil, err
		}
		value, err := p.parse(powerPrecedence)
		return &exprNode{op: op, right: value}, err
	case '√':
		if err := p.takeOperator(); err != nil {
			return nil, err
		}
		if p.token != '(' {
			return nil, SyntaxProblem("Enclose the square root argument in parentheses.")
		}
		group, err := p.group()
		return &exprNode{op: op, left: group}, err
	case '(':
		return p.group()
	default:
		return nil, SyntaxProblem("Expected a number or a parenthesized expression.")
	}
}

func (p *exprParser) group() (*exprNode, *Problem) {
	p.depth++
	if p.depth > maxNesting {
		return nil, SyntaxProblem("Use at most 32 levels of parentheses.")
	}
	p.next()
	value, err := p.parse(1)
	if err != nil {
		return nil, err
	}
	if p.token != ')' {
		return nil, SyntaxProblem("Expected a closing parenthesis.")
	}
	p.next()
	p.depth--
	return value, nil
}

func (n *exprNode) percentage() bool {
	for n.left == nil && (n.op == '+' || n.op == '-') {
		n = n.right
	}
	return n.op == '%'
}

func (n *exprNode) evaluate(budget *powerBudget) (*big.Rat, *Problem) {
	if n.op == numberToken {
		return parseNumber(n.literal)
	}
	if n.left == nil { // Unary sign
		v, err := n.right.evaluate(budget)
		if err == nil && n.op == '-' {
			v.Neg(v)
		}
		return v, err
	}
	a, err := n.left.evaluate(budget)
	if err != nil {
		return nil, err
	}
	if n.op == '√' {
		return squareRoot(a)
	}
	if n.op == '%' {
		a.Quo(a, big.NewRat(100, 1))
	} else {
		b, err := n.right.evaluate(budget)
		if err != nil {
			return nil, err
		}
		switch n.op {
		case '+', '-':
			if n.op == '-' {
				b.Neg(b)
			}
			if n.right.percentage() {
				a.Mul(a, b.Add(b, big.NewRat(1, 1)))
			} else {
				a.Add(a, b)
			}
		case '*':
			a.Mul(a, b)
		case '/':
			if b.Sign() == 0 {
				return nil, ZeroDivisor
			}
			a.Quo(a, b)
		case '^':
			return budget.pow(a, b)
		}
	}
	if !inRange(a) {
		return nil, RangeProblem
	}
	return a, nil
}
