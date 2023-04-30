package discord

import (
	"context"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

func HasPermission(ctx context.Context, session *discordgo.Session, member *discordgo.Member, guildId string, permission int64) bool {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId).WithUser(member.User.ID)
	g, err := session.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("hasPermisson#session.Guild")
		return false
	}
	if g.OwnerID == member.User.ID {
		return true
	}
	for _, roleId := range member.Roles {
		role, err := session.State.Role(guildId, roleId)
		if err != nil {
			log.Println("("+guildId+") "+"hasPermisson#session.State.Role", err)
			return false
		}
		if role.Permissions&permission != 0 {
			return true
		}
	}
	return false
}

func HasRoleById(member *discordgo.Member, roleId string) bool {
	for _, role := range member.Roles {
		if role == roleId {
			return true
		}
	}

	return false
}

func HasAdminPermissions(ctx context.Context, session *discordgo.Session, member *discordgo.Member, adminRoleId, guildId string) bool {
	if HasPermission(ctx, session, member, guildId, 8) {
		return true
	}
	if HasRoleById(member, adminRoleId) {
		return true
	}
	return false
}
