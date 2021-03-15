package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/src/config"
)

type conf struct {
	ConvoyConfig struct {
		OrderTrigger     string   `mapstructure:"order_trigger"`
		FinishTrigger    string   `mapstructure:"finish_trigger"`
		MessageChannelID string   `mapstructure:"message_channel_id"`
		WhitelistRoles   []string `mapstructure:"whitelist_roles"`
	} `mapstructure:"convoy"`
	SuggestionConfig struct {
		ServerTrigger          string   `mapstructure:"server_trigger"`
		VtcTrigger             string   `mapstructure:"vtc_trigger"`
		AcceptTrigger          string   `mapstructure:"accept_trigger"`
		RejectTrigger          string   `mapstructure:"reject_trigger"`
		ServerMessageChannelID string   `mapstructure:"server_message_channel_id"`
		VtcMessageChannelID    string   `mapstructure:"vtc_message_channel_id"`
		AcceptedEmoticonID     string   `mapstructure:"accepted_emoticon_id"`
		RejectedEmoticonID     string   `mapstructure:"rejected_emoticon_id"`
		WaitingEmoticonID      string   `mapstructure:"waiting_emoticon_id"`
		WhitelistRoles         []string `mapstructure:"whitelist_roles"`
	} `mapstructure:"server_suggestion"`
}

var (
	commandsConfig conf
)

func init() {
	config.Load("commands", &commandsConfig)
}

func StartModule(dg *discordgo.Session) {
	// Convoy
	dg.AddHandler(orderHandler)
	dg.AddHandler(finishHandler)

	// Server Suggestion
	dg.AddHandler(suggestionHandler)
}

func isAllowedToUse(member *discordgo.Member, whitelistedRoles []string) bool {
	if member != nil && member.Roles != nil {
		for _, role := range whitelistedRoles {
			for _, memberRole := range member.Roles {
				if role == memberRole {
					return true
				}
			}
		}
	}

	return false
}
