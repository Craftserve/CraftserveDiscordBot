package listeners

import (
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"csrvbot/pkg"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type GuildMemberUpdateListener struct {
	UserRepo         repos.UserRepo
	SavedRoleService services.SavedroleService
}

func NewGuildMemberUpdateListener(userRepo *repos.UserRepo, savedRoleService *services.SavedroleService) GuildMemberUpdateListener {
	return GuildMemberUpdateListener{
		UserRepo:         *userRepo,
		SavedRoleService: *savedRoleService,
	}
}

func (h GuildMemberUpdateListener) Handle(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	ctx := pkg.CreateContext()
	log := logger.GetLoggerFromContext(ctx).WithGuild(m.GuildID).WithUser(m.User.ID)
	if m.GuildID == "" { //can it be even empty?
		return
	}

	log.Debug("Updating member roles")
	h.SavedRoleService.UpdateMemberSavedRoles(ctx, m.Roles, m.User.ID, m.GuildID)
}
