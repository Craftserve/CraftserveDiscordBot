package discord

import (
	"context"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

func GetAllMembers(ctx context.Context, session *discordgo.Session, guildId string) []*discordgo.Member {
	log := logger.GetLoggerFromContext(ctx)
	after := ""
	var allMembers []*discordgo.Member
	for {
		members, err := session.GuildMembers(guildId, after, 1000)
		if err != nil {
			log.WithError(err).Error("getAllMembers#session.GuildMembers")
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
