package listeners

import (
	"context"
	"csrvbot/domain/entities"
	"csrvbot/internal/services"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"time"
)

type GuildCreateListener struct {
	//GiveawaysRepo    entities.GiveawaysRepo // unused
	ServerRepo       entities.ServerRepo
	GiveawayService  services.GiveawayService
	HelperService    services.HelperService
	SavedRoleService services.SavedroleService
}

func NewGuildCreateListener(serverRepo entities.ServerRepo, giveawayService *services.GiveawayService, helperService *services.HelperService, savedRoleService *services.SavedroleService) GuildCreateListener {
	return GuildCreateListener{
		ServerRepo:       serverRepo,
		GiveawayService:  *giveawayService,
		HelperService:    *helperService,
		SavedRoleService: *savedRoleService,
	}
}

func (h GuildCreateListener) Handle(s *discordgo.Session, g *discordgo.GuildCreate) {
	ctx := pkg.CreateContext()
	log := logger.GetLoggerFromContext(ctx)
	log.WithFields(logrus.Fields{
		"guild":      g.Guild.ID,
		"guild_name": g.Guild.Name,
	}).Info("Registered guild")

	log.Debug("Creating configuration if not exists")
	h.createConfigurationIfNotExists(ctx, s, g.Guild.ID)

	log.Debug("Creating missing thx giveaways for guild")
	h.GiveawayService.CreateMissingThxGiveaways(ctx, s, g.Guild)

	log.Debug("Creating missing unconditional giveaways for guild")
	h.GiveawayService.CreateJoinableGiveaway(ctx, s, g.Guild, false)

	log.Debug("Creating missing conditional giveaways for guild")
	h.GiveawayService.CreateJoinableGiveaway(ctx, s, g.Guild, true)

	log.Debug("Updating all members saved roles for guild")
	h.updateAllMembersSavedRoles(ctx, s, g.Guild.ID)

	log.Debug("Checking helpers for guild")
	h.HelperService.CheckHelpers(ctx, s, g.Guild.ID)
}

func (h GuildCreateListener) createConfigurationIfNotExists(ctx context.Context, session *discordgo.Session, guildID string) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildID)
	var giveawayChannel string
	channels, _ := session.GuildChannels(guildID)
	for _, channel := range channels {
		if channel.Name == "giveaway" {
			giveawayChannel = channel.ID
			break
		}
	}
	var adminRole string
	roles, _ := session.GuildRoles(guildID)
	for _, role := range roles {
		if role.Name == "CraftserveBotAdmin" {
			adminRole = role.ID
			break
		}
	}

	_, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("Creating server config")
			err = h.ServerRepo.InsertServerConfig(ctx, guildID, giveawayChannel, adminRole)
			if err != nil {
				log.WithError(err).Error("Could not create server config", err)
			}
		} else {
			log.WithError(err).Error("Could not get server config")
		}
	}
}

func (h GuildCreateListener) updateAllMembersSavedRoles(ctx context.Context, session *discordgo.Session, guildId string) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	startTime := time.Now()
	guildMembers := discord.GetAllMembers(ctx, session, guildId)
	for _, member := range guildMembers {
		if len(member.Roles) == 0 {
			continue
		}
		h.SavedRoleService.UpdateMemberSavedRoles(ctx, member.Roles, member.User.ID, guildId)
	}
	log.Debugf("Finished updating all members saved roles in %s", time.Since(startTime).String())
}
