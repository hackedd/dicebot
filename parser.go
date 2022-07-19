package dicebot

import (
	"fmt"
	"math/rand"
	"sort"
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

type Lookup func(name string) (Expr, error)

type Expr interface {
	String() string
	eval(lookup Lookup, depth int) (int, error)
	explain(lookup Lookup, depth int) string
}

type NumberExpr struct {
	Value int
}

type DiceExpr struct {
	Number int
	Sides  int
	Rolled []int
}

type VariableExpr struct {
	Name  string
	Value Expr
}

type BestOfExpr struct {
	Number int
	Of     *DiceExpr
	Sorted []int
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

func (e *NumberExpr) eval(lookup Lookup, depth int) (int, error) {
	return e.Value, nil
}

func (e *NumberExpr) explain(lookup Lookup, depth int) string {
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

func (e *DiceExpr) eval(lookup Lookup, depth int) (int, error) {
	e.Roll()

	t := 0
	for _, r := range e.Rolled {
		t += r
	}
	return t, nil
}

func (e *DiceExpr) explain(lookup Lookup, depth int) string {
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

func (e *VariableExpr) String() string {
	return e.Name
}

func (e *VariableExpr) Lookup(lookup Lookup) error {
	if e.Value != nil {
		return nil
	}

	expr, err := lookup(e.Name)
	if err != nil {
		return err
	}

	e.Value = expr
	return nil
}

func (e *VariableExpr) eval(lookup Lookup, depth int) (int, error) {
	err := e.Lookup(lookup)
	if err == nil {
		return eval(e.Value, lookup, depth)
	}
	return 0, err
}

func (e *VariableExpr) explain(lookup Lookup, depth int) string {
	err := e.Lookup(lookup)
	if err == nil {
		return explain(e.Value, lookup, depth)
	}
	return "undef"
}

func (e *BestOfExpr) String() string {
	if e.Number == 1 {
		return fmt.Sprintf("best of %s", e.Of)
	} else {
		return fmt.Sprintf("best %d of %s", e.Number, e.Of)
	}
}

func (e *BestOfExpr) Roll() {
	if e.Sorted != nil {
		return
	}

	e.Of.Roll()

	e.Sorted = make([]int, e.Of.Number)
	copy(e.Sorted, e.Of.Rolled)
	sort.Sort(sort.Reverse(sort.IntSlice(e.Sorted)))
}

func (e *BestOfExpr) eval(lookup Lookup, depth int) (int, error) {
	e.Roll()

	t := 0
	for _, r := range e.Sorted[:e.Number] {
		t += r
	}
	return t, nil
}

func indexOf(arr []int, needle int) int {
	for i, element := range arr {
		if element == needle {
			return i
		}
	}
	return -1
}

func (e *BestOfExpr) explain(lookup Lookup, depth int) string {
	e.Roll()

	kept := make([]int, e.Number)
	copy(kept, e.Sorted[:e.Number])

	rolled := make([]string, e.Of.Number)
	for i, r := range e.Of.Rolled {
		if j := indexOf(kept, r); j >= 0 {
			rolled[i] = fmt.Sprintf("__%d__", r)
			kept[j] = 0
		} else {
			rolled[i] = fmt.Sprintf("%d", r)
		}
	}

	if e.Number == 1 {
		return fmt.Sprintf("best of (%s)", strings.Join(rolled, ", "))
	} else {
		return fmt.Sprintf("best %d of (%s)", e.Number, strings.Join(rolled, ", "))
	}
}

func (e *UnaryExpr) String() string {
	return fmt.Sprintf("(%s %s)", e.OpName, e.Value.String())
}

func (e *UnaryExpr) eval(lookup Lookup, depth int) (int, error) {
	value, err := eval(e.Value, lookup, depth)
	if err != nil {
		return 0, err
	}
	return e.Operator(value), nil
}

func (e *UnaryExpr) explain(lookup Lookup, depth int) string {
	return fmt.Sprintf("%s%s", e.OpName, explain(e.Value, lookup, depth))
}

func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.OpName, e.Left.String(), e.Right.String())
}

func (e *BinaryExpr) eval(lookup Lookup, depth int) (int, error) {
	left, err := eval(e.Left, lookup, depth)
	if err != nil {
		return 0, err
	}
	right, err := eval(e.Right, lookup, depth)
	if err != nil {
		return 0, err
	}
	return e.Operator(left, right), nil
}

func (e *BinaryExpr) explain(lookup Lookup, depth int) string {
	return fmt.Sprintf("%s %s %s", explain(e.Left, lookup, depth), e.OpName, explain(e.Right, lookup, depth))
}

func (e *ParenExpr) String() string {
	return e.Expr.String()
}

func (e *ParenExpr) eval(lookup Lookup, depth int) (int, error) {
	return eval(e.Expr, lookup, depth)
}

func (e *ParenExpr) explain(lookup Lookup, depth int) string {
	return fmt.Sprintf("(%s)", explain(e.Expr, lookup, depth))
}

type nudFunc func(parser *Parser, token Token) (Expr, error)
type ledFunc func(parser *Parser, token Token, left Expr) (Expr, error)

type pratt struct {
	lbp int
	nud nudFunc
	led ledFunc
}

func errorNud(parser *Parser, token Token) (Expr, error) {
	return nil, ParseError{"Unexpected input", token.Position}
}

func errorLed(parser *Parser, token Token, left Expr) (Expr, error) {
	return nil, ParseError{"Unexpected input", token.Position}
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

	number := 1
	if token.Matches[1] != "" {
		if number, err = strconv.Atoi(token.Matches[1]); err != nil {
			return nil, ParseError{err.Error(), token.MatchPosition(1)}
		}
		if number == 0 {
			return nil, ParseError{"Can't roll zero dice", token.MatchPosition(1)}
		}
		if number > 100 {
			return nil, ParseError{"Can't roll more than 100 dice", token.MatchPosition(1)}
		}
	}

	sides := 6
	if token.Matches[2] != "" {
		if sides, err = strconv.Atoi(token.Matches[2]); err != nil {
			return nil, ParseError{err.Error(), token.MatchPosition(2)}
		}
		if sides == 0 {
			return nil, ParseError{"Can't roll zero-sided dice", token.MatchPosition(2)}
		}
	}

	return &DiceExpr{number, sides, nil}, nil
}

func identifierNud(parser *Parser, token Token) (Expr, error) {
	return &VariableExpr{Name: token.Text}, nil
}

func bestOfNud(parser *Parser, token Token) (Expr, error) {
	var err error

	number := 1
	if token.Matches[1] != "" {
		number, err = strconv.Atoi(token.Matches[1])
		if err != nil {
			return nil, ParseError{err.Error(), token.MatchPosition(1)}
		}
		if number == 0 {
			return nil, ParseError{"Can't keep zero dice", token.MatchPosition(1)}
		}
	}

	expr, err := diceNud(parser, Token{DICE, token.Matches[2], token.Position, token.Matches[2:], token.Indices[2:]})
	if err != nil {
		return nil, err
	}

	diceExpr := expr.(*DiceExpr)
	if number > diceExpr.Number {
		return nil, ParseError{fmt.Sprintf("Can't keep more than %d dice", diceExpr.Number), token.MatchPosition(1)}
	}
	if number == diceExpr.Number {
		return nil, ParseError{fmt.Sprintf("It doesn't make sense to keep %d of %d dice", number, number), token.MatchPosition(1)}
	}

	return &BestOfExpr{number, diceExpr, nil}, nil
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
		LEFT_PAREN:  {50, parenNud, errorLed},
		RIGHT_PAREN: {0, errorNud, errorLed},
		PLUS:        {10, prefixNud(unaryPlus), infixLed(plus)},
		MINUS:       {10, prefixNud(unaryMinus), infixLed(minus)},
		MULTIPLY:    {20, errorNud, infixLed(multiply)},
		DIVIDE:      {20, errorNud, infixLed(divide)},
		NUMBER:      {0, numberNud, errorLed},
		DICE:        {0, diceNud, errorLed},
		IDENTIFIER:  {0, identifierNud, errorLed},
		BEST_OF:     {0, bestOfNud, errorLed},
		END:         {0, errorNud, errorLed},
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

	expr, err := parser.parseExpression(0)
	if err != nil {
		return nil, err
	}

	token, _ := parser.getCurrent()
	if token.Type != END {
		return nil, ParseError{"Unexpected input", token.Position}
	}

	return expr, nil
}

func ParseString(input string) (Expr, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return nil, err
	}

	return Parse(tokens)
}

const MaxDepth = 50

func Eval(expr Expr, lookup Lookup) (int, error) {
	return eval(expr, lookup, 0)
}

func Explain(expr Expr, lookup Lookup) string {
	return explain(expr, lookup, 0)
}

func eval(expr Expr, lookup Lookup, depth int) (int, error) {
	if depth >= MaxDepth {
		return 0, ParseError{"Expression too complex", 0}
	}
	return expr.eval(lookup, depth+1)
}

func explain(expr Expr, lookup Lookup, depth int) string {
	if depth >= MaxDepth {
		return "too complex"
	}
	return expr.explain(lookup, depth+1)
}
