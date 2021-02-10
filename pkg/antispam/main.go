package antispam

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/admin"
	"github.com/gperis/forza-bot/pkg/config"
	"github.com/gperis/forza-bot/pkg/discord_log"
	"strings"
	"time"
)

type conf struct {
	WarningMessagesCount int `mapstructure:"warning_messages_count"`
	WarningTimespan      int `mapstructure:"warning_timespan"`
	BanWarningsCount     int `mapstructure:"ban_warnings_count"`
	BanWarningsTimespan  int `mapstructure:"ban_warnings_timespan"`
	BanDays              int `mapstructure:"ban_days"`
}

type userState struct {
	UserID               string
	ChannelID            string
	GuildID              string
	Messages             []*discordgo.Message
	LastMessageTimestamp time.Time
	SpamWarningState     *warningState
}

type warningState struct {
	Count                int
	LastWarningTimestamp time.Time
}

var (
	moduleConf    conf
	userStates    []userState
	warningStates []warningState
)

func init() {
	config.Load("antispam", &moduleConf)

	userStates = make([]userState, 1)
	warningStates = make([]warningState, 1)
}

func StartModule(dg *discordgo.Session) {
	go cleanup()

	dg.AddHandler(handler)
}

func handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || admin.IsStaffMember(m.Member) {
		return
	}

	state := getUserState(m.Author.ID, m.ChannelID)

	if time.Since(state.LastMessageTimestamp).Seconds() > float64(moduleConf.WarningTimespan) {
		state.Messages = nil
	}

	state.GuildID = m.GuildID
	state.Messages = append(state.Messages, m.Message)
	state.LastMessageTimestamp = time.Now()

	if len(state.Messages) > moduleConf.WarningMessagesCount {
		state.activateAntiSpam(s)
	}
}

func (us *userState) activateAntiSpam(s *discordgo.Session) {
	var messagesText []string

	for _, m := range us.Messages {
		messagesText = append(messagesText, m.Content)
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}

	us.Messages = nil

	warningEmbed := &discordgo.MessageEmbed{
		Title:       "Auto Moderation",
		Description: "Please refrain from sending so many messages so frequently. Repeating this action will lead to consequences.",
		Color:       10038562,
	}

	s.ChannelMessageSendEmbed(us.ChannelID, warningEmbed)

	discord_log.LogIncident(
		s,
		fmt.Sprintf(
			"<@!%s> (%s) sent more than %d messages within %d seconds in <#%s>.\n\n**The messages:**\n>>> %s",
			us.UserID,
			us.UserID,
			moduleConf.WarningMessagesCount,
			moduleConf.WarningTimespan,
			us.ChannelID,
			strings.Join(messagesText, "\n"),
		),
		"Anti Spam",
	)

	if us.SpamWarningState.Count > moduleConf.BanWarningsCount {
		us.banUser(s)
		return
	}

	us.SpamWarningState.Count++
	us.SpamWarningState.LastWarningTimestamp = time.Now()
}

func (us *userState) banUser(s *discordgo.Session) {
	banMessageEmbed := &discordgo.MessageEmbed{
		Title: "Auto Moderation | User Banned",
		Description: fmt.Sprintf(
			"<@!%s> has been banned from the server by automatic moderation due to violating server rules.\n\n**Reason:**\n>>> Spamming",
			us.UserID,
		),
		Color: 10038562,
	}

	s.ChannelMessageSendEmbed(us.ChannelID, banMessageEmbed)

	discord_log.LogIncident(
		s,
		fmt.Sprintf(
			"<@!%s> (%s) was banned after repeating an action within %d minutes of receiving %d automated warnings.\n\n**Reason:**\n>>> Spamming",
			us.UserID,
			us.UserID,
			moduleConf.BanWarningsTimespan/60,
			moduleConf.BanWarningsCount,
		),
		"Anti Spam",
	)

	s.GuildBanCreateWithReason(
		us.GuildID,
		us.UserID,
		fmt.Sprintf("You have been banned from the server for %d days for violating the server rules.\n\n**Reason**:\n>>> Spamming", moduleConf.BanDays),
		moduleConf.BanDays,
	)
}

func getUserState(UserID string, ChannelID string) *userState {
	for i, s := range userStates {
		if s.UserID == UserID && s.ChannelID == ChannelID {
			return &userStates[i]
		}
	}

	newUserState := userState{
		UserID:               UserID,
		ChannelID:            ChannelID,
		Messages:             nil,
		LastMessageTimestamp: time.Now(),
		SpamWarningState:     &warningState{},
	}

	userStates = append(userStates, newUserState)

	return &userStates[len(userStates)-1]
}

func cleanup() {
	for {
		time.Sleep(60 * time.Second)

		for i, s := range userStates {
			if s.SpamWarningState != nil &&
				s.SpamWarningState.Count > 0 &&
				time.Since(s.SpamWarningState.LastWarningTimestamp).Seconds() < float64(moduleConf.BanWarningsTimespan) {
				continue
			}

			if time.Since(s.LastMessageTimestamp).Seconds() >= 80 && i+1 < len(userStates) {
				userStates = removeState(userStates, i)
			}
		}
	}
}

func removeState(s []userState, i int) []userState {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}
