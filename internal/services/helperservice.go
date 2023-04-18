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

	members := discord.GetAllMembers(session, guildId)

	helpers, err := h.GiveawayRepo.GetParticipantsWithThxAmount(ctx, guildId, serverConfig.HelperRoleThxesNeeded)
	if err != nil {
		log.Printf("(%s) checkHelpers#GiveawayRepo.GetParticipantsWithThxAmount: %v", guildId, err)
		return
	}

	for _, member := range members {
		shouldHaveRole := false
		for _, helper := range helpers {
			isHelperBlacklisted, err := h.UserRepo.IsUserHelperBlacklisted(ctx, member.User.ID, guildId)
			if err != nil {
				log.Printf("(%s) checkHelpers#UserRepo.IsUserHelperBlacklisted: %v", guildId, err)
				continue
			}
			if isHelperBlacklisted {
				shouldHaveRole = false
				break
			}
			if helper.UserId == member.User.ID {
				shouldHaveRole = true
				break
			}
		}
		hasRole := discord.HasRoleById(member, serverConfig.HelperRoleId)
		if shouldHaveRole {
			if hasRole {
				continue
			}
			log.Printf("(%s) Adding helper role to %s (%s)", guildId, member.User.Username, member.User.ID)
			err = session.GuildMemberRoleAdd(guildId, member.User.ID, serverConfig.HelperRoleId)
			if err != nil {
				log.Printf("(%s) checkHelpers#session.GuildMemberRoleAdd: %v", guildId, err)
			}
		} else {
			if !hasRole {
				continue
			}
			log.Printf("(%s) Removing helper role from %s (%s)", guildId, member.User.Username, member.User.ID)
			err = session.GuildMemberRoleRemove(guildId, member.User.ID, serverConfig.HelperRoleId)
			if err != nil {
				log.Printf("(%s) checkHelpers#session.GuildMemberRoleRemove: %v", guildId, err)
			}
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
	if !hasHelperAmount {
		if hasRole {
			log.Printf("(%s) Removing helper role from %s (%s)", guildId, member.User.Username, member.User.ID)
			err = session.GuildMemberRoleRemove(guildId, memberId, serverConfig.HelperRoleId)
			if err != nil {
				log.Printf("(%s) checkHelper#session.GuildMemberRoleRemove: %v", guildId, err)
			}
		}
		return
	}

	if isHelperBlacklisted && hasRole {
		log.Printf("(%s) Removing helper role from %s (%s)", guildId, member.User.Username, member.User.ID)
		err = session.GuildMemberRoleRemove(guildId, memberId, serverConfig.HelperRoleId)
		if err != nil {
			log.Printf("(%s) checkHelper#session.GuildMemberRoleRemove: %v", guildId, err)
		}
		return
	}
	if hasRole || isHelperBlacklisted {
		return
	}
	log.Printf("(%s) Adding helper role to %s (%s)", guildId, member.User.Username, member.User.ID)
	err = session.GuildMemberRoleAdd(guildId, memberId, serverConfig.HelperRoleId)
	if err != nil {
		log.Printf("(%s) checkHelper#session.GuildMemberRoleAdd: %v", guildId, err)
	}

}
