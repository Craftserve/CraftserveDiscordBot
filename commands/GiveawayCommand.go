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
	GiveawaysRepo entities.GiveawaysRepo
	CraftserveUrl string
	VoucherValue  int
}

func NewGiveawayCommand(giveawaysRepo entities.GiveawaysRepo, giveawayHours, craftserveUrl string, voucherValue int) GiveawayCommand {
	return GiveawayCommand{
		Name:          "giveaway",
		Description:   "Wyświetla zasady giveawaya",
		DMPermission:  false,
		GiveawaysRepo: giveawaysRepo,
		GiveawayHours: giveawayHours,
		CraftserveUrl: craftserveUrl,
		VoucherValue:  voucherValue,
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
	giveaway, err := h.GiveawaysRepo.GetGiveawayForGuild(ctx, i.GuildID, entities.ThxGiveawayType)
	if err != nil {
		log.WithError(err).Error("Could not get giveaway")
		return
	}
	participants, err := h.GiveawaysRepo.GetParticipantsForGiveaway(ctx, giveaway.Id, nil)
	if err != nil {
		log.WithError(err).Error("Could not get participants")
		return
	}

	var participantsNames []string
	for _, participant := range participants {
		participantsNames = append(participantsNames, participant.UserName)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				discord.ConstructInfoEmbed(h.CraftserveUrl, participantsNames, h.GiveawayHours, h.VoucherValue),
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not respond to interaction")
		return
	}
}
