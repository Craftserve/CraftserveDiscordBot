package commands

import (
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"log"
)

type GiveawayCommand struct {
	Name          string
	Description   string
	DMPermission  bool
	GiveawayHours string
	GiveawayRepo  repos.GiveawayRepo
}

func NewGiveawayCommand(giveawayRepo *repos.GiveawayRepo, giveawayHours string) GiveawayCommand {
	return GiveawayCommand{
		Name:          "giveaway",
		Description:   "Wyświetla zasady giveawaya",
		DMPermission:  false,
		GiveawayRepo:  *giveawayRepo,
		GiveawayHours: giveawayHours,
	}
}

func (h GiveawayCommand) Register(s *discordgo.Session) {
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
	})
	if err != nil {
		log.Println("Could not register command", err)
	}
}

func (h GiveawayCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := pkg.CreateContext()
	giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, i.GuildID)
	if err != nil {
		log.Printf("(%s) Could not get giveaway: %s", i.GuildID, err)
		return
	}
	participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.Printf("(%s) Could not get participants: %s", i.GuildID, err)
		return
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				discord.ConstructInfoEmbed(participants, h.GiveawayHours),
			},
		},
	})
	if err != nil {
		log.Printf("(%s) Could not respond to interaction (%s): %v", i.GuildID, i.ID, err)
		return
	}
}
