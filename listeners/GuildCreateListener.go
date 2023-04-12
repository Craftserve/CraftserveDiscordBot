package listeners

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
)

type GuildCreateListener struct {
	GiveawayRepo repos.GiveawayRepo
	ServerRepo   repos.ServerRepo
	UserRepo     repos.UserRepo
}

func NewGuildCreateListener(giveawayRepo *repos.GiveawayRepo, serverRepo *repos.ServerRepo, userRepo *repos.UserRepo) GuildCreateListener {
	return GuildCreateListener{
		GiveawayRepo: *giveawayRepo,
		ServerRepo:   *serverRepo,
		UserRepo:     *userRepo,
	}
}

func (h GuildCreateListener) Handle(s *discordgo.Session, g *discordgo.GuildCreate) {
	ctx := pkg.CreateContext()
	log.Println("Registered guild (" + g.Name + "#" + g.ID + ")")

	h.createConfigurationIfNotExists(ctx, s, g.Guild.ID)
	h.createMissingGiveaways(ctx, s, g.Guild)
	h.updateAllMembersSavedRoles(ctx, s, g.Guild.ID)
	discord.CheckHelpers(ctx, s, h.ServerRepo, h.GiveawayRepo, h.UserRepo, g.Guild.ID)
}

func (h GuildCreateListener) createConfigurationIfNotExists(ctx context.Context, session *discordgo.Session, guildID string) {
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
		if err == sql.ErrNoRows {
			err = h.ServerRepo.InsertServerConfig(ctx, guildID, giveawayChannel, adminRole)
			if err != nil {
				log.Println("("+guildID+") Could not create server config", err)
			}
		} else {
			log.Println("("+guildID+") Could not get server config", err)
		}
	}
}

func (h GuildCreateListener) updateAllMembersSavedRoles(ctx context.Context, session *discordgo.Session, guildId string) {
	guildMembers := discord.GetAllMembers(session, guildId)
	for _, member := range guildMembers {
		h.UserRepo.UpdateMemberSavedRoles(ctx, member.Roles, member.User.ID, guildId)
	}
}

func (h GuildCreateListener) createMissingGiveaways(ctx context.Context, s *discordgo.Session, guild *discordgo.Guild) {
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guild.ID)
	if err != nil {
		log.Printf("(%s) createMissingGiveaways#ServerRepo.GetServerConfigForGuild: %v", guild.ID, err)
		return
	}
	giveawayChannelId := serverConfig.MainChannel
	_, err = s.Channel(giveawayChannelId)
	if err != nil {
		log.Printf("(%s) createMissingGiveaways#Session.Channel: %v", guild.ID, err)
		return
	}
	_, err = h.GiveawayRepo.GetGiveawayForGuild(ctx, guild.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("(%s) createMissingGiveaways#giveawayRepo.GetGiveawayForGuild: %v", guild.ID, err)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		err = h.GiveawayRepo.InsertGiveaway(ctx, guild.ID, guild.Name)
		if err != nil {
			log.Printf("(%s) createMissingGiveaways#giveawayRepo.InsertGiveaway: %v", guild.ID, err)
			return
		}
	}
}
