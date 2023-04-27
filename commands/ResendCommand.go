package commands

import (
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"log"
)

type ResendCommand struct {
	Name                string
	Description         string
	DMPermission        bool
	GiveawayRepo        repos.GiveawayRepo
	MessageGiveawayRepo repos.MessageGiveawayRepo
}

func NewResendCommand(giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo) ResendCommand {
	return ResendCommand{
		Name:                "resend",
		Description:         "Wysyła na PW ostatnie 10 wygranych kodów z giveawayi",
		DMPermission:        false,
		GiveawayRepo:        *giveawayRepo,
		MessageGiveawayRepo: *messageGiveawayRepo,
	}
}

func (h ResendCommand) Register(s *discordgo.Session) {
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
	})
	if err != nil {
		log.Println("Could not register command", err)
	}
}

func (h ResendCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := pkg.CreateContext()
	thxCodes, err := h.GiveawayRepo.GetLastCodesForUser(ctx, i.Member.User.ID, 10)
	if err != nil {
		log.Println("ResendCommand#h.GiveawayRepo.GetLastCodesForUser", err)
		return
	}
	msgCodes, err := h.MessageGiveawayRepo.GetLastCodesForUser(ctx, i.Member.User.ID, 10)
	if err != nil {
		log.Println("ResendCommand#h.MessageGiveawayRepo.GetLastCodesForUser", err)
		return
	}
	thxEmbed := discord.ConstructResendEmbed(thxCodes)
	msgEmbed := discord.ConstructResendEmbed(msgCodes)

	dm, err := s.UserChannelCreate(i.Member.User.ID)
	if err != nil {
		log.Println("handleCsrvbotCommand#UserChannelCreate", err)
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
				Embeds:  []*discordgo.MessageEmbed{thxEmbed},
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Println("ResendCommand#session.InteractionRespond", err)
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
		log.Println("ResendCommand#session.InteractionRespond", err)
		return
	}

}
