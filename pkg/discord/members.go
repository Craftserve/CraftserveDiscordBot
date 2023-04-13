package discord

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func GetAllMembers(session *discordgo.Session, guildId string) []*discordgo.Member {
	after := ""
	var allMembers []*discordgo.Member
	for {
		members, err := session.GuildMembers(guildId, after, 1000)
		if err != nil {
			log.Println("("+guildId+") getAllMembers#session.GuildMembers", err)
			return nil
		}
		allMembers = append(allMembers, members...)
		if len(members) != 1000 {
			break
		}
		after = members[999].User.ID
	}
	return allMembers
}
