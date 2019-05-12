package dicebot

import (
	"errors"
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
	db    Database
	moves map[string]Move
}

type MessageContext struct {
	UserId    string
	UserName  string
	ChannelId string
	ServerId  string
}

func NewBot(dbFile string) (*Bot, error) {
	db, err := NewJsonDatabase(dbFile)
	if err != nil {
		return nil, err
	}

	return &Bot{db: db, moves: make(map[string]Move)}, nil
}

func (bot *Bot) LoadMoves(filename string) error {
	return LoadMoves(bot.moves, filename)
}

func (bot *Bot) Usage() string {
	return "Type `!roll d<x>` to roll a *x*-sided die\n" +
		"Type `!roll <n>d<x>` to roll any number of *x*-sided dice (`!roll 3d6` rolls three regular six-sided dice)\n" +
		"You can use simple mathematical expressions too. For example, `d20 + 4` rolls a twenty-sided dice and adds four to the result.\n" +
		"The bot understands addition, subtraction, multiplication, division and brackets.\n" +
		"Type `!save <expr> as <name>` to save an expression. For example you could `!save 2d6+1 as str` and use `!roll str` later.\n" +
		"Type `!move` to get a list of moves, and `!move <name>` to make a move."
}

func (bot *Bot) LookupVariable(context MessageContext, name string) (Expr, error) {
	for _, scope := range []string{"user-" + context.UserId, "channel-" + context.ChannelId, "server-" + context.ServerId} {
		value, found := bot.db.ReadValue(name, scope)
		if found {
			return ParseString(value)
		}
	}

	return nil, errors.New(fmt.Sprintf("Undefined variable `%s`", name))
}

func (bot *Bot) Eval(context MessageContext, input string) (value int, explanation string, err error) {
	expr, err := ParseString(input)
	if err != nil {
		return
	}

	lookup := func(name string) (Expr, error) {
		return bot.LookupVariable(context, name)
	}

	value, err = expr.Eval(lookup)
	if err != nil {
		return
	}

	explanation = expr.Explain(lookup)
	return
}

func (bot *Bot) FormatResult(input string, value int, explanation string) string {
	result := fmt.Sprintf("%d", value)

	s := EscapeMarkdown(input) + " => "
	if input != explanation && result != explanation {
		s += "**" + EscapeMarkdown(explanation) + "** => "
	}
	s += "**" + EscapeMarkdown(result) + "**"
	return s
}

func (bot *Bot) RollDice(context MessageContext, input string) string {
	value, explanation, err := bot.Eval(context, input)

	if err != nil {
		return bot.HandleError(input, err)
	}

	return bot.FormatResult(input, value, explanation)
}

func (bot *Bot) Save(context MessageContext, input, name, for_ string) error {
	_, err := ParseString(input)
	if err != nil {
		return err
	}

	scope := ""
	switch for_ {
	case "server":
		scope = "server-" + context.ServerId
	case "channel":
		scope = "channel-" + context.ChannelId
	case "", "me", "user":
		scope = "user-" + context.UserId
	default:
		return errors.New("undefined scope " + for_)
	}

	return bot.db.StoreValue(name, scope, input)
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

func (bot *Bot) HandleMessage(context MessageContext, msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "!roll" || msg == "!roll help" || msg == "!save" || msg == "!save help" {
		return bot.Usage()
	}

	if strings.Index(msg, "!roll ") == 0 {
		return bot.RollDice(context, strings.TrimSpace(msg[6:]))
	}

	if strings.Index(msg, "!save ") == 0 {
		r := regexp.MustCompile(`\A!save\s+(.*)\s+as\s+(\w+)(\s+for\s+(\w+))?\z`)
		match := r.FindStringSubmatch(msg)
		if match == nil {
			return bot.HandleError(msg[1:], nil)
		}
		err := bot.Save(context, match[1], match[2], match[4])
		if err != nil {
			return bot.HandleError(msg[1:], err)
		}
		return fmt.Sprintf("Saved **%s** as `%s`", match[1], match[2])
	}

	if msg == "!move" {
		response := "I know the following moves:\n"
		for _, move := range bot.moves {
			response += " * " + EscapeMarkdown(move.Name) + "\n"
		}
		return response
	}

	if strings.Index(msg, "!move ") == 0 {
		return bot.MakeMove(context, strings.TrimSpace(msg[6:]))
	}

	return ""
}
