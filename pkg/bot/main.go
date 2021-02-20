package bot

import (
	"fmt"
	"github.com/gperis/forza-bot/pkg/antimention"
	"github.com/gperis/forza-bot/pkg/antispam"
	"github.com/gperis/forza-bot/pkg/antiswear"
	"github.com/gperis/forza-bot/pkg/commands"
	"github.com/gperis/forza-bot/pkg/config"
	"github.com/gperis/forza-bot/pkg/invitation_link"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type conf struct {
	Token string `mapstructure:"token"`
}

var moduleConf conf

func init() {
	config.Load("bot", &moduleConf)
}

func Start() {
	dg, done := authenticate()
	if done {
		return
	}

	startModules(dg)

	fmt.Println("Forza Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func authenticate() (*discordgo.Session, bool) {
	if moduleConf.Token == "" {
		fmt.Println("No token provided. Please add one to the config file.")
		return nil, true
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + moduleConf.Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return nil, true
	}

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	return dg, false
}

func pingHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

func startModules(dg *discordgo.Session) {
	dg.AddHandler(pingHandler)

	antiswear.StartModule(dg)
	invitation_link.StartModule(dg)
	antispam.StartModule(dg)
	antimention.StartModule(dg)

	commands.StartModule(dg)
}
