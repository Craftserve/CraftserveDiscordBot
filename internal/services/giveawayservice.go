package services

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strings"
	"time"
)

type GiveawayService struct {
	CsrvClient                CsrvClient
	ServerRepo                repos.ServerRepo
	GiveawayRepo              repos.GiveawayRepo
	MessageGiveawayRepo       repos.MessageGiveawayRepo
	UnconditionalGiveawayRepo repos.UnconditionalGiveawayRepo
}

func NewGiveawayService(csrvClient *CsrvClient, serverRepo *repos.ServerRepo, giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo, unconditionalGiveawayRepo *repos.UnconditionalGiveawayRepo) *GiveawayService {
	return &GiveawayService{
		CsrvClient:                *csrvClient,
		ServerRepo:                *serverRepo,
		GiveawayRepo:              *giveawayRepo,
		MessageGiveawayRepo:       *messageGiveawayRepo,
		UnconditionalGiveawayRepo: *unconditionalGiveawayRepo,
	}
}

func (h *GiveawayService) FinishGiveaway(ctx context.Context, s *discordgo.Session, guildId string) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	log.Debug("Finishing giveaway for guild")
	giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#h.GiveawayRepo.GetGiveawayForGuild")
		return
	}
	_, err = s.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#s.Guild")
		return
	}

	giveawayChannelId, err := h.ServerRepo.GetMainChannelForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#h.ServerRepo.GetMainChannelForGuild")
		return
	}

	participants, err := h.GiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#h.GiveawayRepo.GetParticipantsForGiveaway")
	}

	if participants == nil || len(participants) == 0 {
		message, err := s.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był w loterii.")
		if err != nil {
			log.WithError(err).Error("FinishGiveaway#s.ChannelMessageSend")
		}
		err = h.GiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID, "", "", "")
		if err != nil {
			log.WithError(err).Error("FinishGiveaway#h.GiveawayRepo.UpdateGiveaway")
		}
		log.Infof("Giveaway ended without any participants.")
		return
	}

	code, err := h.CsrvClient.GetCSRVCode(ctx)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#h.CsrvClient.GetCSRVCode")
		_, err = s.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
		if err != nil {
			log.WithError(err).Error("FinishGiveaway#s.ChannelMessageSend")
			return
		}
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	winner := participants[r.Intn(len(participants))]

	member, err := s.GuildMember(guildId, winner.UserId)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#s.GuildMember")
		return
	}
	dmEmbed := discord.ConstructWinnerEmbed(code)
	dm, err := s.UserChannelCreate(winner.UserId)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#s.UserChannelCreate")
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
						Emoji: &discordgo.ComponentEmoji{
							Name: "🎉",
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.WithError(err).Error("FinishGiveaway#s.ChannelMessageSendComplex")
	}

	log.Debug("Updating giveaway with winner, message and code")
	err = h.GiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID, code, winner.UserId, member.User.Username)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#h.GiveawayRepo.UpdateGiveaway")
	}
	log.Infof("Giveaway ended with a winner: %s", member.User.Username)

}

func (h *GiveawayService) FinishGiveaways(ctx context.Context, s *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx)
	giveaways, err := h.GiveawayRepo.GetUnfinishedGiveaways(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		log.WithError(err).Error("FinishGiveaways#h.GiveawayRepo.GetUnfinishedGiveaways")
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
	log := logger.GetLoggerFromContext(ctx).WithGuild(guild.ID)
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guild.ID)
	if err != nil {
		log.WithError(err).Error("CreateMissingGiveaways#h.ServerRepo.GetServerConfigForGuild")
		return
	}
	giveawayChannelId := serverConfig.MainChannel
	_, err = s.Channel(giveawayChannelId)
	if err != nil {
		log.WithError(err).Error("CreateMissingGiveaways#s.Channel")
		return
	}
	_, err = h.GiveawayRepo.GetGiveawayForGuild(ctx, guild.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).Error("CreateMissingGiveaways#h.GiveawayRepo.GetGiveawayForGuild")
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		log.Debug("Giveaway for guild does not exist, creating...")
		err = h.GiveawayRepo.InsertGiveaway(ctx, guild.ID, guild.Name)
		if err != nil {
			log.WithError(err).Error("CreateMissingGiveaways#h.GiveawayRepo.InsertGiveaway")
			return
		}
	}

}

func (h *GiveawayService) FinishMessageGiveaways(ctx context.Context, session *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx)
	guildIds, err := h.ServerRepo.GetGuildsWithMessageGiveawaysEnabled(ctx)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaways#serverRepo.GetGuildsWithMessageGiveawaysEnabled")
		return
	}

	for _, guildId := range guildIds {
		_, err := session.Guild(guildId)
		if err != nil {
			log.WithError(err).Error("FinishMessageGiveaways#session.Guild")
			continue
		}
		h.FinishMessageGiveaway(ctx, session, guildId)

	}
}

