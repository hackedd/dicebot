package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hackedd/dicebot"
	"github.com/urfave/cli"
)

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Print("Received ready event")
	s.UpdateStatus(0, "")
}

func showUsage(s *discordgo.Session, channelID string) {
	s.ChannelMessageSend(channelID, "Type `!roll d<x>` to roll a *x*-sided die\n"+
		"Type `!roll <n>d<x>` to roll any number of *x*-sided dice (`!roll 3d6` rolls three regular six-sided dice)\n"+
		"You can use simple mathematical expressions too. For example, `d20 + 4` rolls a twenty-sided dice and adds four to the result.\n"+
		"The bot understands addition, subtraction, multiplication, division and brackets.")
}

func escapeMarkdown(input string) string {
	for _, s := range []string{"*", "~", "_", "`"} {
		input = strings.Replace(input, s, "\\"+s, -1)
	}
	return input
}

func rollDice(s *discordgo.Session, channelID, input string) error {
	tokens, err := dicebot.Tokenize(input)
	if err != nil {
		return err
	}

	expr, err := dicebot.Parse(tokens)
	if err != nil {
		return err
	}

	value := fmt.Sprintf("%d", expr.Eval())
	explanation := expr.Explain()
	if input == explanation || value == explanation {
		s.ChannelMessageSend(channelID, fmt.Sprintf("%s => **%s**", escapeMarkdown(input), escapeMarkdown(value)))
	} else {
		s.ChannelMessageSend(channelID, fmt.Sprintf("%s => %s => **%s**", escapeMarkdown(input), escapeMarkdown(explanation), escapeMarkdown(value)))
	}

	return nil
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMessage(s, m.Message)
}

func onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	handleMessage(s, m.Message)
}

func handleMessage(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) <= 0 || m.Content[0] != '!' {
		return
	}

	msg := strings.Replace(m.ContentWithMentionsReplaced(), s.State.Ready.User.Username, "username", 1)

	log.Printf("Received message: %s", msg)

	idx := strings.Index(msg, "!roll")
	if idx < 0 {
		return
	}

	command := strings.TrimSpace(msg[idx+5:])
	if command == "" || command == "help" {
		showUsage(s, m.ChannelID)
	} else {
		err := rollDice(s, m.ChannelID, command)
		if err != nil {
			if parseError, ok := err.(dicebot.ParseError); ok {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, I don't understand how to parse '%s'\n", escapeMarkdown(command)))
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s\n%s^-- %s\n```", command, strings.Repeat(" ", parseError.Position), parseError.Message))
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, I don't understand how to parse '%s'\n%s", escapeMarkdown(command), err))
			}
		}
	}
}

func run(context *cli.Context) error {
	discord, err := discordgo.New("Bot " + context.String("token"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to create Discord session: %s", err), 1)
	}

	discord.ShardID = context.Int("shard")
	discord.ShardCount = context.Int("num-shards")

	discord.AddHandler(onReady)
	discord.AddHandler(onMessageCreate)
	discord.AddHandler(onMessageUpdate)

	if err := discord.Open(); err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to connect to Discord: %s", err), 1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	discord.Close()

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "dicebot"
	app.Usage = "a dice rolling bot for Discord"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "token",
			Usage: "Discord authentication token",
		},
		cli.IntFlag{
			Name:  "shard",
			Usage: "Shard ID",
		},
		cli.IntFlag{
			Name:  "num-shards",
			Usage: "Number of shards",
			Value: 1,
		},
	}

	app.Action = run
	app.Run(os.Args)
}
