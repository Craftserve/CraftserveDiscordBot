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
	CraftserveUrl             string
	ServerRepo                repos.ServerRepo
	GiveawayRepo              repos.GiveawayRepo
	MessageGiveawayRepo       repos.MessageGiveawayRepo
	UnconditionalGiveawayRepo repos.UnconditionalGiveawayRepo
	ConditionalGiveawayRepo   repos.ConditionalGiveawayRepo
}

func NewGiveawayService(csrvClient *CsrvClient, craftserveUrl string, serverRepo *repos.ServerRepo, giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo, unconditionalGiveawayRepo *repos.UnconditionalGiveawayRepo, conditionalGiveawayRepo *repos.ConditionalGiveawayRepo) *GiveawayService {
	return &GiveawayService{
		CsrvClient:                *csrvClient,
		CraftserveUrl:             craftserveUrl,
		ServerRepo:                *serverRepo,
		GiveawayRepo:              *giveawayRepo,
		MessageGiveawayRepo:       *messageGiveawayRepo,
		UnconditionalGiveawayRepo: *unconditionalGiveawayRepo,
		ConditionalGiveawayRepo:   *conditionalGiveawayRepo,
	}
}

func (h *GiveawayService) FinishGiveaway(ctx context.Context, s *discordgo.Session, guildId string) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	log.Debug("Finishing giveaway for guild")

	guild, err := s.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#s.Guild")
		return
	}

	giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, guildId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("Giveaway for guild does not exist, creating...")
			h.CreateMissingGiveaways(ctx, s, guild)
			return
		}

		log.WithError(err).Error("FinishGiveaway#h.GiveawayRepo.GetGiveawayForGuild")
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

		// Create new giveaway
		log.Info("Creating missing giveaways")
		h.CreateMissingGiveaways(ctx, s, guild)

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
	dmEmbed := discord.ConstructWinnerEmbed(h.CraftserveUrl, code)
	dm, err := s.UserChannelCreate(winner.UserId)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#s.UserChannelCreate")
	}
	_, err = s.ChannelMessageSendEmbed(dm.ID, dmEmbed)

	mainEmbed := discord.ConstructChannelWinnerEmbed(h.CraftserveUrl, member.User.Username)
	message, err := s.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed:      mainEmbed,
		Components: discord.ConstructThxWinnerComponents(false),
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

	log.Info("Creating missing giveaways")
	h.CreateMissingGiveaways(ctx, s, guild)
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

		dmEmbed := discord.ConstructWinnerEmbed(h.CraftserveUrl, code)
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

	mainEmbed := discord.ConstructChannelMessageWinnerEmbed(h.CraftserveUrl, winnerNames)
	message, err := session.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed:      mainEmbed,
		Components: discord.ConstructMessageWinnerComponents(false),
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

	guild, err := session.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#session.Guild")
		return
	}

	giveaway, err := h.UnconditionalGiveawayRepo.GetGiveawayForGuild(ctx, guildId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("Unconditional giveaway for guild does not exist, creating...")
			h.CreateUnconditionalGiveaway(ctx, session, guild)
			return
		}
		log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.GetGiveawayForGuild")
		return
	}

	participants, err := h.UnconditionalGiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.GetParticipantsForGiveaway")
		return
	}

	if len(participants) == 0 {
		// Disable join button
		joinEmbed := discord.ConstructUnconditionalGiveawayJoinEmbed(h.CraftserveUrl, 0)
		component := discord.ConstructUnconditionalJoinComponents(true)
		_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    giveawayChannelId,
			ID:         giveaway.InfoMessageId,
			Embed:      joinEmbed,
			Components: &component,
		})
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageEditComplex")
		}

		message, err := session.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był w bezwarunkowej loterii.")
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageSend")
		}

		log.Infof("Unconditional giveaway ended without any participants.")
		err = h.UnconditionalGiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID)
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#h.UnconditionalGiveawayRepo.UpdateGiveaway")
		}

		// Create new unconditional giveaway
		log.Info("Creating missing unconditional giveaways")
		h.CreateUnconditionalGiveaway(ctx, session, guild)

		return
	}

	if len(participants) < serverConfig.UnconditionalGiveawayWinners {
		// Disable join button
		joinEmbed := discord.ConstructUnconditionalGiveawayJoinEmbed(h.CraftserveUrl, 0)
		component := discord.ConstructUnconditionalJoinComponents(true)
		_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    giveawayChannelId,
			ID:         giveaway.InfoMessageId,
			Embed:      joinEmbed,
			Components: &component,
		})
		if err != nil {
			log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageEditComplex")
		}

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

		// Create new unconditional giveaway
		log.Info("Creating missing unconditional giveaways")
		h.CreateUnconditionalGiveaway(ctx, session, guild)

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

		dmEmbed := discord.ConstructWinnerEmbed(h.CraftserveUrl, code)
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

	// Disable join button
	joinEmbed := discord.ConstructUnconditionalGiveawayJoinEmbed(h.CraftserveUrl, len(participants))
	component := discord.ConstructUnconditionalJoinComponents(true)
	_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    giveawayChannelId,
		ID:         giveaway.InfoMessageId,
		Embed:      joinEmbed,
		Components: &component,
	})
	if err != nil {
		log.WithError(err).Error("FinishUnconditionalGiveaway#session.ChannelMessageEditComplex")
	}

	// Send winners message
	winnersEmbed := discord.ConstructUnconditionalGiveawayWinnersEmbed(h.CraftserveUrl, winnerIds)
	message, err := session.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed:      winnersEmbed,
		Components: discord.ConstructUnconditionalWinnerComponents(false),
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

	// Create new unconditional giveaway
	log.Info("Creating missing unconditional giveaways")
	h.CreateUnconditionalGiveaway(ctx, session, guild)
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

		mainEmbed := discord.ConstructUnconditionalGiveawayJoinEmbed(h.CraftserveUrl, 0)
		message, err := s.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
			Embed:      mainEmbed,
			Components: discord.ConstructUnconditionalJoinComponents(false),
		})
		if err != nil {
			log.WithError(err).Error("CreateUnconditionalGiveaway#s.ChannelMessageSendComplex")
			return
		}

		err = h.UnconditionalGiveawayRepo.InsertGiveaway(ctx, guild.ID, message.ID)
		if err != nil {
			log.WithError(err).Error("CreateUnconditionalGiveaway#h.UnconditionalGiveawayRepo.InsertGiveaway")
			return
		}
	}
}

