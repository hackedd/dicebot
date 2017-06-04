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

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	log.Printf("Received guild create event: %s (%s)", event.Guild.Name, event.Guild.ID)

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			s.ChannelMessageSend(channel.ID, "**Dice rolling bot ready for action. Type !roll to activate.**")
			return
		}
	}
}

func showUsage(s *discordgo.Session, channelID string) {
	s.ChannelMessageSend(channelID, "Type `!roll d<x>` to roll a *x*-sided die\n"+
		"Type `!roll <n>d<x>` to roll any number of *x*-sided dice (`!roll 3d6` rolls three regular six-sided dice)\n"+
		"You can use simple mathematical expressions too. For example, `d20 + 4` rolls a twenty-sided dice and adds four to the result.\n"+
		"The bot understands addition, subtraction, multiplication, division and brackets.")
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

	// TODO: Add information about individual rolls in the result
	s.ChannelMessageSend(channelID, fmt.Sprintf("%s => **%d**", input, expr.Eval()))

	return nil
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, I don't understand how to parse '%s'\n%s", command, err))
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
	discord.AddHandler(onGuildCreate)
	discord.AddHandler(onMessageCreate)

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