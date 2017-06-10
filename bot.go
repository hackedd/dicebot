package dicebot

import (
	"fmt"
	"strings"
)

func EscapeMarkdown(input string) string {
	for _, s := range []string{"*", "~", "_", "`"} {
		input = strings.Replace(input, s, "\\"+s, -1)
	}
	return input
}

type Bot struct {
}

func (bot *Bot) Usage() string {
	return "Type `!roll d<x>` to roll a *x*-sided die\n" +
		"Type `!roll <n>d<x>` to roll any number of *x*-sided dice (`!roll 3d6` rolls three regular six-sided dice)\n" +
		"You can use simple mathematical expressions too. For example, `d20 + 4` rolls a twenty-sided dice and adds four to the result.\n" +
		"The bot understands addition, subtraction, multiplication, division and brackets."
}

func (bot *Bot) RollDice(input string) string {
	tokens, err := Tokenize(input)
	if err != nil {
		return bot.HandleError(input, err)
	}

	expr, err := Parse(tokens)
	if err != nil {
		return bot.HandleError(input, err)
	}

	value := fmt.Sprintf("%d", expr.Eval())
	explanation := expr.Explain()

	s := EscapeMarkdown(input) + " => "
	if input != explanation && value != explanation {
		s += "**" + EscapeMarkdown(explanation) + "** => "
	}
	s += "**" + EscapeMarkdown(value) + "**"
	return s
}

func (bot *Bot) HandleError(command string, err error) string {
	s := fmt.Sprintf("Sorry, I don't understand how to parse '%s'\n", EscapeMarkdown(command))

	if parseError, ok := err.(ParseError); ok {
		s += fmt.Sprintf("```\n%s\n%s^-- %s\n```", command, strings.Repeat(" ", parseError.Position), parseError.Message)
	} else {
		s += err.Error()
	}

	return s
}

func (bot *Bot) HandleMessage(msg string) string {
	idx := strings.Index(msg, "!roll")
	if idx != 0 {
		return ""
	}

	command := strings.TrimSpace(msg[idx+5:])
	if command == "" || command == "help" {
		return bot.Usage()
	}

	return bot.RollDice(command)
}
