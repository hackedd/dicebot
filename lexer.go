package dicebot

import (
	"fmt"
	"unicode"
)

type TokenType int

const (
	LEFT_PAREN TokenType = iota
	RIGHT_PAREN
	PLUS
	MINUS
	MULTIPLY
	DIVIDE
	NUMBER
	DICE
	END
)

type Token struct {
	Type     TokenType
	Text     string
	Position int
}

var operators = map[rune]TokenType{
	'(': LEFT_PAREN,
	')': RIGHT_PAREN,
	'+': PLUS,
	'-': MINUS,
	'*': MULTIPLY,
	'/': DIVIDE,
}

type ParseError struct {
	Message  string
	Position int
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s near position %d", e.Message, e.Position)
}

func Tokenize(expression string) ([]Token, error) {
	tokens := make([]Token, 0)

	runes := []rune(expression)
	for i := 0; i < len(runes); i += 1 {
		if unicode.IsSpace(runes[i]) {
			continue
		}

		if op, ok := operators[runes[i]]; ok {
			tokens = append(tokens, Token{op, string(runes[i]), i})
			continue
		}

		s := i
		d := 0
		for i < len(runes) && (runes[i] >= '0' && runes[i] <= '9' || runes[i] == 'd') {
			if runes[i] == 'd' {
				d += 1
			}
			i += 1
		}

		if s == i {
			return nil, ParseError{fmt.Sprintf("Unrecognized character '%c'", runes[s]), s}
		}

		value := string(runes[s:i])
		if d == 0 {
			tokens = append(tokens, Token{NUMBER, value, s})
		} else if d == 1 {
			tokens = append(tokens, Token{DICE, value, s})
		} else {
			return nil, ParseError{fmt.Sprintf("Unexpected 'd' in '%s'", value), s}
		}

		// Prevent skipping the next character.
		i -= 1
	}

	tokens = append(tokens, Token{END, "", len(runes)})

	return tokens, nil
}
