package services

import (
	"context"
	"csrvbot/domain/entities"
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
	CsrvClient           CsrvClient
	CraftserveUrl        string
	ServerRepo           repos.ServerRepo
	GiveawayRepo         repos.GiveawayRepo
	MessageGiveawayRepo  repos.MessageGiveawayRepo
	JoinableGiveawayRepo entities.JoinableGiveawayRepo
}

func NewGiveawayService(csrvClient *CsrvClient, craftserveUrl string, serverRepo *repos.ServerRepo, giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo, joinableGiveawayRepo entities.JoinableGiveawayRepo) *GiveawayService {
	return &GiveawayService{
		CsrvClient:           *csrvClient,
		CraftserveUrl:        craftserveUrl,
		ServerRepo:           *serverRepo,
		GiveawayRepo:         *giveawayRepo,
		MessageGiveawayRepo:  *messageGiveawayRepo,
		JoinableGiveawayRepo: joinableGiveawayRepo,
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
		err = h.GiveawayRepo.FinishGiveaway(ctx, &giveaway, message.ID, "", "", "")
		if err != nil {
			log.WithError(err).Error("FinishGiveaway#h.GiveawayRepo.FinishGiveaway")
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
	if err != nil && !discord.EqualError(err, discordgo.ErrCodeCannotSendMessagesToThisUser) {
		log.WithError(err).Error("FinishGiveaway#s.ChannelMessageSendEmbed")
	}

	mainEmbed := discord.ConstructChannelWinnerEmbed(h.CraftserveUrl, member.User.Username)
	message, err := s.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed:      mainEmbed,
		Components: discord.ConstructThxWinnerComponents(false),
	})

	if err != nil {
		log.WithError(err).Error("FinishGiveaway#s.ChannelMessageSendComplex")
	}

	log.Debug("Updating giveaway with winner, message and code")
	err = h.GiveawayRepo.FinishGiveaway(ctx, &giveaway, message.ID, code, winner.UserId, member.User.Username)
	if err != nil {
		log.WithError(err).Error("FinishGiveaway#h.GiveawayRepo.FinishGiveaway")
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
		if err != nil && !discord.EqualError(err, discordgo.ErrCodeCannotSendMessagesToThisUser) {
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
	err = h.MessageGiveawayRepo.FinishMessageGiveaway(ctx, &giveaway, message.ID)
	if err != nil {
		log.WithError(err).Error("FinishMessageGiveaway#messageGiveawayRepo.FinishMessageGiveaway")
	}
	log.Infof("Message giveaway ended with a winners: %s", strings.Join(winnerNames, ", "))

}

func (h *GiveawayService) FinishJoinableGiveaway(ctx context.Context, session *discordgo.Session, guildId string, withLevel bool) {
	log := logger.GetLoggerFromContext(ctx).WithField("withLevel", withLevel)
	log.Debug("Finishing joinable giveaway for guild")

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("FinishJoinableGiveaway#h.ServerRepo.GetServerConfigForGuild")
		return
	}

	guild, err := session.Guild(guildId)
	if err != nil {
		log.WithError(err).Error("FinishJoinableGiveaway#session.Guild")
		return
	}

	var winnersCount int
	var channelId string
	var giveaway *entities.JoinableGiveaway

	if withLevel {
		winnersCount = serverConfig.ConditionalGiveawayWinners
		channelId = serverConfig.ConditionalGiveawayChannel
		giveaway, err = h.JoinableGiveawayRepo.GetGiveawayForGuild(ctx, guildId, true)
	} else {
		winnersCount = serverConfig.UnconditionalGiveawayWinners
		channelId = serverConfig.UnconditionalGiveawayChannel
		giveaway, err = h.JoinableGiveawayRepo.GetGiveawayForGuild(ctx, guildId, false)
	}

	// Handle error from GetGiveawayForGuild
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("Joinable giveaway for guild does not exist, creating...")

			if withLevel {
				level, err := discord.PickLevelForGiveaway(ctx, h.ServerRepo, guildId)
				if err != nil {
					log.WithError(err).Error("FinishJoinableGiveaway#discord.PickLevelForGiveaway")
					return
				}

				h.CreateJoinableGiveaway(ctx, session, guild, level)
				return
			}

			h.CreateJoinableGiveaway(ctx, session, guild, nil)
			return
		}

		log.WithError(err).Error("FinishJoinableGiveaway#h.JoinableGiveawayRepo.GetGiveawayForGuild")
		return
	}

	// Verify winnersCount count
	if winnersCount == 0 {
		log.Debug("Winners count is set to 0, skipping...")
		return
	}

	// Verify channel
	_, err = session.Channel(channelId)
	if err != nil {
		log.WithError(err).Error("FinishJoinableGiveaway#session.Channel")
		return
	}

	// Get participants
	participants, err := h.JoinableGiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("FinishJoinableGiveaway#h.JoinableGiveawayRepo.GetParticipantsForGiveaway")
		return
	}

	// Check if there are any participants
	if len(participants) == 0 || len(participants) < winnersCount {
		// Disable join button
		var embed *discordgo.MessageEmbed

		if withLevel {
			levelRole, err := discord.GetRoleForLevel(ctx, session, guildId, *giveaway.Level)
			if err != nil {
				log.WithError(err).Error("FinishJoinableGiveaway#discord.GetRoleForLevel")
				return
			}

			embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, len(participants), &levelRole.ID)
		} else {
			embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, len(participants), nil)
		}

		components := discord.ConstructJoinComponents(true)
		_, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    channelId,
			ID:         giveaway.InfoMessageId,
			Embed:      embed,
			Components: &components,
		})
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#session.ChannelMessageEditComplex")
		}

		message, err := session.ChannelMessageSend(channelId, "Nie można wyłonić wszystkich zwycięzców z powodu zbyt małej ilości uczestników.")
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#session.ChannelMessageSend")
		}

		log.Infof("Joinable giveaway ended without any winnersCount.")
		err = h.JoinableGiveawayRepo.FinishGiveaway(ctx, giveaway, message.ID)
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#h.JoinableGiveawayRepo.FinishGiveaway")
		}

		// Create new joinable giveaway
		log.Info("Creating missing joinable giveaways")
		if withLevel {
			level, err := discord.PickLevelForGiveaway(ctx, h.ServerRepo, guildId)
			if err != nil {
				log.WithError(err).Error("FinishJoinableGiveaway#discord.PickLevelForGiveaway")
				return
			}

			h.CreateJoinableGiveaway(ctx, session, guild, level)
		} else {
			h.CreateJoinableGiveaway(ctx, session, guild, nil)
		}

		return
	}

	winnerIds := make([]string, winnersCount)

	for i := 0; i < winnersCount; i++ {
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

			log.WithError(err).Error("FinishJoinableGiveaway#session.GuildMember")
			participants = append(participants[:memberIndex], participants[memberIndex+1:]...)
			i--
			continue
		}

		hasWon, err := h.JoinableGiveawayRepo.HasWonGiveawayByMessageId(ctx, giveaway.InfoMessageId, winner.UserId)
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#h.JoinableGiveawayRepo.HasWonGiveawayByMessageId")
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
			log.WithError(err).Error("FinishJoinableGiveaway#h.CsrvClient.GetCSRVCode")
			_, err = session.ChannelMessageSend(channelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
			if err != nil {
				log.WithError(err).Error("FinishJoinableGiveaway#session.ChannelMessageSend")
				return
			}

			return
		}

		log.Debug("Inserting joinable giveaway winner into database")
		err = h.JoinableGiveawayRepo.InsertWinner(ctx, giveaway.Id, winner.UserId, code)
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#h.JoinableGiveawayRepo.InsertWinner")
			continue
		}

		log.Debug("Sending DM to joinable giveaway winner")

		dmEmbed := discord.ConstructWinnerEmbed(h.CraftserveUrl, code)
		dm, err := session.UserChannelCreate(winner.UserId)
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#session.UserChannelCreate")
			continue
		}

		_, err = session.ChannelMessageSendEmbed(dm.ID, dmEmbed)
		if err != nil && !discord.EqualError(err, discordgo.ErrCodeCannotSendMessagesToThisUser) {
			log.WithError(err).Error("FinishJoinableGiveaway#session.ChannelMessageSendEmbed")
		}
	}

	// Disable join button
	var embed *discordgo.MessageEmbed

	if withLevel {
		levelRole, err := discord.GetRoleForLevel(ctx, session, guildId, *giveaway.Level)
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#discord.GetRoleForLevel")
			return
		}

		embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, len(participants), &levelRole.ID)
	} else {
		embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, len(participants), nil)
	}

	components := discord.ConstructJoinComponents(true)
	_, err = session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    channelId,
		ID:         giveaway.InfoMessageId,
		Embed:      embed,
		Components: &components,
	})
	if err != nil {
		log.WithError(err).Error("FinishJoinableGiveaway#session.ChannelMessageEditComplex")
	}

	// Send winners message
	var winnersEmbed *discordgo.MessageEmbed
	if withLevel {
		levelRole, err := discord.GetRoleForLevel(ctx, session, guildId, *giveaway.Level)
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#discord.GetRoleForLevel")
			return
		}

		winnersEmbed = discord.ConstructJoinableWinnersEmbed(h.CraftserveUrl, winnerIds, &levelRole.ID)
	} else {
		winnersEmbed = discord.ConstructJoinableWinnersEmbed(h.CraftserveUrl, winnerIds, nil)
	}

	message, err := session.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
		Embed:      winnersEmbed,
		Components: discord.ConstructJoinableGiveawayWinnerComponents(false),
	})
	if err != nil {
		log.WithError(err).Error("FinishJoinableGiveaway#session.ChannelMessageSendComplex")
	}

	log.Debug("Updating joinable giveaway in database with winners")
	err = h.JoinableGiveawayRepo.FinishGiveaway(ctx, giveaway, message.ID)
	if err != nil {
		log.WithError(err).Error("FinishJoinableGiveaway#h.JoinableGiveawayRepo.FinishGiveaway")
	}

	log.Infof("Joinable giveaway ended with winners: %s", strings.Join(winnerIds, ", "))

	// Create new joinable giveaway
	log.Info("Creating missing joinable giveaways")
	if withLevel {
		level, err := discord.PickLevelForGiveaway(ctx, h.ServerRepo, guildId)
		if err != nil {
			log.WithError(err).Error("FinishJoinableGiveaway#discord.PickLevelForGiveaway")
			return
		}

		h.CreateJoinableGiveaway(ctx, session, guild, level)
	} else {
		h.CreateJoinableGiveaway(ctx, session, guild, nil)
	}
}

