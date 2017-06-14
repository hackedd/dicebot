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

var bot *dicebot.Bot

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Print("Received ready event")
	s.UpdateStatus(0, "")
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

	msg := strings.Replace(m.ContentWithMentionsReplaced(), s.State.Ready.User.Username, "username", 1)

	log.Printf("Received message: %s", msg)

	response := bot.HandleMessage(msg)
	if response != "" {
		s.ChannelMessageSend(m.ChannelID, response)
	}
}

func run(context *cli.Context) error {
	token := context.String("token")
	if token == "" {
		return cli.NewExitError("Authentication token is required.", 1)
	}

	var err error
	bot, err = dicebot.NewBot(context.String("database"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to open database: %s", err), 1)
	}

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to create Discord session: %s", err), 1)
	}

	switch context.String("log-level") {
	case "error":
		discord.LogLevel = discordgo.LogError
		break
	case "warning":
		discord.LogLevel = discordgo.LogWarning
		break
	case "info":
		discord.LogLevel = discordgo.LogInformational
		break
	case "debug":
		discord.LogLevel = discordgo.LogDebug
		break
	default:
		return cli.NewExitError(fmt.Sprintf("Unknown log level '%s'", context.String("log-level")), 1)
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
		cli.StringFlag{
			Name:  "log-level",
			Usage: "Log level (error, warning, info or debug)",
			Value: "error",
		},
		cli.StringFlag{
			Name:  "database",
			Usage: "Database filename",
			Value: "dicebot.db",
		},
	}

	app.Action = run
	app.Run(os.Args)
}
