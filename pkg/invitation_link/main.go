package invitation_link

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/admin"
	"github.com/gperis/forza-bot/pkg/discord_log"
	"mvdan.cc/xurls/v2"
	"net/http"
	"net/url"
	"strings"
)

func StartModule(dg *discordgo.Session) {
	dg.AddHandler(handler)
}

func handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || admin.IsStaffMember(m.Member) {
		return
	}

	rxRelaxed := xurls.Relaxed()
	matches := rxRelaxed.FindAllString(m.Content, -1)

	for i := range matches {
		parsedUrl, err := url.Parse(matches[i])

		if err != nil {
			continue
		}

		if parsedUrl.Scheme == "" {
			parsedUrl.Scheme = "http"
		}

		resp, err := http.Get(parsedUrl.String())

		if err != nil {
			continue
		}

		finalURL := resp.Request.URL.String()

		if strings.Contains(finalURL, "/invite/") != false {
			s.ChannelMessageDelete(m.ChannelID, m.ID)

			userWarningEmbed := &discordgo.MessageEmbed{
				Title:       "Auto Moderation",
				Description: "Please refrain from sharing Discord invitation links in the server. Repeating this action will lead to consequences.",
				Color:       12386317,
			}

			s.ChannelMessageSendEmbed(m.ChannelID, userWarningEmbed)

			discord_log.LogIncident(
				s,
				fmt.Sprintf(
					"<@!%s> (%s) sent a message containing an invitation link in <#%s>.\n\n**The message:**\n>>> %s",
					m.Author.ID,
					m.Author.ID,
					m.ChannelID,
					m.Content,
				),
				"Anti Invitation Link",
			)

			break
		}
	}
}
