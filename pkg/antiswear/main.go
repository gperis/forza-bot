package antiswear

import (
	"bufio"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/config"
	"log"
	"os"
	"strings"
)

type conf struct {
	ListPath     string `mapstructure:"list_path"`
	LogChannelID string `mapstructure:"log_channel_id"`
	Warnings     []struct {
		Count   int    `mapstructure:"count"`
		Message string `mapstructure:"message"`
	} `mapstructure:"warnings"`
}

var (
	dictionary []string
	moduleConf conf
)

func init() {
	config.Load("antiswear", &moduleConf)
	loadDictionary()
}

func InitialiseModule(dg *discordgo.Session) {
	dg.AddHandler(antiSwearHandler)
}

func antiSwearHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	contentSliced := strings.Fields(m.Content)
	for _, word := range contentSliced {
		for _, forbiddenWord := range dictionary {

			// NOTE : change between Index and EqualFold to see the different result

			if test := strings.Index(strings.ToLower(word), forbiddenWord); test > -1 {
				// Delete the swear message
				s.ChannelMessageDelete(m.ChannelID, m.ID)

				logMessageEmbed := &discordgo.MessageEmbed{
					Title: "Auto Moderation | Anti Swearing",
					Description: fmt.Sprintf(
						"<@!%s> (%s) sent a message containing swear word(s) in <#%s>.\n\nThe message:\n>>> %s",
						m.Author.ID,
						m.Author.ID,
						m.ChannelID,
						m.Content,
					),
				}

				// Send notification to log channel
				s.ChannelMessageSendEmbed(moduleConf.LogChannelID, logMessageEmbed)
			}
		}
	}
}

func loadDictionary() {
	d, err := os.Open(moduleConf.ListPath)

	if err != nil {
		log.Fatalf("Please double check that the profanity list is in the configured folder.")
	}
	defer d.Close()

	scanner := bufio.NewScanner(d)
	for scanner.Scan() {
		dictionary = append(dictionary, scanner.Text())
	}
}