func (h *GiveawayService) FinishJoinableGiveaways(ctx context.Context, session *discordgo.Session, withLevel bool) {
	log := logger.GetLoggerFromContext(ctx).WithField("withLevel", withLevel)
	giveaways, err := h.JoinableGiveawayRepo.GetUnfinishedGiveaways(ctx, withLevel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		log.WithError(err).Error("FinishJoinableGiveaways#h.JoinableGiveawayRepo.GetUnfinishedGiveaways")
		return
	}

	for _, giveaway := range giveaways {
		h.FinishJoinableGiveaway(ctx, session, giveaway.GuildId, withLevel)
	}
}

func (h *GiveawayService) CreateJoinableGiveaway(ctx context.Context, session *discordgo.Session, guild *discordgo.Guild, level *int) {
	log := logger.GetLoggerFromContext(ctx)
	log.Info("Creating joinable giveaway for guild")

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, guild.ID)
	if err != nil {
		log.WithError(err).Error("CreateJoinableGiveaway#h.ServerRepo.GetServerConfigForGuild")
		return
	}

	if level == nil {
		log.Debug("Checking if giveaway without level for guild exists")
		_, err = h.JoinableGiveawayRepo.GetGiveawayForGuild(ctx, guild.ID, false)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Error("CreateJoinableGiveaway#h.JoinableGiveawayRepo.GetGiveawayForGuild")
			return
		}
	} else {
		log.Debug("Checking if giveaway with level for guild exists")
		_, err = h.JoinableGiveawayRepo.GetGiveawayForGuild(ctx, guild.ID, true)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Error("CreateJoinableGiveaway#h.JoinableGiveawayRepo.GetGiveawayForGuild")
			return
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		log.Debug("Joinable giveaway for guild does not exist, creating...")

		var channelId string
		var embed *discordgo.MessageEmbed

		if level == nil {
			log.Debug("Sending info for unconditional giveaway")
			channelId = serverConfig.UnconditionalGiveawayChannel
			embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, 0, nil)
		} else {
			log.Debug("Sending info for conditional giveaway")
			channelId = serverConfig.ConditionalGiveawayChannel

			levelRole, err := discord.GetRoleForLevel(ctx, session, guild.ID, *level)
			if err != nil {
				log.WithError(err).Error("CreateJoinableGiveaway#discord.GetRoleForLevel")
				return
			}

			embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, 0, &levelRole.ID)
		}

		message, err := session.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
			Embed:      embed,
			Components: discord.ConstructJoinComponents(false),
		})

		err = h.JoinableGiveawayRepo.InsertGiveaway(ctx, guild.ID, message.ID, level)
		if err != nil {
			log.WithError(err).Error("CreateJoinableGiveaway#h.JoinableGiveawayRepo.InsertGiveaway")
			return
		}
	}
}
