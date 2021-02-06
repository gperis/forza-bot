package bot

import (
	"fmt"
	"github.com/gperis/forza-bot/pkg/config"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/antiswear"
)

type conf struct {
	Token string `mapstructure:"token"`
}

var moduleConf conf

func init() {
	config.Load("bot", &moduleConf)
}

// Start the bot
func Start() {
	if moduleConf.Token == "" {
		fmt.Println("No token provided. Please add one to the config file.")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + moduleConf.Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	initialiseModules(dg)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Forza Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

func initialiseModules(dg *discordgo.Session) {
	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	antiswear.InitialiseModule(dg)
}
