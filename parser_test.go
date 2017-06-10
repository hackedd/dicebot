package dicebot

import (
	"regexp"
	"strings"
	"testing"
)

type parseExample struct {
	input    string
	expected string
	min      int
	max      int
}

var parseExamples = []parseExample{
	{"3d6 + 2", "(+ 3d6 2)", 3*1 + 2, 3*6 + 2},
	{"1 + 2 + 3", "(+ (+ 1 2) 3)", 6, 6},
	{"1 + 2 * 3", "(+ 1 (* 2 3))", 7, 7},
	{"(1 + 2)", "(+ 1 2)", 3, 3},
	{"1 * (2 + 3)", "(* 1 (+ 2 3))", 5, 5},
	{"-1 + 2", "(+ (- 1) 2)", 1, 1},
	{"+-1", "(+ (- 1))", -1, -1},
	{"10 / 2 - 1", "(- (/ 10 2) 1)", 4, 4},
	{"d20", "1d20", 1, 20},
	{"20d6", "20d6", 20 * 1, 20 * 6},
	{"(1 + 2) + 3", "(+ (+ 1 2) 3)", 6, 6},
	{"a + b", "(+ a b)", 3, 3},
	{"c * d6", "(* c 1d6)", 3 * 1, 3 * 6},
}

func testLookup(name string) int {
	return int(name[0]-'a') + 1
}

func TestParse(t *testing.T) {
	for _, example := range parseExamples {
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

		value := actual.Eval(testLookup)
		if value < example.min || value > example.max {
			t.Errorf("Evaluating '%s' failed: expected %d <= %d <= %d", example.input, example.min, value, example.max)
			continue
		}

		secondValue := actual.Eval(testLookup)
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

func MakeRegexp(expr string) *regexp.Regexp {
	re := strings.Replace(regexp.QuoteMeta(expr), "N", "\\d+", -1)
	return regexp.MustCompile("^" + re + "$")
}

var explainExamples = []explainExample{
	{"d6", "N"},
	{"3d6", "(N + N + N)"},
	{"1 + 2", "1 + 2"},
	{"-5", "-5"},
	{"1 * (2 + 3)", "1 * (2 + 3)"},
	{"a*b", "a * b"},
}

func TestExplain(t *testing.T) {
	for _, example := range explainExamples {
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

		actual := expr.Explain()

		if strings.Index(example.expected, "N") != -1 {
			re := MakeRegexp(example.expected)
			if !re.MatchString(actual) {
				t.Errorf("Explaining '%s' failed: %s does not match %s (%s)", example.input, actual, example.expected, re)
			}
		} else if actual != example.expected {
			t.Errorf("Explaining '%s' failed: %s does not match %s", example.input, actual, example.expected)
		}
	}
}
