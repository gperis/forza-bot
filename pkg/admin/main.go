package admin

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gperis/forza-bot/pkg/config"
)

type conf struct {
	Mode       string   `mapstructure:"mode"`
	StaffRoles []string `mapstructure:"staff_roles"`
}

var moduleConf conf

func init() {
	config.Load("admin", &moduleConf)
}

func IsStaffMember(member *discordgo.Member) bool {
	if member == nil || len(member.Roles) == 0 {
		return false
	}

	for _, memberRole := range member.Roles {
		for _, staffRole := range moduleConf.StaffRoles {
			if memberRole == staffRole {
				return true
			}
		}
	}

	return false
}

func IsDevelopment() bool {
	return true
}