func (h *GiveawayService) FinishConditionalGiveaway(ctx context.Context, session *discordgo.Session, guildId string) {
	log := logger.GetLoggerFromContext(ctx)
	log.Debug("Finishing conditional giveaway for guild")

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#h.ServerRepo.GetServerConfigForGuild")
		return
	}

	if serverConfig.ConditionalGiveawayWinners == 0 {
		log.Debug("Conditional giveaway winners is set to 0, skipping...")
		return
	}

	giveawayChannelId := serverConfig.ConditionalGiveawayChannel
	_, err = session.Channel(giveawayChannelId)
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#session.Channel")
		return
	}

	guild, err := session.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#session.Guild")
		return
	}

	giveaway, err := h.ConditionalGiveawayRepo.GetGiveawayForGuild(ctx, guildId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("Conditional giveaway for guild does not exist, creating...")
			h.CreateConditionalGiveaway(ctx, session, guild)
			return
		}

		log.WithError(err).Error("FinishConditionalGiveaway#h.ConditionalGiveawayRepo.GetGiveawayForGuild")
		return
	}

	participants, err := h.ConditionalGiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#h.ConditionalGiveawayRepo.GetParticipantsForGiveaway")
		return
	}

	levelRole, err := discord.GetRoleForLevel(ctx, session, guildId, giveaway.Level)
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#discord.GetRoleForLevel")
		return
	}

	if len(participants) == 0 {
		// Disable join button
		joinEmbed := discord.ConstructConditionalGiveawayJoinEmbed(h.CraftserveUrl, levelRole.ID, 0)
		component := discord.ConstructConditionalJoinComponents(true)
		_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    giveawayChannelId,
			ID:         giveaway.InfoMessageId,
			Embed:      joinEmbed,
			Components: &component,
		})
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageEditComplex")
		}

		message, err := session.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był w warunkowej loterii.")
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageSend")
		}

		log.Infof("Conditional giveaway ended without any participants.")
		err = h.ConditionalGiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID)
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#h.ConditionalGiveawayRepo.UpdateGiveaway")
		}

		// Create new conditional giveaway
		log.Info("Creating missing conditional giveaways")
		h.CreateConditionalGiveaway(ctx, session, guild)

		return
	}

	if len(participants) < serverConfig.ConditionalGiveawayWinners {
		// Disable join button
		joinEmbed := discord.ConstructConditionalGiveawayJoinEmbed(h.CraftserveUrl, levelRole.ID, 0)
		component := discord.ConstructConditionalJoinComponents(true)
		_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    giveawayChannelId,
			ID:         giveaway.InfoMessageId,
			Embed:      joinEmbed,
			Components: &component,
		})
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageEditComplex")
		}

		log.Debug("Not enough participants for the giveaway, ending with less winners")
		message, err := session.ChannelMessageSend(giveawayChannelId, "Za mało uczestników, nie można wyłonić wszystkich zwycięzców.")
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageSend")
		}

		log.Infof("Conditional giveaway ended without any winners.")
		err = h.ConditionalGiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID)
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#h.ConditionalGiveawayRepo.UpdateGiveaway")
		}

		// Create new conditional giveaway
		log.Info("Creating missing conditional giveaways")
		h.CreateConditionalGiveaway(ctx, session, guild)

		return
	}

	winnerIds := make([]string, serverConfig.ConditionalGiveawayWinners)

	for i := 0; i < serverConfig.ConditionalGiveawayWinners; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		winner := participants[r.Intn(len(participants))]
		_, err := session.GuildMember(guildId, winner.UserId)
		if err != nil {
			var memberIndex int
			for j, p := range participants {
				if p.UserId == winner.UserId {
					memberIndex = j
					break
				}
			}

			log.WithError(err).Error("FinishConditionalGiveaway#session.GuildMember")
			participants = append(participants[:memberIndex], participants[memberIndex+1:]...)
			i--
			continue
		}

		hasWon, err := h.ConditionalGiveawayRepo.HasWonGiveawayByMessageId(ctx, giveaway.InfoMessageId, winner.UserId)
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#h.ConditionalGiveawayRepo.HasWonGiveawayByMessageId")
			continue
		}

		if hasWon {
			log.WithField("userId", winner.UserId).Debug("User has already won the giveaway, rolling again")
			i--
			continue
		}

		winnerIds[i] = winner.UserId
		code, err := h.CsrvClient.GetCSRVCode(ctx)
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#h.CsrvClient.GetCSRVCode")
			_, err = session.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
			if err != nil {
				log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageSend")
				return
			}

			return
		}

		log.Debug("Inserting conditional giveaway winner into database")
		err = h.ConditionalGiveawayRepo.InsertWinner(ctx, giveaway.Id, winner.UserId, code)
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#h.ConditionalGiveawayRepo.InsertWinner")
			continue
		}

		log.Debug("Sending DM to conditional giveaway winner")

		dmEmbed := discord.ConstructWinnerEmbed(h.CraftserveUrl, code)
		dm, err := session.UserChannelCreate(winner.UserId)
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#session.UserChannelCreate")
			continue
		}

		_, err = session.ChannelMessageSendEmbed(dm.ID, dmEmbed)
		if err != nil {
			log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageSendEmbed")
		}
	}

	// Disable join button
	joinEmbed := discord.ConstructConditionalGiveawayJoinEmbed(h.CraftserveUrl, levelRole.ID, len(participants))
	component := discord.ConstructConditionalJoinComponents(true)
	_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    giveawayChannelId,
		ID:         giveaway.InfoMessageId,
		Embed:      joinEmbed,
		Components: &component,
	})
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageEditComplex")
	}

	// Send winners message
	winnersEmbed := discord.ConstructConditionalGiveawayWinnersEmbed(h.CraftserveUrl, levelRole.ID, winnerIds)
	message, err := session.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed:      winnersEmbed,
		Components: discord.ConstructConditionalWinnerComponents(false),
	})
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#session.ChannelMessageSendComplex")
	}

	log.Debug("Updating conditional giveaway in database with winners")
	err = h.ConditionalGiveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID)
	if err != nil {
		log.WithError(err).Error("FinishConditionalGiveaway#h.ConditionalGiveawayRepo.UpdateGiveaway")
	}

	log.Infof("Conditional giveaway ended with winners: %s", strings.Join(winnerIds, ", "))

	// Create new conditional giveaway
	log.Info("Creating missing conditional giveaways")
	h.CreateConditionalGiveaway(ctx, session, guild)
}

