package dicebot

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type Parser struct {
	tokens   []Token
	position int
}

func (parser *Parser) next() Token {
	t := parser.tokens[parser.position]
	parser.position += 1
	return t
}

func (parser *Parser) peek() Token {
	return parser.tokens[parser.position]
}

type Expr interface {
	String() string
	Eval() int
	Explain() string
}

type NumberExpr struct {
	Value int
}

type DiceExpr struct {
	Number int
	Sides  int
	Rolled []int
}

type UnaryFunc func(value int) int

type UnaryExpr struct {
	OpName   string
	Operator UnaryFunc
	Value    Expr
}

type BinaryFunc func(int, int) int

type BinaryExpr struct {
	OpName   string
	Operator BinaryFunc
	Left     Expr
	Right    Expr
}

type ParenExpr struct {
	Expr Expr
}

func (e *NumberExpr) String() string {
	return fmt.Sprintf("%d", e.Value)
}

func (e *NumberExpr) Eval() int {
	return e.Value
}

func (e *NumberExpr) Explain() string {
	return e.String()
}

func (e *DiceExpr) String() string {
	return fmt.Sprintf("%dd%d", e.Number, e.Sides)
}

func (e *DiceExpr) Roll() {
	if e.Rolled != nil {
		return
	}

	e.Rolled = make([]int, e.Number)
	for i := 0; i < e.Number; i += 1 {
		e.Rolled[i] = rand.Intn(e.Sides) + 1
	}
}

func (e *DiceExpr) Eval() int {
	e.Roll()

	t := 0
	for _, r := range e.Rolled {
		t += r
	}
	return t
}

func (e *DiceExpr) Explain() string {
	e.Roll()

	if e.Number == 1 {
		return fmt.Sprintf("%d", e.Rolled[0])
	}

	parts := make([]string, e.Number)
	for i, r := range e.Rolled {
		parts[i] = fmt.Sprintf("%d", r)
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, " + "))
}

func (e *UnaryExpr) String() string {
	return fmt.Sprintf("(%s %s)", e.OpName, e.Value.String())
}

func (e *UnaryExpr) Eval() int {
	return e.Operator(e.Value.Eval())
}

func (e *UnaryExpr) Explain() string {
	return fmt.Sprintf("%s%s", e.OpName, e.Value.Explain())
}

func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.OpName, e.Left.String(), e.Right.String())
}

func (e *BinaryExpr) Eval() int {
	return e.Operator(e.Left.Eval(), e.Right.Eval())
}

func (e *BinaryExpr) Explain() string {
	return fmt.Sprintf("%s %s %s", e.Left.Explain(), e.OpName, e.Right.Explain())
}

func (e *ParenExpr) String() string {
	return e.Expr.String()
}

func (e *ParenExpr) Eval() int {
	return e.Expr.Eval()
}

func (e *ParenExpr) Explain() string {
	return fmt.Sprintf("(%s)", e.Expr.Explain())
}

type nudFunc func(parser *Parser, token Token) (Expr, error)
type ledFunc func(parser *Parser, token Token, left Expr) (Expr, error)

type pratt struct {
	lbp int
	nud nudFunc
	led ledFunc
}

func parenNud(parser *Parser, token Token) (Expr, error) {
	expr, err := parser.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if parser.peek().Type != RIGHT_PAREN {
		return nil, ParseError{"Expected )", parser.peek().Position}
	}
	parser.next()
	return &ParenExpr{expr}, nil
}

func numberNud(parser *Parser, token Token) (Expr, error) {
	value, err := strconv.Atoi(token.Text)
	if err != nil {
		return nil, ParseError{err.Error(), token.Position}
	}
	return &NumberExpr{value}, nil
}

func diceNud(parser *Parser, token Token) (Expr, error) {
	var err error

	parts := strings.Split(token.Text, "d")

	number := 1
	if parts[0] != "" {
		if number, err = strconv.Atoi(parts[0]); err != nil {
			return nil, ParseError{err.Error(), token.Position}
		}
	}
	sides := 6
	if parts[1] != "" {
		if sides, err = strconv.Atoi(parts[1]); err != nil {
			return nil, ParseError{err.Error(), token.Position}
		}
	}

	return &DiceExpr{number, sides, nil}, nil
}

func prefixNud(operator UnaryFunc) nudFunc {
	return func(parser *Parser, token Token) (Expr, error) {
		left, err := parser.parseExpression(100)
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{token.Text, operator, left}, nil
	}
}

func infixLed(operator BinaryFunc) ledFunc {
	return func(parser *Parser, token Token, left Expr) (Expr, error) {
		right, err := parser.parseExpression(tokens[token.Type].lbp)
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{token.Text, operator, left, right}, nil
	}
}

func unaryPlus(value int) int {
	return +value
}

func unaryMinus(value int) int {
	return -value
}

func plus(left, right int) int {
	return left + right
}

func minus(left, right int) int {
	return left - right
}

func multiply(left, right int) int {
	return left * right
}

func divide(left, right int) int {
	return left / right
}

var tokens map[TokenType]*pratt

func init() {
	tokens = map[TokenType]*pratt{
		LEFT_PAREN:  {50, parenNud, nil},
		RIGHT_PAREN: {0, nil, nil},
		PLUS:        {10, prefixNud(unaryPlus), infixLed(plus)},
		MINUS:       {10, prefixNud(unaryMinus), infixLed(minus)},
		MULTIPLY:    {20, nil, infixLed(multiply)},
		DIVIDE:      {20, nil, infixLed(divide)},
		NUMBER:      {0, numberNud, nil},
		DICE:        {0, diceNud, nil},
		END:         {0, nil, nil},
	}
}

func (parser *Parser) getCurrent() (Token, *pratt) {
	token := parser.peek()
	pratt, ok := tokens[token.Type]
	if !ok {
		panic(fmt.Errorf("Parse error, unknown token %+v", token))
	}
	return token, pratt
}

func (parser *Parser) parseExpression(rbp int) (Expr, error) {
	token, pratt := parser.getCurrent()
	parser.next()
	left, err := pratt.nud(parser, token)
	if err != nil {
		return nil, err
	}

	token, pratt = parser.getCurrent()
	for rbp < pratt.lbp {
		parser.next()
		left, err = pratt.led(parser, token, left)
		if err != nil {
			return nil, err
		}
		token, pratt = parser.getCurrent()
	}

	return left, nil
}

func Parse(t []Token) (Expr, error) {
	if t[0].Type == END {
		return nil, ParseError{"Empty input", 0}
	}
	parser := &Parser{t, 0}
	return parser.parseExpression(0)
}
