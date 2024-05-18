package commands

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type ThxmeCommand struct {
	Name          string
	Description   string
	DMPermission  bool
	GiveawayHours string
	GiveawayRepo  repos.GiveawayRepo
	UserRepo      repos.UserRepo
	ServerRepo    repos.ServerRepo
}

func NewThxmeCommand(giveawayRepo *repos.GiveawayRepo, userRepo *repos.UserRepo, serverRepo *repos.ServerRepo, giveawayHours string) ThxmeCommand {
	return ThxmeCommand{
		Name:          "thxme",
		Description:   "Poproszenie użytkownika o podziękowanie",
		DMPermission:  false,
		GiveawayRepo:  *giveawayRepo,
		UserRepo:      *userRepo,
		ServerRepo:    *serverRepo,
		GiveawayHours: giveawayHours,
	}
}

func (h ThxmeCommand) Register(ctx context.Context, s *discordgo.Session) {
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
				Description: "Użytkownik, którego chcesz poprosić o podziękowanie",
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

func (h ThxmeCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.WithError(err).Error("Could not get guild")
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
			log.WithError(errors.New("could not get selectedUser")).Error("handleThxmeCommand")
			return
		}
	} else {
		selectedUser = data.Options[0].UserValue(s)
	}
	author := i.Member.User
	if author.ID == selectedUser.ID {
		log.Debug("User and author are the same")
		discord.RespondWithMessage(ctx, s, i, "Nie można poprosić o podziękowanie samego siebie!")
		return
	}
	if selectedUser.Bot {
		log.Debug("User is a bot")
		discord.RespondWithMessage(ctx, s, i, "Nie można prosić o podziękowanie bota!")
		return
	}
	isUserBlacklisted, err := h.UserRepo.IsUserBlacklisted(ctx, author.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleThxmeCommand#UserRepo.IsUserBlacklisted")
		return
	}
	if isUserBlacklisted {
		log.Debug("Author is blacklisted")
		discord.RespondWithMessage(ctx, s, i, "Nie możesz poprosić o podziękowanie, gdyż jesteś na czarnej liście!")
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "",
							Style:    discordgo.SuccessButton,
							CustomID: "accept",
							Emoji: &discordgo.ComponentEmoji{
								Name: "✅",
							},
						},
						&discordgo.Button{
							Label:    "",
							Style:    discordgo.DangerButton,
							CustomID: "reject",
							Emoji: &discordgo.ComponentEmoji{
								Name: "⛔",
							},
						},
					},
				},
			},
			Content: fmt.Sprintf("%s, czy chcesz podziękować użytkownikowi %s?", selectedUser.Mention(), author.Username),
		},
	})
	if err != nil {
		log.WithError(err).Error("handleThxmeCommand#InteractionRespond")
		return
	}

	response, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.WithError(err).Error("handleThxmeCommand#InteractionResponse")
		return
	}

	log.Debug("Inserting participant candidate into database")
	err = h.GiveawayRepo.InsertParticipantCandidate(ctx, i.GuildID, guild.Name, author.ID, author.Username, selectedUser.ID, selectedUser.Username, i.ChannelID, response.ID)
	if err != nil {
		log.WithError(err).Error("handleThxmeCommand#GiveawayRepo.InsertParticipantCandidate")
		str := "Coś poszło nie tak przy dodawaniu kandydata do podziękowania :("
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &str,
		})
		return
	}
	log.Infof("%s has requested thx from %s", author.Username, selectedUser.Username)

}
