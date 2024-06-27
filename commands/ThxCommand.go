package commands

import (
	"context"
	"csrvbot/domain/entities"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
)

type ThxCommand struct {
	Name          string
	Description   string
	DMPermission  bool
	GiveawayHours string
	CraftserveUrl string
	GiveawaysRepo entities.GiveawaysRepo
	UserRepo      entities.UserRepo
	ServerRepo    entities.ServerRepo
}

func NewThxCommand(giveawaysRepo entities.GiveawaysRepo, userRepo entities.UserRepo, serverRepo entities.ServerRepo, giveawayHours, craftserveUrl string) ThxCommand {
	return ThxCommand{
		Name:          "thx",
		Description:   "Podziękowanie innemu użytkownikowi",
		DMPermission:  false,
		GiveawaysRepo: giveawaysRepo,
		UserRepo:      userRepo,
		ServerRepo:    serverRepo,
		GiveawayHours: giveawayHours,
		CraftserveUrl: craftserveUrl,
	}
}

func (h ThxCommand) Register(ctx context.Context, s *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)
	log.Debug("Registering command")
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Użytkownik, któremu chcesz podziękować",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not register command")
	}

	log.Debug("Registering message context command")
	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		DMPermission: &h.DMPermission,
		Type:         discordgo.MessageApplicationCommand,
	})
	if err != nil {
		log.WithError(err).Error("Could not register message context command")
	}

	log.Debug("Registering user context command")
	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		DMPermission: &h.DMPermission,
		Type:         discordgo.UserApplicationCommand,
	})
	if err != nil {
		log.WithError(err).Error("Could not register user context command")
	}
}

func (h ThxCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#session.Guild")
		return
	}
	var selectedUser *discordgo.User
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		if len(data.Resolved.Messages) != 0 {
			selectedUser = data.Resolved.Messages[data.TargetID].Author
		} else if len(data.Resolved.Users) != 0 {
			selectedUser = data.Resolved.Users[data.TargetID]
		} else {
			log.WithError(errors.New("could not get selectedUser")).Error("handleThxCommand#")
			return
		}
	} else {
		selectedUser = data.Options[0].UserValue(s)
	}
	author := i.Member.User
	if author.ID == selectedUser.ID {
		log.Debug("User and author are the same")
		discord.RespondWithMessage(ctx, s, i, "Nie można dziękować sobie!")
		return
	}
	if selectedUser.Bot {
		log.Debug("User is a bot")
		discord.RespondWithMessage(ctx, s, i, "Nie można dziękować botom!")
		return
	}
	isUserBlacklisted, err := h.UserRepo.IsUserBlacklisted(ctx, selectedUser.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#UserRepo.IsUserBlacklisted")
		return
	}
	if isUserBlacklisted {
		log.Debug("User is blacklisted")
		discord.RespondWithMessage(ctx, s, i, "Ten użytkownik jest na czarnej liście i nie może brać udziału :(")
		return
	}
	giveaway, err := h.GiveawaysRepo.GetGiveawayForGuild(ctx, i.GuildID, entities.ThxGiveawayType)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#GiveawaysRepo.GetGiveawayForGuild")
		return
	}
	participants, err := h.GiveawaysRepo.GetParticipantsForGiveaway(ctx, giveaway.Id, nil)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#GiveawaysRepo.GetParticipantNamesForGiveaway")
		return
	}

	var participantsNames []string
	for _, participant := range participants {
		participantsNames = append(participantsNames, participant.UserName)
	}

	embed := discord.ConstructThxEmbed(h.CraftserveUrl, participantsNames, h.GiveawayHours, selectedUser.ID, "", "wait")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: discord.ConstructAcceptRejectComponents(false),
			Embeds:     []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.WithError(err).Error("handleThxCommand#session.InteractionRespond")
		return
	}

	response, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#session.InteractionResponse")
		return
	}

	log.Debug("Inserting participant into database")
	level, err := discord.GetMemberLevel(ctx, s, i.Member, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#discord.GetMemberLevel")
		return
	}

	err = h.GiveawaysRepo.InsertParticipant(ctx, giveaway.Id, level, guild.ID, selectedUser.ID, selectedUser.Username, &response.ID, &i.ChannelID)
	//err = h.GiveawaysRepo.InsertParticipant(ctx, giveaway.Id, i.GuildID, guild.Name, selectedUser.ID, selectedUser.Username, i.ChannelID, response.ID)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#GiveawaysRepo.InsertParticipant")
		str := "Coś poszło nie tak przy dodawaniu podziękowania :("
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &str,
		})
		if err != nil {
			log.WithError(err).Error("handleThxCommand#session.InteractionResponseEdit")
		}
		return
	}
	log.Infof("%s has thanked %s", author.Username, selectedUser.Username)

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleThxCommand#ServerRepo.GetServerConfigForGuild")
		return
	}

	thxNotification, err := h.GiveawaysRepo.GetThxNotification(ctx, response.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).Error("handleThxCommand#GiveawaysRepo.GetThxNotification")
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, response.ID, selectedUser.ID, "", "wait", h.CraftserveUrl)
		if err != nil {
			log.WithError(err).Error("handleThxCommand#discord.NotifyThxOnThxInfoChannel")
			return
		}

		log.Debug("Inserting notification into database")
		err = h.GiveawaysRepo.InsertThxNotification(ctx, response.ID, notificationMessageId)
		if err != nil {
			log.WithError(err).Error("handleThxCommand#GiveawaysRepo.InsertThxNotification")
			return
		}
	} else {
		_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, response.ID, selectedUser.ID, "", "wait", h.CraftserveUrl)
		if err != nil {
			log.WithError(err).Error("handleThxCommand#discord.NotifyThxOnThxInfoChannel")
			return
		}
	}

}
