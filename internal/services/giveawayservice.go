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
	"strings"
	"time"
)

type GiveawayService struct {
	CsrvClient          CsrvClient
	ServerRepo          repos.ServerRepo
	GiveawayRepo        repos.GiveawayRepo
	MessageGiveawayRepo repos.MessageGiveawayRepo
}

func NewGiveawayService(csrvClient *CsrvClient, serverRepo *repos.ServerRepo, giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo) *GiveawayService {
	return &GiveawayService{
		CsrvClient:          *csrvClient,
		ServerRepo:          *serverRepo,
		GiveawayRepo:        *giveawayRepo,
		MessageGiveawayRepo: *messageGiveawayRepo,
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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	winner := participants[r.Intn(len(participants))]

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
						CustomID: "thxwinnercode",
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

func (h *GiveawayService) FinishMessageGiveaways(ctx context.Context, session *discordgo.Session) {
	guildIds, err := h.ServerRepo.GetGuildsWithMessageGiveawaysEnabled(ctx)
	if err != nil {
		log.Printf("finishMessageGiveaways#serverRepo.GetGuildsWithMessageGiveawaysEnabled: %v", err)
		return
	}

	for _, guildId := range guildIds {
		_, err := session.Guild(guildId)
		if err != nil {
			log.Printf("(%s) finishMessageGiveaways#session.Guild: %v", guildId, err)
			continue
		}
		h.FinishMessageGiveaway(ctx, session, guildId)

	}
}

func (h *GiveawayService) FinishMessageGiveaway(ctx context.Context, session *discordgo.Session, guildId string) {
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#serverRepo.GetServerConfigForGuild: %v", guildId, err)
		return
	}

	if serverConfig.MessageGiveawayWinners == 0 {
		return
	}

	giveawayChannelId, err := h.ServerRepo.GetMainChannelForGuild(ctx, guildId)
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#serverRepo.GetMainChannelForGuild: %v", guildId, err)
		return
	}

	participants, err := h.MessageGiveawayRepo.GetUsersWithMessagesFromLastDays(ctx, 30, guildId)
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#messageGiveawayRepo.GetUsersWithMessagesFromLastDays: %v", guildId, err)
		return
	}

	if len(participants) == 0 {
		_, err := session.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był aktywny.")
		if err != nil {
			log.Printf("(%s) finishMessageGiveaway#session.ChannelMessageSend: %v", guildId, err)
		}
		log.Printf("(%s) Message giveaway ended without any participants.", guildId)
		return
	}

	messageGiveawayRepoTx, tx, err := h.MessageGiveawayRepo.WithTx(ctx, nil)
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#MessageGiveawayRepo.WithTx: %v", guildId, err)
		return
	}
	defer tx.Rollback()

	err = messageGiveawayRepoTx.InsertMessageGiveaway(ctx, guildId)
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#messageGiveawayRepoTx.InsertMessageGiveaway: %v", guildId, err)
		return
	}

	giveaway, err := messageGiveawayRepoTx.GetMessageGiveaway(ctx, guildId)
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#MessageGiveawayRepo.GetMessageGiveaway: %v", guildId, err)
		return
	}

	winnerIds := make([]string, serverConfig.MessageGiveawayWinners)
	winnerNames := make([]string, serverConfig.MessageGiveawayWinners)

	for i := 0; i < serverConfig.MessageGiveawayWinners; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		winnerId := participants[r.Intn(len(participants))]
		member, err := session.GuildMember(guildId, winnerId)
		if err != nil {
			var memberIndex int
			for j, p := range participants {
				if p == winnerId {
					memberIndex = j
					break
				}
			}
			log.Printf("(%s) finishMessageGiveaway#session.GuildMember: %v", guildId, err)
			participants = append(participants[:memberIndex], participants[memberIndex+1:]...)
			i--
			continue
		}
		winnerIds[i] = winnerId
		winnerNames[i] = member.User.Username
		code, err := h.CsrvClient.GetCSRVCode()
		if err != nil {
			log.Printf("(%s) finishMessageGiveaway#csrvClient.GetCSRVCode: %v", guildId, err)
			i--
			continue
		}

		err = messageGiveawayRepoTx.InsertMessageGiveawayWinner(ctx, giveaway.Id, winnerId, code)
		if err != nil {
			log.Printf("(%s) finishMessageGiveaway#MessageGiveawayRepo.InsertMessageGiveawayWinner: %v", guildId, err)
			i--
			continue
		}

		dmEmbed := discord.ConstructWinnerEmbed(code)
		dm, err := session.UserChannelCreate(winnerId)
		if err != nil {
			log.Printf("(%s) finishMessageGiveaway#session.UserChannelCreate: %v", guildId, err)
			continue
		}
		_, err = session.ChannelMessageSendEmbed(dm.ID, dmEmbed)
		if err != nil {
			log.Printf("(%s) finishMessageGiveaway#session.ChannelMessageSendEmbed: %v", guildId, err)
		}

	}
	err = tx.Commit()
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#tx.Commit: %v", guildId, err)
	}

	mainEmbed := discord.ConstructChannelMessageWinnerEmbed(winnerNames)
	message, err := session.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed: mainEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Kliknij tutaj, aby wyświetlić kod",
						Style:    discordgo.SuccessButton,
						CustomID: "msgwinnercode",
						Emoji: discordgo.ComponentEmoji{
							Name: "🎉",
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#session.ChannelMessageSendComplex: %v", guildId, err)
	}

	err = h.MessageGiveawayRepo.UpdateMessageGiveaway(ctx, &giveaway, message.ID)
	if err != nil {
		log.Printf("(%s) finishMessageGiveaway#giveawayRepo.UpdateGiveaway: %v", guildId, err)
	}
	log.Printf("(%s) Message giveaway ended with a winners: %s", guildId, strings.Join(winnerNames, ", "))

}