func (h *GiveawayService) FinishMessageGiveaway(ctx context.Context, session *discordgo.Session, guildId string) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	log.Debug("Finishing message giveaway for guild")
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#serverRepo.GetServerConfigForGuild")
		return
	}

	if serverConfig.MessageGiveawayWinners == 0 {
		return
	}

	giveawayChannelId, err := h.ServerRepo.GetMainChannelForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#serverRepo.GetMainChannelForGuild")
		return
	}

	participants, err := h.MessageGiveawayRepo.GetUsersWithMessagesFromLastDays(ctx, 30, guildId)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#messageGiveawayRepo.GetUsersWithMessagesFromLastDays")
		return
	}

	if len(participants) == 0 {
		_, err := session.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był aktywny.")
		if err != nil {
			log.WithError(err).Error("FinishMessageGiveaway#session.ChannelMessageSend")
		}
		log.Infof("Message giveaway ended without any participants.")
		return
	}

	messageGiveawayRepoTx, tx, err := h.MessageGiveawayRepo.WithTx(ctx, nil)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#messageGiveawayRepo.WithTx")
		return
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	log.Debug("Inserting message giveaway into database")
	err = messageGiveawayRepoTx.InsertMessageGiveaway(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#messageGiveawayRepo.InsertMessageGiveaway")
		return
	}

	giveaway, err := messageGiveawayRepoTx.GetMessageGiveaway(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#messageGiveawayRepo.GetMessageGiveaway")
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
			log.WithError(err).Error("FinishMessageGiveaway#session.GuildMember")
			participants = append(participants[:memberIndex], participants[memberIndex+1:]...)
			i--
			continue
		}
		winnerIds[i] = winnerId
		winnerNames[i] = member.User.Username
		code, err := h.CsrvClient.GetCSRVCode(ctx)
		if err != nil {
			log.WithError(err).Error("FinishMessageGiveaway#csrvClient.GetCSRVCode")
			_, err = session.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
			if err != nil {
				log.WithError(err).Error("FinishMessageGiveaway#s.ChannelMessageSend")
				return
			}
			return
		}

		log.Debug("Inserting message giveaway winner into database")
		err = messageGiveawayRepoTx.InsertMessageGiveawayWinner(ctx, giveaway.Id, winnerId, code)
		if err != nil {
			log.WithError(err).Error("FinishMessageGiveaway#messageGiveawayRepo.InsertMessageGiveawayWinner")
			continue
		}

		dmEmbed := discord.ConstructWinnerEmbed(code)
		dm, err := session.UserChannelCreate(winnerId)
		if err != nil {
			log.WithError(err).Error("FinishMessageGiveaway#session.UserChannelCreate")
			continue
		}
		_, err = session.ChannelMessageSendEmbed(dm.ID, dmEmbed)
		if err != nil {
			log.WithError(err).Error("FinishMessageGiveaway#session.ChannelMessageSendEmbed")
		}

	}
	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#tx.Commit")
		return
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
						Emoji: &discordgo.ComponentEmoji{
							Name: "🎉",
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#session.ChannelMessageSendComplex")
	}

	log.Debug("Updating message giveaway in database with message id")
	err = h.MessageGiveawayRepo.UpdateMessageGiveaway(ctx, &giveaway, message.ID)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#messageGiveawayRepo.UpdateMessageGiveaway")
	}
	log.Infof("Message giveaway ended with a winners: %s", strings.Join(winnerNames, ", "))

}

