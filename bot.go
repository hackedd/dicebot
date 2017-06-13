package dicebot

import (
	"fmt"
	"regexp"
	"strings"
)

func EscapeMarkdown(input string) string {
	for _, s := range []string{"*", "~", "_", "`"} {
		input = strings.Replace(input, s, "\\"+s, -1)
	}
	return input
}

type Bot struct {
	Vars map[string]Expr
}

func NewBot() *Bot {
	return &Bot{Vars: make(map[string]Expr)}
}

func (bot *Bot) Usage() string {
	return "Type `!roll d<x>` to roll a *x*-sided die\n" +
		"Type `!roll <n>d<x>` to roll any number of *x*-sided dice (`!roll 3d6` rolls three regular six-sided dice)\n" +
		"You can use simple mathematical expressions too. For example, `d20 + 4` rolls a twenty-sided dice and adds four to the result.\n" +
		"The bot understands addition, subtraction, multiplication, division and brackets."
}

func (bot *Bot) LookupVariable(name string) (Expr, bool) {
	expr, ok := bot.Vars[name]
	return expr, ok
}

func (bot *Bot) RollDice(input string) string {
	expr, err := ParseString(input)
	if err != nil {
		return bot.HandleError(input, err)
	}

	v, err := expr.Eval(bot.LookupVariable)
	if err != nil {
		return bot.HandleError(input, err)
	}

	value := fmt.Sprintf("%d", v)
	explanation := expr.Explain(bot.LookupVariable)

	s := EscapeMarkdown(input) + " => "
	if input != explanation && value != explanation {
		s += "**" + EscapeMarkdown(explanation) + "** => "
	}
	s += "**" + EscapeMarkdown(value) + "**"
	return s
}

func (bot *Bot) Save(input, name string) string {
	expr, err := ParseString(input)
	if err != nil {
		return bot.HandleError(input, err)
	}

	bot.Vars[name] = expr
	return fmt.Sprintf("Saved **%s** as `%s`", input, name)
}

func (bot *Bot) HandleError(command string, err error) string {
	s := fmt.Sprintf("Sorry, I don't understand how to parse '%s'", EscapeMarkdown(command))

	if err == nil {
		return s
	}

	if parseError, ok := err.(ParseError); ok {
		return s + fmt.Sprintf("\n```\n%s\n%s^-- %s\n```", command, strings.Repeat(" ", parseError.Position), parseError.Message)
	}

	return s + ": " + err.Error()
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

	if strings.Index(command, "save") == 0 {
		r := regexp.MustCompile(`\Asave\s+(.*)\s+as\s+(\w+)\z`)
		match := r.FindStringSubmatch(command)
		if match == nil {
			return bot.HandleError(command, nil)
		}

		return bot.Save(match[1], match[2])
	}

	return bot.RollDice(command)
}
