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
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type InteractionCreateListener struct {
	GiveawayCommand     commands.GiveawayCommand
	ThxCommand          commands.ThxCommand
	ThxmeCommand        commands.ThxmeCommand
	CsrvbotCommand      commands.CsrvbotCommand
	DocCommand          commands.DocCommand
	ResendCommand       commands.ResendCommand
	GiveawayHours       string
	GiveawayRepo        repos.GiveawayRepo
	MessageGiveawayRepo repos.MessageGiveawayRepo
	ServerRepo          repos.ServerRepo
	HelperService       services.HelperService
}

func NewInteractionCreateListener(giveawayCommand commands.GiveawayCommand, thxCommand commands.ThxCommand, thxmeCommand commands.ThxmeCommand, csrvbotCommand commands.CsrvbotCommand, docCommand commands.DocCommand, resendCommand commands.ResendCommand, giveawayHours string, giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo, serverRepo *repos.ServerRepo, helperService *services.HelperService) InteractionCreateListener {
	return InteractionCreateListener{
		GiveawayCommand:     giveawayCommand,
		ThxCommand:          thxCommand,
		ThxmeCommand:        thxmeCommand,
		CsrvbotCommand:      csrvbotCommand,
		DocCommand:          docCommand,
		ResendCommand:       resendCommand,
		GiveawayHours:       giveawayHours,
		GiveawayRepo:        *giveawayRepo,
		MessageGiveawayRepo: *messageGiveawayRepo,
		ServerRepo:          *serverRepo,
		HelperService:       *helperService,
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
	case "thxwinnercode":
		hasWon, err := h.GiveawayRepo.HasWonGiveawayByMessageId(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#GiveawayRepo.HasWonGiveawayByMessageId: %v", i.GuildID, err)
			return
		}
		if !hasWon {
			discord.RespondWithEphemeralMessage(s, i, "Nie wygrałeś tego giveawayu!")
			return
		}

		code, err := h.GiveawayRepo.GetCodeForInfoMessage(ctx, i.Message.ID)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#GiveawayRepo.GetCodeForInfoMessage: %v", i.GuildID, err)
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
			log.Printf("(%s) handleMessageComponents#session.InteractionRespond: %v", i.GuildID, err)
			return
		}
	case "msgwinnercode":
		hasWon, err := h.MessageGiveawayRepo.HasWonGiveawayByMessageId(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#MessageGiveawayRepo.HasWonGiveawayByMessageId: %v", i.GuildID, err)
			return
		}
		if !hasWon {
			discord.RespondWithEphemeralMessage(s, i, "Nie wygrałeś tego giveawayu!")
			return
		}

		codes, err := h.MessageGiveawayRepo.GetCodesForInfoMessage(ctx, i.Message.ID)
		if err != nil {
			log.Printf("(%s) handleMessageComponents#GiveawayRepo.GetCodeForInfoMessage: %v", i.GuildID, err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{discord.ConstructMessageWinnerEmbed(codes)},
			},
		})
		if err != nil {
			log.Printf("(%s) handleMessageComponents#session.InteractionRespond: %v", i.GuildID, err)
			return
		}
	case "accept", "reject":
		h.handleAcceptDeclineButtons(ctx, s, i)
	}
}

