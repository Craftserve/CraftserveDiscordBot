package listeners

import (
	"csrvbot/domain/entities"
	"csrvbot/internal/services"
	"csrvbot/pkg"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type GuildMemberUpdateListener struct {
	UserRepo         entities.UserRepo
	SavedRoleService services.SavedroleService
}

func NewGuildMemberUpdateListener(userRepo entities.UserRepo, savedRoleService *services.SavedroleService) GuildMemberUpdateListener {
	return GuildMemberUpdateListener{
		UserRepo:         userRepo,
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
