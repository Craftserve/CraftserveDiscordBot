package services

import (
	"context"
	"csrvbot/domain/entities"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type HelperService struct {
	ServerRepo   entities.ServerRepo
	GiveawayRepo entities.GiveawayRepo
	UserRepo     entities.UserRepo
}

func NewHelperService(serverRepo entities.ServerRepo, giveawayRepo entities.GiveawayRepo, userRepo entities.UserRepo) *HelperService {
	return &HelperService{
		ServerRepo:   serverRepo,
		GiveawayRepo: giveawayRepo,
		UserRepo:     userRepo,
	}
}

func (h *HelperService) CheckHelpers(ctx context.Context, session *discordgo.Session, guildId string) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("CheckHelpers#ServerRepo.GetServerConfigForGuild")
		return
	}
	if serverConfig.HelperRoleThxesNeeded <= 0 {
		log.Debug("HelperRoleThxesNeeded is 0 or less, not checking helpers")
		return
	}
	if serverConfig.HelperRoleId == "" {
		log.Debug("HelperRoleId is empty, not checking helpers")
		return
	}

	helpers, err := h.GiveawayRepo.GetParticipantsWithThxAmount(ctx, guildId, serverConfig.HelperRoleThxesNeeded)
	if err != nil {
		log.WithError(err).Error("CheckHelpers#GiveawayRepo.GetParticipantsWithThxAmount")
		return
	}

	for _, helper := range helpers {
		isHelperBlacklisted, err := h.UserRepo.IsUserHelperBlacklisted(ctx, helper.UserId, guildId)
		if err != nil {
			log.WithError(err).Error("CheckHelpers#UserRepo.IsUserHelperBlacklisted")
			continue
		}
		member, err := session.GuildMember(guildId, helper.UserId)
		if err != nil {
			log.WithError(err).Error("CheckHelpers#session.GuildMember")
			continue
		}
		hasRole := discord.HasRoleById(member, serverConfig.HelperRoleId)
		if !hasRole && !isHelperBlacklisted {
			log.Infof("Adding helper role to %s (%s)", member.User.Username, helper.UserId)
			err = session.GuildMemberRoleAdd(guildId, helper.UserId, serverConfig.HelperRoleId)
			if err != nil {
				log.WithError(err).Error("CheckHelpers#session.GuildMemberRoleAdd")
			}
			continue
		}
		if isHelperBlacklisted && hasRole {
			log.Infof("Removing helper role from %s (%s)", member.User.Username, helper.UserId)
			err = session.GuildMemberRoleRemove(guildId, helper.UserId, serverConfig.HelperRoleId)
			if err != nil {
				log.WithError(err).Error("CheckHelpers#session.GuildMemberRoleRemove")
			}
			continue
		}
	}
}

func (h *HelperService) CheckHelper(ctx context.Context, session *discordgo.Session, guildId, memberId string) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("CheckHelper#ServerRepo.GetServerConfigForGuild")
		return
	}
	if serverConfig.HelperRoleThxesNeeded <= 0 {
		log.Debug("HelperRoleThxesNeeded is 0 or less, not checking helper")
		return
	}
	if serverConfig.HelperRoleId == "" {
		log.Debug("HelperRoleId is empty, not checking helper")
		return
	}

	member, err := session.GuildMember(guildId, memberId)
	if err != nil {
		log.WithError(err).Error("CheckHelper#session.GuildMember")
		return
	}

	hasHelperAmount, err := h.GiveawayRepo.HasThxAmount(ctx, guildId, memberId, serverConfig.HelperRoleThxesNeeded)
	if err != nil {
		log.WithError(err).Error("CheckHelper#GiveawayRepo.HasThxAmount")
		return
	}
	isHelperBlacklisted, err := h.UserRepo.IsUserHelperBlacklisted(ctx, memberId, guildId)
	if err != nil {
		log.WithError(err).Error("CheckHelper#UserRepo.IsUserHelperBlacklisted")
		return
	}
	hasRole := discord.HasRoleById(member, serverConfig.HelperRoleId)

	if isHelperBlacklisted && hasRole {
		log.Debug("User is blacklisted, has role")
		log.Infof("Removing helper role from %s (%s)", member.User.Username, memberId)
		err = session.GuildMemberRoleRemove(guildId, memberId, serverConfig.HelperRoleId)
		if err != nil {
			log.WithError(err).Error("CheckHelper#session.GuildMemberRoleRemove")
		}
		return
	}
	if hasRole || isHelperBlacklisted || !hasHelperAmount {
		log.Debug("User has role, is blacklisted or doesn't have enough thxes - not adding role")
		return
	}
	log.Infof("Adding helper role to %s (%s)", member.User.Username, memberId)
	err = session.GuildMemberRoleAdd(guildId, memberId, serverConfig.HelperRoleId)
	if err != nil {
		log.WithError(err).Error("CheckHelper#session.GuildMemberRoleAdd")
	}

}
