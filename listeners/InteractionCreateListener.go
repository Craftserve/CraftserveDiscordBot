package listeners

import (
	"context"
	"csrvbot/commands"
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
)

type InteractionCreateListener struct {
	GiveawayCommand commands.GiveawayCommand
	ThxCommand      commands.ThxCommand
	ThxmeCommand    commands.ThxmeCommand
	CsrvbotCommand  commands.CsrvbotCommand
	DocCommand      commands.DocCommand
	ResendCommand   commands.ResendCommand
	GiveawayHours   string
	GiveawayRepo    repos.GiveawayRepo
	ServerRepo      repos.ServerRepo
	HelperService   services.HelperService
}

func NewInteractionCreateListener(giveawayCommand commands.GiveawayCommand, thxCommand commands.ThxCommand, thxmeCommand commands.ThxmeCommand, csrvbotCommand commands.CsrvbotCommand, docCommand commands.DocCommand, resendCommand commands.ResendCommand, giveawayRepo *repos.GiveawayRepo, serverRepo *repos.ServerRepo, helperService *services.HelperService) InteractionCreateListener {
	return InteractionCreateListener{
		GiveawayCommand: giveawayCommand,
		ThxCommand:      thxCommand,
		ThxmeCommand:    thxmeCommand,
		CsrvbotCommand:  csrvbotCommand,
		DocCommand:      docCommand,
		ResendCommand:   resendCommand,
		GiveawayHours:   giveawayCommand.GiveawayHours, // todo: pass this as a parameter
		GiveawayRepo:    *giveawayRepo,
		ServerRepo:      *serverRepo,
		HelperService:   *helperService,
	}
}

func (h InteractionCreateListener) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := pkg.CreateContext()
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleApplicationCommands(s, i)
	case discordgo.InteractionApplicationCommandAutocomplete:
		h.handleApplicationCommandsAutocomplete(s, i)
	case discordgo.InteractionMessageComponent:
		h.handleMessageComponents(ctx, s, i)
	}
}

func (h InteractionCreateListener) handleApplicationCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "giveaway":
		h.GiveawayCommand.Handle(s, i)
	case "thx":
		h.ThxCommand.Handle(s, i)
	case "thxme":
		h.ThxmeCommand.Handle(s, i)
	case "doc":
		h.DocCommand.Handle(s, i)
	case "csrvbot":
		h.CsrvbotCommand.Handle(s, i)
	case "resend":
		h.ResendCommand.Handle(s, i)
	}
}

func (h InteractionCreateListener) handleApplicationCommandsAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "doc":
		h.DocCommand.HandleAutocomplete(s, i)
	}
}

func (h InteractionCreateListener) handleMessageComponents(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.MessageComponentData().CustomID {
	case "winnercode":
		hasWon, err := h.GiveawayRepo.HasWonGiveawayByMessageId(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.Println("("+i.GuildID+") handleMessageComponents#GiveawayRepo.HasWonGiveawayByMessageId", err)
			return
		}
		if !hasWon {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Nie wygrałeś tego giveawayu!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Println("("+i.GuildID+") handleMessageComponents#session.InteractionRespond", err)
			}
			return
		}

		code, err := h.GiveawayRepo.GetCodeForInfoMessage(ctx, i.Message.ID)
		if err != nil {
			log.Println("("+i.GuildID+") handleMessageComponents#session.InteractionRespond", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{discord.ConstructWinnerEmbed(code)},
			},
		})
		if err != nil {
			log.Println("("+i.GuildID+") handleMessageComponents#session.InteractionRespond", err)
			return
		}
	case "thx_accept":
		isThxMessage, err := h.GiveawayRepo.IsThxMessage(ctx, i.Message.ID)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#h.GiveawayRepo.IsThxMessage: %v", i.GuildID, err)
			return
		}
		if !isThxMessage {
			return
		}

		serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
		if err != nil {
			log.Printf("Could not get server config for guild %s", i.GuildID)
			return
		}

		if !discord.HasAdminPermissions(s, i.Member, serverConfig.AdminRoleId, i.GuildID) {
			// TODO discord.RespondWithEphemeralMessage(s, i.Interaction, "Nie masz uprawnień do akceptacji!")
			return
		}

		participant, err := h.GiveawayRepo.GetParticipant(ctx, i.Message.ID)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#h.GiveawayRepo.GetParticipant: %v", i.GuildID, err)
			return
		}

		thxNotification, notificationErr := h.GiveawayRepo.GetThxNotification(ctx, i.Message.ID)
		if notificationErr != nil && !errors.Is(notificationErr, sql.ErrNoRows) {
			log.Printf("Could not get thx notification for message %s", i.Message.ID)
			return
		}

		err = h.GiveawayRepo.UpdateParticipant(ctx, &participant, i.Member.User.ID, i.Member.User.Username, true)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#h.GiveawayRepo.UpdateParticipant: %v", i.GuildID, err)
			return
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Udział użytkownika został potwierdzony!",
			},
		})
		if err != nil {
			log.Printf("(%s) handleMessageComponents#session.InteractionRespond: %v", i.GuildID, err)
			return
		}
		log.Printf("(%s) %s accepted %s participation in giveaway %d", i.GuildID, i.Member.User.Username, participant.UserName, participant.GiveawayId)

		participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, participant.GiveawayId)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#h.GiveawayRepo.GetParticipantNamesForGiveaway: %v", i.GuildID, err)
			return
		}

		embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, participant.UserId, i.Member.User.ID, "confirm")

		_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#session.ChannelMessageEditEmbed: %v", i.GuildID, err)
			return
		}

		if errors.Is(notificationErr, sql.ErrNoRows) {
			notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, i.Member.User.ID, "confirm")
			if err != nil {
				log.Printf("(%s) Could not notify thx on thx info channel: %v", i.GuildID, err)
				return
			}

			err = h.GiveawayRepo.InsertThxNotification(ctx, i.Message.ID, notificationMessageId)
			if err != nil {
				log.Printf("(%s) Could not insert thx notification", i.GuildID)
				return
			}
		} else {
			_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, i.Member.User.ID, "confirm")
			if err != nil {
				log.Printf("(%s) Could not notify thx on thx info channel: %v", i.GuildID, err)
				return
			}
		}

		h.HelperService.CheckHelper(ctx, s, i.GuildID, participant.UserId)

	}
}
