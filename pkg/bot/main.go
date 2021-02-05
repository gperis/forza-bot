package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var buffer = make([][]byte, 0)

// Start the bot
func Start() {
	if token == "" {
		fmt.Println("No token provided. Please run: main -t <bot token>")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
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

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

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
	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(AntiSwear)
}
