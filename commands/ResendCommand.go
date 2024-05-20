package commands

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type ResendCommand struct {
	Name                string
	Description         string
	DMPermission        bool
	CraftserveUrl       string
	GiveawayRepo        repos.GiveawayRepo
	MessageGiveawayRepo repos.MessageGiveawayRepo
}

func NewResendCommand(giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo, craftserveUrl string) ResendCommand {
	return ResendCommand{
		Name:                "resend",
		Description:         "Wysyła na PW ostatnie 10 wygranych kodów z giveawayi",
		DMPermission:        false,
		CraftserveUrl:       craftserveUrl,
		GiveawayRepo:        *giveawayRepo,
		MessageGiveawayRepo: *messageGiveawayRepo,
	}
}

func (h ResendCommand) Register(ctx context.Context, s *discordgo.Session) {
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

func (h ResendCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	thxCodes, err := h.GiveawayRepo.GetLastCodesForUser(ctx, i.Member.User.ID, 10)
	if err != nil {
		log.WithError(err).Error("ResendCommand#h.GiveawayRepo.GetLastCodesForUser")
		return
	}
	msgCodes, err := h.MessageGiveawayRepo.GetLastCodesForUser(ctx, i.Member.User.ID, 10)
	if err != nil {
		log.WithError(err).Error("ResendCommand#h.MessageGiveawayRepo.GetLastCodesForUser")
		return
	}
	thxEmbed := discord.ConstructResendEmbed(h.CraftserveUrl, thxCodes)
	msgEmbed := discord.ConstructResendEmbed(h.CraftserveUrl, msgCodes)

	log.Debug("Trying to create DM channel")
	dm, err := s.UserChannelCreate(i.Member.User.ID)
	if err != nil {
		log.WithError(err).Error("ResendCommand#s.UserChannelCreate")
		return
	}

	_, err = s.ChannelMessageSendComplex(dm.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{thxEmbed, msgEmbed},
	})
	if err != nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Nie udało się wysłać kodów, ponieważ masz wyłączone wiadomości prywatne, oto kopia wiadomości:",
				Embeds:  []*discordgo.MessageEmbed{thxEmbed, msgEmbed},
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.WithError(err).Error("ResendCommand#session.InteractionRespond")
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Poprzez prywatną wiadomość, wysłano twoje 10 ostatnich wygranych kodów",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.WithError(err).Error("ResendCommand#session.InteractionRespond")
		return
	}

}
