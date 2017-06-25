package dicebot

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func EscapeMarkdown(input string) string {
	for _, s := range []string{"*", "~", "_", "`"} {
		input = strings.Replace(input, s, "\\"+s, -1)
	}
	return input
}

type Bot struct {
	db *bolt.DB
}

func NewBot(dbFile string) (*Bot, error) {
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &Bot{db: db}, nil
}

func (bot *Bot) Usage() string {
	return "Type `!roll d<x>` to roll a *x*-sided die\n" +
		"Type `!roll <n>d<x>` to roll any number of *x*-sided dice (`!roll 3d6` rolls three regular six-sided dice)\n" +
		"You can use simple mathematical expressions too. For example, `d20 + 4` rolls a twenty-sided dice and adds four to the result.\n" +
		"The bot understands addition, subtraction, multiplication, division and brackets."
}

func (bot *Bot) LookupVariable(name, userId, channelId, serverId string) (Expr, error) {
	var v []byte

	err := bot.db.View(func(tx *bolt.Tx) error {
		for _, bucketName := range []string{"user-" + userId, "channel-" + channelId, "server-" + serverId} {
			b := tx.Bucket([]byte(bucketName))
			if b != nil {
				v = b.Get([]byte(name))
				if v != nil {
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, errors.New(fmt.Sprintf("Undefined variable `%s`", name))
	}

	expr, err := ParseString(string(v))
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (bot *Bot) RollDice(input, userId, channelId, serverId string) string {
	expr, err := ParseString(input)
	if err != nil {
		return bot.HandleError(input, err)
	}

	lookup := func(name string) (Expr, error) {
		return bot.LookupVariable(name, userId, channelId, serverId)
	}

	v, err := expr.Eval(lookup)
	if err != nil {
		return bot.HandleError(input, err)
	}

	value := fmt.Sprintf("%d", v)
	explanation := expr.Explain(lookup)

	s := EscapeMarkdown(input) + " => "
	if input != explanation && value != explanation {
		s += "**" + EscapeMarkdown(explanation) + "** => "
	}
	s += "**" + EscapeMarkdown(value) + "**"
	return s
}

func (bot *Bot) Save(input, name, scope string) string {
	_, err := ParseString(input)
	if err != nil {
		return bot.HandleError(input, err)
	}

	err = bot.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(scope))
		if err != nil {
			return err
		}
		return b.Put([]byte(name), []byte(input))
	})

	if err != nil {
		return bot.HandleError(input, err)
	}

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

func (bot *Bot) HandleMessage(msg, serverId, channelId, userId string) string {
	idx := strings.Index(msg, "!roll")
	if idx != 0 {
		return ""
	}

	command := strings.TrimSpace(msg[idx+5:])
	if command == "" || command == "help" {
		return bot.Usage()
	}

	if strings.Index(command, "save") == 0 {
		r := regexp.MustCompile(`\Asave\s+(.*)\s+as\s+(\w+)(\s+for\s+(\w+))?\z`)
		match := r.FindStringSubmatch(command)
		if match == nil {
			return bot.HandleError(command, nil)
		}

		scope := ""
		switch match[4] {
		case "server":
			scope = "server-" + serverId
		case "channel":
			scope = "channel-" + channelId
		case "", "me", "user":
			scope = "user-" + userId
		default:
			return bot.HandleError(command, errors.New("Undefined scope "+match[4]))
		}

		return bot.Save(match[1], match[2], scope)
	}

	return bot.RollDice(command, userId, channelId, serverId)
}
