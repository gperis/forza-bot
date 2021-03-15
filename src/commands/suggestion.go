package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type suggestionType struct {
	triggerWord    string
	embedChannelID string
}

var (
	suggestionTypes []*suggestionType = []*suggestionType{
		{
			triggerWord:    commandsConfig.SuggestionConfig.ServerTrigger,
			embedChannelID: commandsConfig.SuggestionConfig.ServerMessageChannelID,
		},
		{
			triggerWord:    commandsConfig.SuggestionConfig.VtcTrigger,
			embedChannelID: commandsConfig.SuggestionConfig.VtcMessageChannelID,
		},
	}
)

func suggestionHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	st, found := getSuggestionType(m.Content)

	if found != true {
		return
	}

	sm, _ := s.ChannelMessageSendEmbed(
		st.embedChannelID,
		&discordgo.MessageEmbed{
			Description: fmt.Sprintf(
				`%s
				
				<%s> Suggestion Awaiting Response
				This suggestion has not received a response yet. Please react below and help us decide.
				`,
				strings.Replace(m.Content, st.triggerWord, "", -1),
				commandsConfig.SuggestionConfig.WaitingEmoticonID,
			),
			Color: 12386317,
			Footer: &discordgo.MessageEmbedFooter{
				IconURL: fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s", m.Author.ID, m.Author.Avatar),
				Text:    fmt.Sprintf("Suggested by %s", m.Author.Username),
			},
		},
	)

	s.MessageReactionAdd(sm.ChannelID, sm.ID, commandsConfig.SuggestionConfig.AcceptedEmoticonID)
	s.MessageReactionAdd(sm.ChannelID, sm.ID, commandsConfig.SuggestionConfig.RejectedEmoticonID)

	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

// Get suggestion type by the content
func getSuggestionType(c string) (*suggestionType, bool found) {
	for _, suggestionType := range suggestionTypes {
		if strings.Contains(c, suggestionType.triggerWord) {
			return suggestionType, true
		}
	}

	return &suggestionType{}, false
}
