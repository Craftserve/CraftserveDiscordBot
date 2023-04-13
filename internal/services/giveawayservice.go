package services

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg/discord"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"time"
)

type GiveawayService struct {
	CsrvClient   CsrvClient
	ServerRepo   repos.ServerRepo
	GiveawayRepo repos.GiveawayRepo
}

func NewGiveawayService(csrvClient *CsrvClient, serverRepo *repos.ServerRepo, giveawayRepo *repos.GiveawayRepo) *GiveawayService {
	return &GiveawayService{
		CsrvClient:   *csrvClient,
		ServerRepo:   *serverRepo,
		GiveawayRepo: *giveawayRepo,
	}
}

func (h *GiveawayService) FinishGiveaway(ctx context.Context, s *discordgo.Session, guildId string) {
	giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, guildId)
	if err != nil {
		log.Printf("(%s) Could not get giveaway: %v", guildId, err)
		return
	}
	_, err = s.Guild(guildId)
	if err != nil {
		log.Printf("(%s) finishGiveaway#session.Guild: %v", guildId, err)
		return
	}

	giveawayChannelId, err := h.ServerRepo.GetMainChannelForGuild(ctx, guildId)
	if err != nil {
		log.Printf("(%s) finishGiveaway#serverRepo.GetMainChannelForGuild: %v", guildId, err)
		return
	}

	participants, err := h.GiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.Printf("(%s) finishGiveaway#giveawayRepo.GetParticipantsForGiveaway: %v", guildId, err)
	}

	if participants == nil || len(participants) == 0 {
		message, err := s.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był w loterii.")
		if err != nil {
			log.Printf("(%s) finishGiveaway#session.ChannelMessageSend: %v", guildId, err)
		}
		err = h.GiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID, "", "", "")
		if err != nil {
			log.Printf("(%s) finishGiveaway#giveawayRepo.UpdateGiveaway: %v", guildId, err)
		}
		log.Printf("(%s) Giveaway ended without any participants.", guildId)
		return
	}

	code, err := h.CsrvClient.GetCSRVCode()
	if err != nil {
		log.Printf("(%s) finishGiveaway#csrvClient.GetCSRVCode: %v", guildId, err)
		_, err = s.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
		if err != nil {
			return
		}
		return
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	winner := participants[rand.Intn(len(participants))]

	member, err := s.GuildMember(guildId, winner.UserId)
	if err != nil {
		log.Printf("(%s) finishGiveaway#session.GuildMember: %v", guildId, err)
		return
	}
	dmEmbed := discord.ConstructWinnerEmbed(code)
	dm, err := s.UserChannelCreate(winner.UserId)
	if err != nil {
		log.Printf("(%s) finishGiveaway#session.UserChannelCreate: %v", guildId, err)
	}
	_, err = s.ChannelMessageSendEmbed(dm.ID, dmEmbed)

	mainEmbed := discord.ConstructChannelWinnerEmbed(member.User.Username)
	message, err := s.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed: mainEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Kliknij tutaj, aby wyświetlić kod",
						Style:    discordgo.SuccessButton,
						CustomID: "winnercode",
						Emoji: discordgo.ComponentEmoji{
							Name: "🎉",
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.Printf("(%s) finishGiveaway#session.ChannelMessageSendComplex: %v", guildId, err)
	}

	err = h.GiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID, code, winner.UserId, member.User.Username)
	if err != nil {
		log.Printf("(%s) finishGiveaway#giveawayRepo.UpdateGiveaway: %v", guildId, err)
	}
	log.Printf("(%s) Giveaway ended with a winner: %s", guildId, member.User.Username)

}

func (h *GiveawayService) FinishGiveaways(ctx context.Context, s *discordgo.Session) {
	giveaways, err := h.GiveawayRepo.GetUnfinishedGiveaways(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		log.Printf("finishGiveaways#giveawayRepo.GetUnfinishedGiveaways: %v", err)
		return
	}
	for _, giveaway := range giveaways {
		h.FinishGiveaway(ctx, s, giveaway.GuildId)
		guild, err := s.Guild(giveaway.GuildId)
		if err == nil {
			h.CreateMissingGiveaways(ctx, s, guild)
		}
	}

}

func (h *GiveawayService) CreateMissingGiveaways(ctx context.Context, s *discordgo.Session, guild *discordgo.Guild) {
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
