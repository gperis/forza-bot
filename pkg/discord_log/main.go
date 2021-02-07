package discord_log

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/config"
)

type conf struct {
	LogChannelID string `mapstructure:"log_channel_id"`
}

var moduleConf conf

func init() {
	config.Load("logs", &moduleConf)
}

func LogIncident(s *discordgo.Session, message string, moduleName string) {
	logMessageEmbed := &discordgo.MessageEmbed{
		Title:       "Auto Moderation | " + moduleName,
		Description: message,
		Color:       10038562,
	}

	s.ChannelMessageSendEmbed(moduleConf.LogChannelID, logMessageEmbed)
}
