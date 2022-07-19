package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hackedd/dicebot"
	"github.com/urfave/cli/v2"
)

var bot *dicebot.Bot

func logMessage(s *discordgo.Session, level int, format string, args ...interface{}) {
	if s.LogLevel >= level {
		log.Printf(format, args...)
	}
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	logMessage(s, discordgo.LogDebug, "Received ready event: %+v", event)
	logMessage(s, discordgo.LogInformational, "Logged in as %s", s.State.User)

	if err := s.UpdateGameStatus(0, ""); err != nil {
		logMessage(s, discordgo.LogError, "Unable to set status: %v", err)
	}
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMessage(s, m.Message)
}

func onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	handleMessage(s, m.Message)
}

func handleMessage(s *discordgo.Session, m *discordgo.Message) {
	logMessage(s, discordgo.LogDebug, "Received message event: %+v", m)

	if m.Author == nil || s.State == nil || s.State.User == nil || m.Author.ID == s.State.User.ID {
		return
	}

	msg := strings.Replace(m.ContentWithMentionsReplaced(), s.State.User.Username, "username", 1)

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		logMessage(s, discordgo.LogError, "Unable to retrieve channel info for %s: %s", m.ChannelID, err)
		channel = &discordgo.Channel{GuildID: "unknown"}
	}

	context := dicebot.MessageContext{
		UserId:    m.Author.ID,
		UserName:  m.Author.Username,
		ChannelId: m.ChannelID,
		ServerId:  channel.GuildID,
	}

	response := bot.HandleMessage(context, msg)
	if response != "" {
		if len(response) < 2000 {
			_, err = s.ChannelMessageSend(m.ChannelID, response)
		} else {
			_, err = s.ChannelMessageSend(m.ChannelID, "Sorry, the result of your command is too long. Try rolling fewer dice.")
		}
		if err != nil {
			logMessage(s, discordgo.LogError, "Unable to send message to %s: %s", m.ChannelID, err)
		}
	}
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	logMessage(s, discordgo.LogDebug, "Received guild event: %+v", event.Guild)
}

var commands = []discordgo.ApplicationCommand{
	{
		Name:        "roll",
		Description: "Roll some dice",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "dice",
				Description: "What dice to roll (2d6, d20+2)",
				Required:    true,
			},
		},
	},
	{
		Name:        "save",
		Description: "Save a dice roll",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "dice",
				Description: "What dice to roll (2d6, d20+2)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "The name for this roll",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "for",
				Description: "Who should be able to use this roll",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "just me", Value: "user"},
					{Name: "this channel", Value: "channel"},
					{Name: "whole server", Value: "server"},
				},
			},
		},
	},
}

func GetUser(i *discordgo.Interaction) *discordgo.User {
	if i.User != nil {
		return i.User
	}
	if i.Member != nil {
		return i.Member.User
	}
	return nil
}

func onInteractionCreate(s *discordgo.Session, event *discordgo.InteractionCreate) {
	logMessage(s, discordgo.LogDebug, "Received interaction: %+v", event.Interaction)

	channel, err := s.State.Channel(event.ChannelID)
	if err != nil {
		logMessage(s, discordgo.LogError, "Unable to retrieve channel info for %s: %s", event.ChannelID, err)
		channel = &discordgo.Channel{GuildID: "unknown"}
	}

	user := GetUser(event.Interaction)

	context := dicebot.MessageContext{
		UserId:    user.ID,
		UserName:  user.Username,
		ChannelId: event.ChannelID,
		ServerId:  channel.GuildID,
	}

	commandData := event.ApplicationCommandData()
	options := make(map[string]string, len(commandData.Options))
	for _, option := range commandData.Options {
		options[option.Name] = option.StringValue()
	}

	var response string

	if commandData.Name == "roll" {
		dice := options["dice"]
		response = bot.RollDice(context, dice)
	} else if commandData.Name == "save" {
		dice, name, for_ := options["dice"], options["name"], options["for"]
		if err := bot.Save(context, dice, name, for_); err != nil {
			response = bot.HandleError(fmt.Sprintf("save %v as %v", dice, name), err)
		} else {
			response = fmt.Sprintf("Saved **%s** as `%s`", dice, name)
		}
	} else {
		logMessage(s, discordgo.LogError, "Unknown interaction command %v", commandData.Name)
		return
	}

	err = s.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		logMessage(s, discordgo.LogError, "Unable to send interaction response: %s", err)
	}
}

func run(context *cli.Context) error {
	token := context.String("token")
	if token == "" {
		return cli.Exit("Authentication token is required.", 1)
	}

	var err error
	bot, err = dicebot.NewBot(context.String("database"))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Unable to open database: %s", err), 1)
	}

	for _, filename := range context.StringSlice("moves") {
		err = bot.LoadMoves(filename)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Unable to load moves from %s: %s", filename, err), 1)
		}
	}

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Unable to create Discord session: %s", err), 1)
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
		return cli.Exit(fmt.Sprintf("Unknown log level '%s'", context.String("log-level")), 1)
	}

	discord.ShardID = context.Int("shard")
	discord.ShardCount = context.Int("num-shards")

	discord.AddHandler(onReady)
	discord.AddHandler(onMessageCreate)
	discord.AddHandler(onMessageUpdate)
	discord.AddHandler(onGuildCreate)
	discord.AddHandler(onInteractionCreate)

	if err := discord.Open(); err != nil {
		return cli.Exit(fmt.Sprintf("Unable to connect to Discord: %s", err), 1)
	}

	for _, command := range commands {
		if _, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", &command); err != nil {
			return cli.Exit(fmt.Sprintf("Unable to create command %v", command.Name), 1)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	if err = discord.Close(); err != nil {
		logMessage(discord, discordgo.LogError, "Unable to close connection: %s", err)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "dicebot"
	app.Usage = "a dice rolling bot for Discord"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "token",
			Usage: "Discord authentication token",
		},
		&cli.IntFlag{
			Name:  "shard",
			Usage: "Shard ID",
		},
		&cli.IntFlag{
			Name:  "num-shards",
			Usage: "Number of shards",
			Value: 1,
		},
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "Log level (error, warning, info or debug)",
			Value: "error",
		},
		&cli.StringFlag{
			Name:  "database",
			Usage: "Database filename",
			Value: "dicebot.json",
		},
		&cli.StringSliceFlag{
			Name:  "moves",
			Usage: "Load moves from file",
		},
	}

	app.Action = run
	app.Run(os.Args)
}
