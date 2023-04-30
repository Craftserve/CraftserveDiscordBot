package listeners

import (
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"csrvbot/pkg"
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
	if m.GuildID == "" { //can it be even empty?
		return
	}

	h.SavedRoleService.UpdateMemberSavedRoles(ctx, m.Roles, m.User.ID, m.GuildID)
}
