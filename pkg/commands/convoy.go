package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/config"
	"strings"
)

type conf struct {
	ConvoyConfig struct {
		OrderTrigger     string   `mapstructure:"order_trigger"`
		FinishTrigger    string   `mapstructure:"finish_trigger"`
		MessageChannelID string   `mapstructure:"message_channel_id"`
		WhitelistRoles   []string `mapstructure:"whitelist_roles"`
	} `mapstructure:"convoy"`
}

var (
	commandsConfig conf
)

func init() {
	config.Load("commands", &commandsConfig)
}

func StartModule(dg *discordgo.Session) {
	dg.AddHandler(orderHandler)
	dg.AddHandler(finishHandler)
}

func orderHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		!isAllowedToUse(m.Member) ||
		!strings.Contains(m.Content, commandsConfig.ConvoyConfig.OrderTrigger) {
		return
	}

	messageEmbed := &discordgo.MessageEmbed{
		Color: 12386317,
		Description: fmt.Sprintf("We're only a few minutes away from the convoy's scheduled departure time. "+
			"When we're asked to leave, please make sure you leave in the given order.\n \n"+
			"**Leaving Order** \n%s \n"+
			"If you're not included in the order, make sure to depart last.\n \n"+
			"**Please Note** \nWhen it's your turn to leave, please be very quick and take wide turns when leaving "+
			"so that you do not get stuck. If you're lagging too much, pull over or quick-travel to Service (F7 + Enter, then 1 + Enter)\n"+
			"When you've left the parking slot, maintain a gap of 70m (press Tab to view the distance you have with the truck in front).\n"+
			"Do not overtake unless otherwise stated.\n \nEnjoy the convoy! If you have any complaints, "+
			"please fill-up the feedback form and let us know how we can improve your experience.", strings.Replace(m.Content, commandsConfig.ConvoyConfig.OrderTrigger, "", 1)),
	}

	cm, _ := s.ChannelMessageSendEmbed(commandsConfig.ConvoyConfig.MessageChannelID, messageEmbed)

	s.MessageReactionAdd(cm.ChannelID, cm.ID, ":Check:729628532887781426")
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

func finishHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID ||
		!isAllowedToUse(m.Member) ||
		!strings.Contains(m.Content, commandsConfig.ConvoyConfig.FinishTrigger) {
		return
	}

	messageEmbed := &discordgo.MessageEmbed{
		Color: 12386317,
		Description: "Thank you for participating in today's convoy! We hope you had a good time. " +
			"Please feel free to post pictures in appropriate channels.\n \nWe hope to see you again at the upcoming convoy.",
	}

	s.ChannelMessageSendEmbed(commandsConfig.ConvoyConfig.MessageChannelID, messageEmbed)
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

func isAllowedToUse(member *discordgo.Member) bool {
	if member != nil && member.Roles != nil {
		for _, role := range commandsConfig.ConvoyConfig.WhitelistRoles {
			for _, memberRole := range member.Roles {
				if role == memberRole {
					return true
				}
			}
		}
	}

	return false
}
