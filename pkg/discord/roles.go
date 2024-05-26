package discord

import (
	"context"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
	"strconv"
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

var LevelPrefix string

func GetMemberLevel(ctx context.Context, session *discordgo.Session, member *discordgo.Member, guildId string) (int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId).WithUser(member.User.ID)
	log.Debug("Checking member level")

	for _, roleId := range member.Roles {
		role, err := session.State.Role(guildId, roleId)
		if err != nil {
			log.WithError(err).Error("GetMemberLevel#session.State.Role")
			return 0, err
		}

		if len(role.Name) > len(LevelPrefix) && role.Name[:len(LevelPrefix)] == LevelPrefix {
			level, err := strconv.Atoi(role.Name[len(LevelPrefix):])
			if err != nil {
				log.WithError(err).Error("GetMemberLevel#strconv.Atoi")
				return 0, err
			}

			return level, nil
		}
	}

	return 0, nil
}

func GetRoleForLevel(ctx context.Context, session *discordgo.Session, guildId string, level int) (*discordgo.Role, error) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId).WithField("level", level)
	log.Debug("Getting role for level")

	guild, err := session.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("GetRoleForLevel#session.Guild")
		return nil, err
	}

	for _, role := range guild.Roles {
		if len(role.Name) > len(LevelPrefix) && role.Name[:len(LevelPrefix)] == LevelPrefix {
			roleLevel, err := strconv.Atoi(role.Name[len(LevelPrefix):])
			if err != nil {
				log.WithError(err).Error("GetRoleForLevel#strconv.Atoi")
				return nil, err
			}

			if roleLevel == level {
				return role, nil
			}
		}
	}

	return nil, nil
}

func ValidateLevels(ctx context.Context, session *discordgo.Session, guildId string, levels []int) (bool, error) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	log.Debug("Validating levels")

	guild, err := session.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("ValidateLevels#session.Guild")
		return false, err
	}

	for _, level := range levels {
		found := false
		for _, role := range guild.Roles {
			if len(role.Name) > len(LevelPrefix) && role.Name[:len(LevelPrefix)] == LevelPrefix {
				roleLevel, err := strconv.Atoi(role.Name[len(LevelPrefix):])
				if err != nil {
					log.WithError(err).Error("ValidateLevels#strconv.Atoi")
					return false, err
				}

				if roleLevel == level {
					found = true
					break
				}
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}
