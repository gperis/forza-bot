package antiswear

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/admin"
	"github.com/gperis/forza-bot/pkg/config"
	"github.com/gperis/forza-bot/pkg/discord_log"
)

type conf struct {
	ListPath string `mapstructure:"list_path"`
}

var (
	dictionary []string
	moduleConf conf
)

func init() {
	config.Load("antiswear", &moduleConf)
	loadDictionary()
}

func StartModule(dg *discordgo.Session) {
	dg.AddHandler(handler)
}

func handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || admin.IsStaffMember(m.Member) || m.Author.Bot == true {
		return
	}

	contentSliced := strings.Fields(m.Content)
	for _, word := range contentSliced {
		for _, forbiddenWord := range dictionary {

			if strings.Index(strings.ToLower(word), forbiddenWord) > -1 {
				s.ChannelMessageDelete(m.ChannelID, m.ID)

				sendWarningToUser(s, m)
				logWarning(s, m)
			}
		}
	}
}

func logWarning(s *discordgo.Session, m *discordgo.MessageCreate) {
	discord_log.LogIncident(
		s,
		fmt.Sprintf(
			"<@!%s> (%s) sent a message containing swear word(s) in <#%s>.\n\n**The message:**\n>>> %s",
			m.Author.ID,
			m.Author.ID,
			m.ChannelID,
			m.Content,
		),
		"Anti Swearing",
	)
}

func sendWarningToUser(s *discordgo.Session, m *discordgo.MessageCreate) {
	privateChannel, err := s.UserChannelCreate(m.Author.ID)

	if err != nil {
		log.Printf("There was an error trying to DM user %s: %v", m.Author.Username, err)
	}

	privateMessageEmbed := &discordgo.MessageEmbed{
		Title:       "Auto Moderation",
		Description: "Please refrain from swearing as that's against the server rules. Repeating this action may lead to consequences.",
		Color:       12386317,
	}

	s.ChannelMessageSendEmbed(privateChannel.ID, privateMessageEmbed)
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
