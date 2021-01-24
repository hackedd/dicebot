package dicebot

import (
	"fmt"
	"regexp"
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
	IDENTIFIER
	BEST_OF
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

type tokenPattern struct {
	Type    TokenType
	Pattern *regexp.Regexp
}

// Dice is before identifier, so that things like 'd6' are parsed as a dice, not identifier.
var patterns = []tokenPattern{
	{NUMBER, regexp.MustCompile(`\d+`)},
	{DICE, regexp.MustCompile(`(?i)(\d*)d(\d+)`)},
	{BEST_OF, regexp.MustCompile(`(?i)best\s+(\d+\s+)?of\s+(\d*)d(\d+)`)},
	{IDENTIFIER, regexp.MustCompile(`(?i)[a-z_][a-z0-9_]*`)},
}

type ParseError struct {
	Message  string
	Position int
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s near position %d", e.Message, e.Position)
}

func longestMatch(input string) (tokenType TokenType, match string) {
	tokenType = END
	match = ""

	for _, p := range patterns {
		if loc := p.Pattern.FindStringIndex(input); loc != nil && loc[0] == 0 && loc[1] > len(match) {
			tokenType = p.Type
			match = input[:loc[1]]
		}
	}

	return
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

		tokenType, match := longestMatch(string(runes[i:]))
		if tokenType == END {
			return nil, ParseError{"Input not matched", i}
		}

		tokens = append(tokens, Token{tokenType, match, i})
		i += len(match) - 1
	}

	tokens = append(tokens, Token{END, "", len(runes)})

	return tokens, nil
}