func (h *GiveawayService) FinishConditionalGiveaways(ctx context.Context, session *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx)
	giveaways, err := h.ConditionalGiveawayRepo.GetUnfinishedGiveaways(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		log.WithError(err).Error("FinishConditionalGiveaways#h.ConditionalGiveawayRepo.GetUnfinishedGiveaways")
		return
	}

	for _, giveaway := range giveaways {
		h.FinishConditionalGiveaway(ctx, session, giveaway.GuildId)
	}
}

func (h *GiveawayService) CreateConditionalGiveaway(ctx context.Context, session *discordgo.Session, guild *discordgo.Guild) {
	log := logger.GetLoggerFromContext(ctx)
	log.Info("Creating conditional giveaway for guild")

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guild.ID)
	if err != nil {
		log.WithError(err).Error("CreateConditionalGiveaway#h.ServerRepo.GetServerConfigForGuild")
		return
	}

	giveawayChannelId := serverConfig.ConditionalGiveawayChannel
	_, err = session.Channel(giveawayChannelId)
	if err != nil {
		log.WithError(err).Error("CreateConditionalGiveaway#session.Channel")
		return
	}

	_, err = h.ConditionalGiveawayRepo.GetGiveawayForGuild(ctx, guild.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).Error("CreateConditionalGiveaway#h.ConditionalGiveawayRepo.GetGiveawayForGuild")
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		log.Debug("Conditional giveaway for guild does not exist, creating...")

		log.Debug("Getting one of the levels")
		levels, err := h.ServerRepo.GetConditionalGiveawayLevels(ctx, guild.ID)
		if err != nil {
			log.WithError(err).Error("CreateConditionalGiveaway#h.ServerRepo.GetConditionalGiveawayLevels")
			return
		}

		level := levels[rand.Intn(len(levels))]
		levelRole, err := discord.GetRoleForLevel(ctx, session, guild.ID, level)
		if err != nil {
			log.WithError(err).Error("CreateConditionalGiveaway#discord.GetRoleForLevel")
			return
		}

		mainEmbed := discord.ConstructConditionalGiveawayJoinEmbed(h.CraftserveUrl, levelRole.ID, 0)
		message, err := session.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
			Embed:      mainEmbed,
			Components: discord.ConstructConditionalJoinComponents(false),
		})
		if err != nil {
			log.WithError(err).Error("CreateConditionalGiveaway#session.ChannelMessageSendComplex")
			return
		}

		err = h.ConditionalGiveawayRepo.InsertGiveaway(ctx, guild.ID, message.ID, level)
		if err != nil {
			log.WithError(err).Error("CreateConditionalGiveaway#h.ConditionalGiveawayRepo.InsertGiveaway")
			return
		}
	}
}
