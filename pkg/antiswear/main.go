package antiswear

import (
	"bufio"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/config"
	"github.com/gperis/forza-bot/pkg/database"
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
	initDb()
	loadDictionary()
}

func initDb() {
	db := database.OpenDatabase()
	stmt := "CREATE TABLE IF NOT EXISTS antiswear_count (`UserID` integer not null primary key, `Count` integer);"

	_, err := db.Exec(stmt)
	if err != nil {
		log.Printf("%q: %s\n", err, stmt)
		return
	}
	defer db.Close()
}

func StartModule(dg *discordgo.Session) {
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

				sendWarningToUser(s, m)
				logWarning(s, m)
			}
		}
	}
}

func logWarning(s *discordgo.Session, m *discordgo.MessageCreate) {
	db := database.OpenDatabase()

	tx, err := db.Begin()

	if err != nil {
		log.Printf("%q: %s\n", err)
	}

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO antiswear_count(`UserID`, `Count`) values(?, ?)")

	if err != nil {
		log.Printf("%q: %s\n", err, stmt)
	}

	defer stmt.Close()

	userCount := getUserWarningCount(m.Author.ID) + 1

	_, err = stmt.Exec(m.Author.ID, userCount)

	if err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	defer db.Close()

	logMessageEmbed := &discordgo.MessageEmbed{
		Title: "Auto Moderation | Anti Swearing",
		Description: fmt.Sprintf(
			"<@!%s> (%s) sent a message containing swear word(s) in <#%s>.\n\n**The message:**\n>>> %s",
			m.Author.ID,
			m.Author.ID,
			m.ChannelID,
			m.Content,
		),
		Color: 10038562,
	}

	// Send notification to log channel
	s.ChannelMessageSendEmbed(moduleConf.LogChannelID, logMessageEmbed)
}

func sendWarningToUser(s *discordgo.Session, m *discordgo.MessageCreate) {
	privateChannel, err := s.UserChannelCreate(m.Author.ID)

	if err != nil {
		log.Printf("There was an error trying to DM user %s: %v", m.Author.Username, err)
	}

	privateMessageEmbed := &discordgo.MessageEmbed{
		Title:       "Auto Moderation",
		Description: getWarningMessageForUser(m.Author.ID),
		Color:       10038562,
	}

	s.ChannelMessageSendEmbed(privateChannel.ID, privateMessageEmbed)
}

func getWarningMessageForUser(id string) string {
	var messageToSend string

	userCount := getUserWarningCount(id)

	for _, warning := range moduleConf.Warnings {
		if userCount >= warning.Count {
			messageToSend = warning.Message
		}
	}

	return messageToSend
}

func getUserWarningCount(id string) int {
	var userCount int

	db := database.OpenDatabase()

	stmt, err := db.Prepare("SELECT `Count` FROM antiswear_count WHERE `UserID` = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&userCount)

	if err != nil {
		userCount = 0
	}

	return userCount
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
