package dicebot

import "testing"

type token struct {
	Type TokenType
	Text string
}

func checkTokenizer(t *testing.T, expression string, expected []token) {
	actual, err := Tokenize(expression)
	if err != nil {
		t.Errorf("Unexpected error parsing '%s': %s", expression, err)
		return
	}

	for len(actual) < len(expected) {
		actual = append(actual, Token{})
	}
	for len(expected) < len(actual) {
		actual = append(actual, Token{})
	}

	same := true
	for i := 0; i < len(expected); i += 1 {
		if expected[i].Type == actual[i].Type && expected[i].Text == actual[i].Text {
			t.Logf("%d: %+v %+v  ok", i, expected[i], actual[i])
		} else {
			t.Logf("%d: %+v %+v", i, expected[i], actual[i])
			same = false
		}
	}
	if !same {
		t.Errorf("Token mismatch parsing '%s'", expression)
	}
}

func TestTokenize(t *testing.T) {
	checkTokenizer(t, "3d6", []token{{DICE, "3d6"}, {END, ""}})
	checkTokenizer(t, "3d6 + 2", []token{{DICE, "3d6"}, {PLUS, "+"}, {NUMBER, "2"}, {END, ""}})
	checkTokenizer(t, "4 * (3d6+2)", []token{{NUMBER, "4"}, {MULTIPLY, "*"}, {LEFT_PAREN, "("}, {DICE, "3d6"}, {PLUS, "+"}, {NUMBER, "2"}, {RIGHT_PAREN, ")"}, {END, ""}})
	checkTokenizer(t, "var", []token{{IDENTIFIER, "var"}, {END, ""}})
	checkTokenizer(t, "VAR", []token{{IDENTIFIER, "VAR"}, {END, ""}})
	checkTokenizer(t, "d6", []token{{DICE, "d6"}, {END, ""}})
	checkTokenizer(t, "d", []token{{IDENTIFIER, "d"}, {END, ""}})
	checkTokenizer(t, "best of 2d6", []token{{BEST_OF, "best of 2d6"}, {END, ""}})
	checkTokenizer(t, "best 2 of 3d6", []token{{BEST_OF, "best 2 of 3d6"}, {END, ""}})

	if _, err := Tokenize("1.2"); err == nil {
		t.Error("Unexpected success parsing '1.2'")
	} else if err.Error() != "Input not matched near position 1" {
		t.Errorf("Unexpected error parsing '1.2', got %+v", err)
	}
}
