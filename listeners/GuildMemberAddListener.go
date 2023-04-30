package listeners

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type GuildMemberAddListener struct {
	UserRepo repos.UserRepo
}

func NewGuildMemberAddListener(userRepo *repos.UserRepo) GuildMemberAddListener {
	return GuildMemberAddListener{
		UserRepo: *userRepo,
	}
}

func (h GuildMemberAddListener) Handle(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	ctx := pkg.CreateContext()
	if m.GuildID == "" { //can it be even empty?
		return
	}
	log := logger.GetLoggerFromContext(ctx).WithGuild(m.GuildID).WithUser(m.User.ID)
	ctx = logger.ContextWithLogger(ctx, log)
	h.restoreMemberRoles(ctx, s, m.Member, m.GuildID)
}

func (h GuildMemberAddListener) restoreMemberRoles(ctx context.Context, s *discordgo.Session, member *discordgo.Member, guildId string) {
	log := logger.GetLoggerFromContext(ctx)
	memberRoles, err := h.UserRepo.GetRolesForMember(ctx, guildId, member.User.ID)
	for _, role := range memberRoles {
		err = s.GuildMemberRoleAdd(guildId, member.User.ID, role.RoleId)
		if err != nil {
			log.WithError(err).Error("restoreMemberRoles#session.GuildMemberRoleAdd")
			continue
		}
	}
}
