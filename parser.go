package dicebot

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type TokenIter struct {
	tokens   []Token
	position int
}

func (it *TokenIter) next() Token {
	t := it.tokens[it.position]
	it.position += 1
	return t
}

func (it *TokenIter) peek() Token {
	return it.tokens[it.position]
}

type Expr interface {
	String() string
	Eval() int
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

func (e *NumberExpr) String() string {
	return fmt.Sprintf("%d", e.Value)
}

func (e *NumberExpr) Eval() int {
	return e.Value
}

func (e *DiceExpr) String() string {
	return fmt.Sprintf("%dd%d", e.Number, e.Sides)
}

func (e *DiceExpr) Eval() int {
	if e.Rolled == nil {
		e.Rolled = make([]int, e.Sides)
		for i := 0; i < e.Number; i += 1 {
			e.Rolled[i] = rand.Intn(e.Sides) + 1
		}
	}
	t := 0
	for _, r := range e.Rolled {
		t += r
	}
	return t
}

func (e *UnaryExpr) String() string {
	return fmt.Sprintf("(%s %s)", e.OpName, e.Value.String())
}

func (e *UnaryExpr) Eval() int {
	return e.Operator(e.Value.Eval())
}

func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.OpName, e.Left.String(), e.Right.String())
}

func (e *BinaryExpr) Eval() int {
	return e.Operator(e.Left.Eval(), e.Right.Eval())
}

type nudFunc func(token Token, it *TokenIter) (Expr, error)
type ledFunc func(left Expr, token Token, it *TokenIter) (Expr, error)

type pratt struct {
	lbp int
	nud nudFunc
	led ledFunc
}

func parenNud(token Token, it *TokenIter) (Expr, error) {
	expr, err := parseExpression(it, 0)
	if err != nil {
		return nil, err
	}
	if it.peek().Type != RIGHT_PAREN {
		return nil, ParseError{"Expected )", it.peek().Position}
	}
	return expr, nil
}

func numberNud(token Token, it *TokenIter) (Expr, error) {
	value, err := strconv.Atoi(token.Text)
	if err != nil {
		return nil, ParseError{err.Error(), token.Position}
	}
	return &NumberExpr{value}, nil
}

func diceNud(token Token, it *TokenIter) (Expr, error) {
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
	return func(token Token, it *TokenIter) (Expr, error) {
		left, err := parseExpression(it, 100)
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{token.Text, operator, left}, nil
	}
}

func infixLed(operator BinaryFunc) ledFunc {
	return func(left Expr, token Token, it *TokenIter) (Expr, error) {
		right, err := parseExpression(it, tokens[token.Type].lbp)
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

func getCurrent(it *TokenIter) (Token, *pratt) {
	token := it.peek()
	pratt, ok := tokens[token.Type]
	if !ok {
		panic(fmt.Errorf("Parse error, unknown token %+v", token))
	}
	return token, pratt
}

func parseExpression(it *TokenIter, rbp int) (Expr, error) {
	token, pratt := getCurrent(it)
	it.next()
	left, err := pratt.nud(token, it)
	if err != nil {
		return nil, err
	}

	token, pratt = getCurrent(it)
	for rbp < pratt.lbp {
		it.next()
		left, err = pratt.led(left, token, it)
		if err != nil {
			return nil, err
		}
		token, pratt = getCurrent(it)
	}

	return left, nil
}

func Parse(t []Token) (Expr, error) {
	if t[0].Type == END {
		return nil, ParseError{"Empty input", 0}
	}
	return parseExpression(&TokenIter{t, 0}, 0)
}
