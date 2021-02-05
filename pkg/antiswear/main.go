package antiswear

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

type antiSwearConfig struct {
	ListPath     string `yaml:"list_path"`
	LogChannelID int    `yaml:"log_channel_id"`
	Warnings     []struct {
		Count   int    `yaml:"count"`
		Message string `yaml:"message"`
	} `yaml:"warnings"`
}

func (BotModule) LoadModule() {
	loadDictionary()
}

func loadDictionary() {
	var config antiSwearConfig

	err = yaml.Unmarshal("./config/antiswear.yaml", &config)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(config.ListPath, os.O_O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println("Please double check that the profanity list is in the configured folder.")
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		_ = sc.Text() // GET the line string
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
		return
	}
}

func startAntiSwear(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	contentSliced := strings.Fields(m.Content)
	for _, word := range contentSliced {
		for _, forbiddenWord := range swears {

			// NOTE : change between Index and EqualFold to see the different result

			if test := strings.Index(strings.ToLower(word), forbiddenWord); test > -1 {
				s.ChannelMessageDelete(m.ChannelID, m.ID)
				s.ChannelMessageSend(m.ChannelID, "Message deleted")
			}
		}
	}
}
