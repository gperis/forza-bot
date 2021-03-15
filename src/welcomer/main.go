package welcomer

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/src/config"
)

type conf struct {
	CommunityChannelID string `mapstructure:"community_channel_id"`
	InfoChannelID      string `mapstructure:"info_channel_id"`
	RulesChannelID     string `mapstructure:"rules_channel_id"`
	ImageURL           string `mapstructure:"image_url"`
	NewJoinerRole      string `mapstructure:"new_joiner_role"`
}

var moduleConf conf

func init() {
	config.Load("welcomer", &moduleConf)
}

func StartModule(dg *discordgo.Session) {
	dg.AddHandler(handler)
}

func handler(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	s.ChannelMessageSend(moduleConf.CommunityChannelID, fmt.Sprintf("<@%s>", m.User.ID))
	s.ChannelMessageSendEmbed(
		moduleConf.CommunityChannelID,
		&discordgo.MessageEmbed{
			Description: fmt.Sprintf(
				`Hey there :wave:

				Welcome to Forza Trucking's discord server. We're truly glad to have you here with us! If you're wondering what Forza is all about please have a look at the <#%s> channel.
				
				If you've applied to join the VTC please be patient, someone from our HR Department will get in touch with you as soon as possible. But, if you're looking to apply, please type %s.
				
				<:__:820050335979929632> Please make sure you go through the rules in <#%s>. Enjoy your stay!`,
				moduleConf.InfoChannelID,
				"`/apply`",
				moduleConf.RulesChannelID,
			),
			Color:  12386317,
			Image:  &discordgo.MessageEmbedImage{URL: moduleConf.ImageURL},
			Footer: &discordgo.MessageEmbedFooter{Text: "Forza Trucking - Certainly the Finest"},
		},
	)
}
