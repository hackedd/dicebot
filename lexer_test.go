package dicebot

import "testing"

func checkTokenizer(t *testing.T, expression string, expected []Token) {
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
		if expected[i] == actual[i] {
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
	checkTokenizer(t, "3d6", []Token{{DICE, "3d6", 0}, {END, "", 3}})
	checkTokenizer(t, "3d6 + 2", []Token{{DICE, "3d6", 0}, {PLUS, "+", 4}, {NUMBER, "2", 6}, {END, "", 7}})
	checkTokenizer(t, "4 * (3d6+2)", []Token{{NUMBER, "4", 0}, {MULTIPLY, "*", 2}, {LEFT_PAREN, "(", 4}, {DICE, "3d6", 5}, {PLUS, "+", 8}, {NUMBER, "2", 9}, {RIGHT_PAREN, ")", 10}, {END, "", 11}})

	if _, err := Tokenize("foo"); err == nil {
		t.Error("Unexpected success parsing 'foo'")
	} else if err.Error() != "Unrecognized character 'f' near position 0" {
		t.Errorf("Unexpected error parsing 'foo', got %+v", err)
	}

	if _, err := Tokenize("dd6"); err == nil {
		t.Error("Unexpected success parsing 'dd6'")
	} else if err.Error() != "Unexpected 'd' in 'dd6' near position 0" {
		t.Errorf("Unexpected error parsing 'dd6', got %+v", err)
	}
}