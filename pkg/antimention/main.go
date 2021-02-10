package antimention

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
	WarningMentionsCount    int `mapstructure:"warning_mentions_count"`
	WarningTimespan         int `mapstructure:"warning_timespan"`
	BanWarningsCount        int `mapstructure:"ban_warnings_count"`
	BanWarningsTimespan     int `mapstructure:"ban_warnings_timespan"`
	InstantBanMentionsCount int `mapstructure:"instant_ban_mentions_count"`
	BanDays                 int `mapstructure:"ban_days"`
}

type userState struct {
	UserID               string
	ChannelID            string
	GuildID              string
	Messages             []*discordgo.Message
	MentionsCount        int
	LastMessageTimestamp time.Time
	WarningState         *warningState
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
	config.Load("antimention", &moduleConf)

	userStates = make([]userState, 1)
	warningStates = make([]warningState, 1)
}

func StartModule(dg *discordgo.Session) {
	go cleanup()

	dg.AddHandler(handler)
}

func handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	mentionsCount := len(m.Mentions)
	if m.Author.ID == s.State.User.ID || mentionsCount == 0 || admin.IsStaffMember(m.Member) {
		return
	}

	state := getUserState(m.Author.ID, m.ChannelID)
	state.GuildID = m.GuildID
	state.LastMessageTimestamp = time.Now()
	state.MentionsCount += mentionsCount

	if state.MentionsCount >= moduleConf.InstantBanMentionsCount {
		state.banUser(s, true)
		return
	}

	state.Messages = append(state.Messages, m.Message)

	if state.MentionsCount >= moduleConf.WarningMentionsCount &&
		time.Since(state.LastMessageTimestamp).Seconds() < float64(moduleConf.WarningTimespan) {
		state.showWarning(s)
	}
}

func (us *userState) showWarning(s *discordgo.Session) {
	var messagesText []string

	for _, m := range us.Messages {
		messagesText = append(messagesText, m.Content)
	}

	warningEmbed := &discordgo.MessageEmbed{
		Title:       "Auto Moderation",
		Description: "Please refrain from making so many mentions. Repeating this action will lead to consequences.",
		Color:       10038562,
	}

	s.ChannelMessageSendEmbed(us.ChannelID, warningEmbed)

	discord_log.LogIncident(
		s,
		fmt.Sprintf(
			"<@!%s> (%s) sent more than %d messages containing mentions within %d seconds in <#%s>.\n\n**The messages:**\n>>> %s",
			us.UserID,
			us.UserID,
			moduleConf.WarningMentionsCount,
			moduleConf.WarningTimespan,
			us.ChannelID,
			strings.Join(messagesText, "\n"),
		),
		"Anti Mention",
	)

	if us.WarningState.Count > moduleConf.BanWarningsCount {
		us.banUser(s, false)
		return
	}

	us.MentionsCount = 0
	us.Messages = nil
	us.WarningState.Count++
	us.WarningState.LastWarningTimestamp = time.Now()
}

func (us *userState) banUser(s *discordgo.Session, instantBan bool) {
	banMessageEmbed := &discordgo.MessageEmbed{
		Title: "Auto Moderation | User Banned",
		Description: fmt.Sprintf(
			"<@!%s> has been banned from the server by automatic moderation due to violating server rules."+
				"\n\n**Reason:**\n>>> Spamming, excessive mentioning",
			us.UserID,
		),
		Color: 10038562,
	}

	s.ChannelMessageSendEmbed(us.ChannelID, banMessageEmbed)

	logMessage := fmt.Sprintf(
		"<@!%s> (%s) was banned after repeating an action within %d minutes of receiving %d automated warnings."+
			"\n\n**Reason:**\n>>> Spamming, excessive mentioning",
		us.UserID,
		us.UserID,
		moduleConf.BanWarningsTimespan/60,
		moduleConf.BanWarningsCount,
	)

	if instantBan {
		logMessage = fmt.Sprintf(
			"<@!%s> (%s) was banned without any automated warnings."+
				"\n\n**Reason:**\n>>> Spamming, excessive mentioning",
			us.UserID,
			us.UserID,
		)
	}

	discord_log.LogIncident(
		s,
		logMessage,
		"Anti Spam",
	)

	s.GuildBanCreateWithReason(
		us.GuildID,
		us.UserID,
		fmt.Sprintf("You have been banned from the server for %d days for violating the server rules.\n\n**Reason**:\n>>> Spamming, excessive mentioning", moduleConf.BanDays),
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
		LastMessageTimestamp: time.Now(),
		WarningState:         &warningState{},
	}

	userStates = append(userStates, newUserState)

	return &userStates[len(userStates)-1]
}

func cleanup() {
	for {
		time.Sleep(60 * time.Second)

		for i, s := range userStates {
			if s.WarningState != nil &&
				s.WarningState.Count > 0 &&
				time.Since(s.WarningState.LastWarningTimestamp).Seconds() < float64(moduleConf.BanWarningsTimespan) {
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
