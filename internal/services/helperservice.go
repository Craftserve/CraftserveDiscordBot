package services

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"log"
)

type HelperService struct {
	ServerRepo   repos.ServerRepo
	GiveawayRepo repos.GiveawayRepo
	UserRepo     repos.UserRepo
}

func NewHelperService(serverRepo *repos.ServerRepo, giveawayRepo *repos.GiveawayRepo, userRepo *repos.UserRepo) *HelperService {
	return &HelperService{
		ServerRepo:   *serverRepo,
		GiveawayRepo: *giveawayRepo,
		UserRepo:     *userRepo,
	}
}

func (h *HelperService) CheckHelpers(ctx context.Context, session *discordgo.Session, guildId string) {
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.Printf("(%s) checkHelpers#ServerRepo.GetServerConfigForGuild: %v", guildId, err)
		return
	}
	if serverConfig.HelperRoleThxesNeeded <= 0 {
		return
	}
	if serverConfig.HelperRoleId == "" {
		return
	}

	helpers, err := h.GiveawayRepo.GetParticipantsWithThxAmount(ctx, guildId, serverConfig.HelperRoleThxesNeeded)
	if err != nil {
		log.Printf("(%s) checkHelpers#GiveawayRepo.GetParticipantsWithThxAmount: %v", guildId, err)
		return
	}

	for _, helper := range helpers {
		isHelperBlacklisted, err := h.UserRepo.IsUserHelperBlacklisted(ctx, helper.UserId, guildId)
		if err != nil {
			log.Printf("(%s) checkHelpers#UserRepo.IsUserHelperBlacklisted: %v", guildId, err)
			continue
		}
		member, err := session.GuildMember(guildId, helper.UserId)
		if err != nil {
			log.Printf("(%s) checkHelpers#session.GuildMember: %v", guildId, err)
			continue
		}
		hasRole := discord.HasRoleById(member, serverConfig.HelperRoleId)
		if !hasRole && !isHelperBlacklisted {
			log.Printf("(%s) Adding helper role to %s (%s)", guildId, member.User.Username, helper.UserId)
			err = session.GuildMemberRoleAdd(guildId, helper.UserId, serverConfig.HelperRoleId)
			if err != nil {
				log.Printf("(%s) checkHelpers#session.GuildMemberRoleAdd: %v", guildId, err)
			}
			continue
		}
		if isHelperBlacklisted && hasRole {
			log.Printf("(%s) Removing helper role from %s (%s)", guildId, member.User.Username, helper.UserId)
			err = session.GuildMemberRoleRemove(guildId, helper.UserId, serverConfig.HelperRoleId)
			if err != nil {
				log.Printf("(%s) checkHelpers#session.GuildMemberRoleRemove: %v", guildId, err)
			}
			continue
		}
	}
}

func (h *HelperService) CheckHelper(ctx context.Context, session *discordgo.Session, guildId, memberId string) {
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.Printf("(%s) checkHelper#ServerRepo.GetServerConfigForGuild: %v", guildId, err)
		return
	}
	if serverConfig.HelperRoleThxesNeeded <= 0 {
		return
	}
	if serverConfig.HelperRoleId == "" {
		return
	}

	member, err := session.GuildMember(guildId, memberId)
	if err != nil {
		log.Printf("(%s) checkHelper#session.GuildMember: %v", guildId, err)
		return
	}

	hasHelperAmount, err := h.GiveawayRepo.HasThxAmount(ctx, guildId, memberId, serverConfig.HelperRoleThxesNeeded)
	if err != nil {
		log.Printf("(%s) checkHelper#GiveawayRepo.HasThxAmount: %v", guildId, err)
		return
	}
	isHelperBlacklisted, err := h.UserRepo.IsUserHelperBlacklisted(ctx, memberId, guildId)
	if err != nil {
		log.Printf("(%s) checkHelper#UserRepo.IsUserHelperBlacklisted: %v", guildId, err)
		return
	}
	hasRole := discord.HasRoleById(member, serverConfig.HelperRoleId)

	if isHelperBlacklisted && hasRole {
		log.Printf("(%s) Removing helper role from %s (%s)", guildId, member.User.Username, member.User.ID)
		err = session.GuildMemberRoleRemove(guildId, memberId, serverConfig.HelperRoleId)
		if err != nil {
			log.Printf("(%s) checkHelper#session.GuildMemberRoleRemove: %v", guildId, err)
		}
		return
	}
	if hasRole || isHelperBlacklisted || !hasHelperAmount {
		return
	}
	log.Printf("(%s) Adding helper role to %s (%s)", guildId, member.User.Username, member.User.ID)
	err = session.GuildMemberRoleAdd(guildId, memberId, serverConfig.HelperRoleId)
	if err != nil {
		log.Printf("(%s) checkHelper#session.GuildMemberRoleAdd: %v", guildId, err)
	}

}