func (h InteractionCreateListener) handleAcceptDeclineButtons(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	isThxMessage, err := h.GiveawayRepo.IsThxMessage(ctx, i.Message.ID)
	if err != nil {
		log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.IsThxMessage: %v", i.GuildID, err)
		return
	}
	isThxmeMessage, err := h.GiveawayRepo.IsThxmeMessage(ctx, i.Message.ID)
	if err != nil {
		log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.IsThxmeMessage: %v", i.GuildID, err)
		return
	}

	if !isThxMessage && !isThxmeMessage {
		return
	}

	componentId := i.MessageComponentData().CustomID

	member := i.Member

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.Printf("Could not get server config for guild %s", i.GuildID)
		return
	}

	if isThxMessage {
		isAdmin := discord.HasAdminPermissions(s, member, serverConfig.AdminRoleId, i.GuildID)
		if !isAdmin {
			discord.RespondWithEphemeralMessage(s, i, "Nie masz uprawnień do akceptacji!")
			return
		}

		participant, err := h.GiveawayRepo.GetParticipant(ctx, i.Message.ID)
		if err != nil {
			log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipant: %v", i.GuildID, err)
			return
		}

		giveawayEnded, err := h.GiveawayRepo.IsGiveawayEnded(ctx, participant.GiveawayId)
		if err != nil {
			log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.IsGiveawayEnded: %v", i.GuildID, err)
			return
		}
		if giveawayEnded {
			discord.RespondWithEphemeralMessage(s, i, "Giveaway już się zakończył!")
			return
		}

		thxNotification, notificationErr := h.GiveawayRepo.GetThxNotification(ctx, i.Message.ID)
		if notificationErr != nil && !errors.Is(notificationErr, sql.ErrNoRows) {
			log.Printf("Could not get thx notification for message %s", i.Message.ID)
			return
		}

		switch componentId {
		case "accept":
			err = h.GiveawayRepo.UpdateParticipant(ctx, &participant, member.User.ID, member.User.Username, true)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipant: %v", i.GuildID, err)
				return
			}
			discord.RespondWithEphemeralMessage(s, i, "Udział użytkownika został potwierdzony!")
			log.Printf("(%s) %s accepted %s participation in giveaway %d", i.GuildID, member.User.Username, participant.UserName, participant.GiveawayId)

			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, participant.GiveawayId)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantNamesForGiveaway: %v", i.GuildID, err)
				return
			}

			embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, participant.UserId, member.User.ID, "confirm")

			_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#session.ChannelMessageEditEmbed: %v", i.GuildID, err)
				return
			}

			if errors.Is(notificationErr, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "confirm")
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
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "confirm")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", i.GuildID, err)
					return
				}
			}

			h.HelperService.CheckHelper(ctx, s, i.GuildID, participant.UserId)
		case "reject":
			err := h.GiveawayRepo.UpdateParticipant(ctx, &participant, member.User.ID, member.User.Username, false)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipant: %v", i.GuildID, err)
				return
			}
			discord.RespondWithEphemeralMessage(s, i, "Udział użytkownika został odrzucony!")
			log.Printf("(%s) %s rejected %s participation in giveaway %d", i.GuildID, member.User.Username, participant.UserName, participant.GiveawayId)

			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, participant.GiveawayId)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantNamesForGiveaway: %v", i.GuildID, err)
				return
			}

			embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, participant.UserId, member.User.ID, "reject")

			_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#session.ChannelMessageEditEmbed: %v", i.GuildID, err)
				return
			}

			if errors.Is(notificationErr, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "reject")
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
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "reject")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", i.GuildID, err)
					return
				}
			}

			h.HelperService.CheckHelper(ctx, s, i.GuildID, participant.UserId)
		}

	} else if isThxmeMessage {
		candidate, err := h.GiveawayRepo.GetParticipantCandidate(ctx, i.Message.ID)
		if err != nil {
			log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantCandidate: %v", i.GuildID, err)
			return
		}

		if member.User.ID != candidate.CandidateApproverId {
			discord.RespondWithEphemeralMessage(s, i, "Nie masz uprawnień do zmiany statusu tej prośby!")
			return
		}

		switch componentId {
		case "accept":
			err := h.GiveawayRepo.UpdateParticipantCandidate(ctx, &candidate, true)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipantCandidate: %v", i.GuildID, err)
				return
			}
			discord.RespondWithEphemeralMessage(s, i, "Prośba o podziękowanie zaakceptowana!")
			log.Printf("(%s) %s accepted %s request for thx", i.GuildID, member.User.Username, candidate.CandidateName)

			giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, i.GuildID)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.GetGiveawayForGuild: %v", i.GuildID, err)
				return
			}
			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, giveaway.Id)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantNamesForGiveaway: %v", i.GuildID, err)
				return
			}

			embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, candidate.CandidateId, "", "wait")

			_, err = s.ChannelMessageEdit(i.ChannelID, i.Message.ID, "Prośba o podziękowanie zaakceptowana przez: "+member.User.Mention())
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#session.ChannelMessageEdit: %v", i.GuildID, err)
				return
			}
			_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#session.ChannelMessageEditEmbed: %v", i.GuildID, err)
				return
			}

			guild, err := s.Guild(i.GuildID)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#session.Guild: %v", i.GuildID, err)
				return
			}

			err = h.GiveawayRepo.InsertParticipant(ctx, giveaway.Id, guild.ID, guild.Name, candidate.CandidateId, candidate.CandidateName, i.ChannelID, i.Message.ID)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.InsertParticipant: %v", i.GuildID, err)
				str := "Coś poszło nie tak przy dodawaniu podziękowania :("
				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &str,
				})
				if err != nil {
					log.Printf("(%s) handleAcceptDeclineButtons#session.InteractionResponseEdit: %v", i.GuildID, err)
				}
				return
			}

			log.Printf("(%s) %s thanked %s", i.GuildID, member.User.Username, candidate.CandidateName)

			thxNotification, err := h.GiveawayRepo.GetThxNotification(ctx, i.Message.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Could not get thx notification for message %s", i.Message.ID)
				return
			}

			if errors.Is(err, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, i.Message.ID, candidate.CandidateId, "", "wait")
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
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, i.Message.ID, candidate.CandidateId, "", "wait")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", i.GuildID, err)
					return
				}
			}
		case "reject":
			err := h.GiveawayRepo.UpdateParticipantCandidate(ctx, &candidate, false)
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipantCandidate: %v", i.GuildID, err)
				return
			}

			_, err = s.ChannelMessageEdit(i.ChannelID, i.Message.ID, fmt.Sprintf("%s, czy chcesz podziękować użytkownikowi %s? - Odrzucono", member.User.Mention(), candidate.CandidateName))
			if err != nil {
				log.Printf("(%s) handleAcceptDeclineButtons#session.ChannelMessageEdit: %v", i.GuildID, err)
				return
			}

			discord.RespondWithEphemeralMessage(s, i, "Prośba o podziękowanie odrzucona!")
			log.Printf("(%s) %s rejected %s request for thx", i.GuildID, member.User.Username, candidate.CandidateName)
		}

	}
}
