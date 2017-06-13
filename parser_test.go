package dicebot

import (
	"math/rand"
	"testing"
)

type parseExample struct {
	input    string
	expected string
	value    int
}

var parseExamples = []parseExample{
	{"3d6 + 2", "(+ 3d6 2)", 18},
	{"1 + 2 + 3", "(+ (+ 1 2) 3)", 6},
	{"1 + 2 * 3", "(+ 1 (* 2 3))", 7},
	{"(1 + 2)", "(+ 1 2)", 3},
	{"1 * (2 + 3)", "(* 1 (+ 2 3))", 5},
	{"-1 + 2", "(+ (- 1) 2)", 1},
	{"+-1", "(+ (- 1))", -1},
	{"10 / 2 - 1", "(- (/ 10 2) 1)", 4},
	{"d20", "1d20", 2},
	{"20d6", "20d6", 75},
	{"(1 + 2) + 3", "(+ (+ 1 2) 3)", 6},
	{"a + b", "(+ a b)", 3},
	{"c * d6", "(* c 1d6)", 18},
	{"x", "x", 123},
}

func testLookup(name string) (Expr, bool) {
	if name == "r" {
		return &DiceExpr{Number: 2, Sides: 6}, true
	}
	if name == "x" {
		return &VariableExpr{Name: "y"}, true
	}
	if name == "y" {
		return &VariableExpr{Name: "z"}, true
	}
	if name == "z" {
		return &NumberExpr{Value: 123}, true
	}
	if name >= "a" && name <= "c" {
		return &NumberExpr{Value: int(name[0]-'a') + 1}, true
	}
	return nil, false
}

func TestParse(t *testing.T) {
	for _, example := range parseExamples {
		rand.Seed(1)

		tokens, err := Tokenize(example.input)
		if err != nil {
			t.Errorf("Tokenizing '%s' failed: %s", example.input, err)
			continue
		}

		actual, err := Parse(tokens)
		if err != nil {
			t.Errorf("Parsing '%s' failed: %s", example.input, err)
			continue
		}

		if actual.String() != example.expected {
			t.Errorf("Parsing '%s' failed: expected '%s' got '%s'", example.input, example.expected, actual.String())
			continue
		}

		value, err := actual.Eval(testLookup)
		if err != nil {
			t.Errorf("Evaluating '%s' failed: %s", example.input, err)
			continue
		}
		if value != example.value {
			t.Errorf("Evaluating '%s' failed: expected %d, got %d", example.input, example.value, value)
			continue
		}

		secondValue, err := actual.Eval(testLookup)
		if err != nil {
			t.Errorf("Evaluating '%s' failed: %s", example.input, err)
			continue
		}
		if secondValue != value {
			t.Errorf("Evaluating '%s' failed: value changed (%d != %d)", example.input, secondValue, value)
			continue
		}
	}
}

type parseErrorExample struct {
	input    string
	expected string
}

var parseErrorExamples = []parseErrorExample{
	{"", "Empty input near position 0"},
	{"(1", "Expected ) near position 2"},
	{"1 +", "Unexpected input near position 3"},
	{")", "Unexpected input near position 0"},
	{"1 1", "Unexpected input near position 2"},
}

func TestParseErrors(t *testing.T) {
	for _, example := range parseErrorExamples {
		tokens, err := Tokenize(example.input)
		if err != nil {
			t.Errorf("Tokenizing '%s' failed: %s", example.input, err)
			continue
		}

		actual, err := Parse(tokens)
		if err == nil {
			t.Errorf("Parsing '%s' unexpectedly succeeded: %s", example.input, actual)
		} else if err.Error() != example.expected {
			t.Errorf("Parsing '%s' failed: expected '%s' got '%s'", example.input, example.expected, err.Error())
		}
	}
}

type explainExample struct {
	input    string
	expected string
}

var explainExamples = []explainExample{
	{"d6", "6"},
	{"3d6", "(6 + 4 + 6)"},
	{"1 + 2", "1 + 2"},
	{"-5", "-5"},
	{"1 * (2 + 3)", "1 * (2 + 3)"},
	{"a*b", "1 * 2"},
	{"r+r", "(6 + 4) + (6 + 6)"},
	{"qux", "undef"},
}

func TestExplain(t *testing.T) {
	for _, example := range explainExamples {
		rand.Seed(1)

		tokens, err := Tokenize(example.input)
		if err != nil {
			t.Errorf("Tokenizing '%s' failed: %s", example.input, err)
			continue
		}

		expr, err := Parse(tokens)
		if err != nil {
			t.Errorf("Parsing '%s' failed: %s", example.input, err)
			continue
		}

		actual := expr.Explain(testLookup)
		if actual != example.expected {
			t.Errorf("Explaining '%s' failed: %s does not match %s", example.input, actual, example.expected)
		}
	}
}
