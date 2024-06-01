package commands

import (
	"context"
	"csrvbot/domain/entities"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type GiveawayCommand struct {
	Name          string
	Description   string
	DMPermission  bool
	GiveawayHours string
	GiveawayRepo  entities.GiveawayRepo
	CraftserveUrl string
}

func NewGiveawayCommand(giveawayRepo entities.GiveawayRepo, giveawayHours, craftserveUrl string) GiveawayCommand {
	return GiveawayCommand{
		Name:          "giveaway",
		Description:   "Wyświetla zasady giveawaya",
		DMPermission:  false,
		GiveawayRepo:  giveawayRepo,
		GiveawayHours: giveawayHours,
		CraftserveUrl: craftserveUrl,
	}
}

func (h GiveawayCommand) Register(ctx context.Context, s *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)
	log.Debug("Registering command")
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
	})
	if err != nil {
		log.WithError(err).Error("Could not register command")
	}
}

func (h GiveawayCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("Could not get giveaway")
		return
	}
	participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("Could not get participants")
		return
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				discord.ConstructInfoEmbed(h.CraftserveUrl, participants, h.GiveawayHours),
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not respond to interaction")
		return
	}
}