func (h *GiveawayService) FinishUnconditionalGiveaway(ctx context.Context, session *discordgo.Session, guildId string) {
	log := logger.GetLoggerFromContext(ctx)
	log.Debug("Finishing unconditional giveaway for guild")

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#h.ServerRepo.GetServerConfigForGuild")
		return
	}

	if serverConfig.UnconditionalGiveawayWinners == 0 {
		log.Debug("Unconditional giveaway winners is set to 0, skipping...")
		return
	}

	giveawayChannelId := serverConfig.UnconditionalGiveawayChannel
	_, err = session.Channel(giveawayChannelId)
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#session.Channel")
		return
	}

	giveaway, err := h.UnconditionalGiveawayRepo.GetGiveawayForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.GetGiveawayForGuild")
		return
	}

	participants, err := h.UnconditionalGiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.GetParticipantsForGiveaway")
		return
	}

	if len(participants) == 0 {
		message, err := session.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był w loterii.")
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageSend")
		}

		log.Infof("Unconditional giveaway ended without any participants.")
		err = h.UnconditionalGiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.UpdateGiveaway")
		}

		return
	}

	if len(participants) < serverConfig.UnconditionalGiveawayWinners {
		log.Debug("Not enough participants for the giveaway, ending with less winners")
		message, err := session.ChannelMessageSend(giveawayChannelId, "Za mało uczestników, nie można wyłonić wszystkich zwycięzców.")
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageSend")
		}

		log.Infof("Unconditional giveaway ended without any winners.")
		err = h.UnconditionalGiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.UpdateGiveaway")
		}

		return
	}

	winnerIds := make([]string, serverConfig.UnconditionalGiveawayWinners)
	winnerNames := make([]string, serverConfig.UnconditionalGiveawayWinners)

	for i := 0; i < serverConfig.UnconditionalGiveawayWinners; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		winner := participants[r.Intn(len(participants))]
		member, err := session.GuildMember(guildId, winner.UserId)
		if err != nil {
			var memberIndex int
			for j, p := range participants {
				if p.UserId == winner.UserId {
					memberIndex = j
					break
				}
			}

			log.WithError(err).Error("FinishUnconditionalGiveaway#session.GuildMember")
			participants = append(participants[:memberIndex], participants[memberIndex+1:]...)
			i--
			continue
		}

		hasWon, err := h.UnconditionalGiveawayRepo.HasWonGiveawayByMessageId(ctx, giveaway.InfoMessageId, winner.UserId)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.HasWonGiveawayByMessageId")
			continue
		}

		if hasWon {
			log.WithField("userId", winner.UserId).Debug("User has already won the giveaway, rolling again")
			i--
			continue
		}

		winnerIds[i] = winner.UserId
		winnerNames[i] = member.User.Username
		code, err := h.CsrvClient.GetCSRVCode(ctx)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#h.CsrvClient.GetCSRVCode")
			_, err = session.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
			if err != nil {
				log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageSend")
				return
			}

			return
		}

		log.Debug("Inserting unconditional giveaway winner into database")
		err = h.UnconditionalGiveawayRepo.InsertWinner(ctx, giveaway.Id, winner.UserId, code)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.InsertWinner")
			continue
		}

		log.Debug("Sending DM to unconditional giveaway winner")

		dmEmbed := discord.ConstructWinnerEmbed(code)
		dm, err := session.UserChannelCreate(winner.UserId)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#session.UserChannelCreate")
			continue
		}

		_, err = session.ChannelMessageSendEmbed(dm.ID, dmEmbed)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageSendEmbed")
		}
	}

	winnersEmbed := discord.ConstructUnconditionalGiveawayWinnersEmbed(winnerIds)
	message, err := session.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed: winnersEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Kliknij tutaj, aby wyświetlić kod",
						Style:    discordgo.SuccessButton,
						CustomID: "unconditionalwinnercode",
						Emoji: &discordgo.ComponentEmoji{
							Name: "🎉",
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageSendComplex")
	}

	log.Debug("Updating unconditional giveaway in database with winners")
	err = h.UnconditionalGiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID)
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.UpdateGiveaway")
	}

	log.Infof("Unconditional giveaway ended with winners: %s", strings.Join(winnerNames, ", "))
}

func (h *GiveawayService) FinishUnconditionalGiveaways(ctx context.Context, s *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx)
	giveaways, err := h.UnconditionalGiveawayRepo.GetUnfinishedGiveaways(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		log.WithError(err).Error("FinishUnconditionalGiveaways#h.UnconditionalGiveawayRepo.GetUnfinishedGiveaways")
		return
	}

	for _, giveaway := range giveaways {
		h.FinishUnconditionalGiveaway(ctx, s, giveaway.GuildId)
		guild, err := s.Guild(giveaway.GuildId)
		if err == nil {
			h.CreateUnconditionalGiveaway(ctx, s, guild)
		} else {
			log.WithError(err).Error("FinishUnconditionalGiveaways#s.Guild")
		}
	}
}

func (h *GiveawayService) CreateUnconditionalGiveaway(ctx context.Context, s *discordgo.Session, guild *discordgo.Guild) {
	log := logger.GetLoggerFromContext(ctx)
	log.Info("Creating unconditional giveaway for guild")

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guild.ID)
	if err != nil {
		log.WithError(err).Error("CreateUnconditionalGiveaway#h.ServerRepo.GetServerConfigForGuild")
		return
	}

	giveawayChannelId := serverConfig.UnconditionalGiveawayChannel
	_, err = s.Channel(giveawayChannelId)
	if err != nil {
		log.WithError(err).Error("CreateUnconditionalGiveaway#s.Channel")
		return
	}

	// Check if there is already an active giveaway
	_, err = h.UnconditionalGiveawayRepo.GetGiveawayForGuild(ctx, guild.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).Error("CreateUnconditionalGiveaway#h.UnconditionalGiveawayRepo.GetGiveawayForGuild")
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		log.Debug("Unconditional giveaway for guild does not exist, creating...")

		mainEmbed := discord.ConstructUnconditionalGiveawayJoinEmbed(make([]string, 0))
		message, err := s.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
			Embed: mainEmbed,
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "Weź udział",
							Style:    discordgo.SuccessButton,
							CustomID: "unconditionalgiveawayjoin",
							Emoji: &discordgo.ComponentEmoji{
								Name: "🙋",
							},
						},
					},
				},
			},
		})

		err = h.UnconditionalGiveawayRepo.InsertGiveaway(ctx, guild.ID, message.ID)
		if err != nil {
			log.WithError(err).Error("CreateUnconditionalGiveaway#h.UnconditionalGiveawayRepo.InsertGiveaway")
			return
		}
	}
}
