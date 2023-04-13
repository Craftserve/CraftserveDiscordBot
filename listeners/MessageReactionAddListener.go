package listeners

import (
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

type MessageReactionAddListener struct {
	GiveawayHours string
	UserRepo      repos.UserRepo
	GiveawayRepo  repos.GiveawayRepo
	ServerRepo    repos.ServerRepo
	HelperService services.HelperService
}

func NewMessageReactionAddListener(giveawayHours string, userRepo *repos.UserRepo, giveawayRepo *repos.GiveawayRepo, serverRepo *repos.ServerRepo, helperService *services.HelperService) MessageReactionAddListener {
	return MessageReactionAddListener{
		GiveawayHours: giveawayHours,
		UserRepo:      *userRepo,
		GiveawayRepo:  *giveawayRepo,
		ServerRepo:    *serverRepo,
		HelperService: *helperService,
	}
}

func (h MessageReactionAddListener) Handle(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	ctx := pkg.CreateContext()
	if r.UserID == s.State.User.ID {
		return
	}

	isThxMessage, err := h.GiveawayRepo.IsThxMessage(ctx, r.MessageID)
	if err != nil {
		log.Println("("+r.GuildID+") "+"handleGiveawayReactions#h.GiveawayRepo.IsThxMessage", err)
		return
	}
	isThxmeMessage, err := h.GiveawayRepo.IsThxmeMessage(ctx, r.MessageID)
	if err != nil {
		log.Println("("+r.GuildID+") "+"handleGiveawayReactions#h.GiveawayRepo.IsThxmeMessage", err)
		return
	}

	if !isThxMessage && !isThxmeMessage {
		return
	}

	if r.Emoji.Name != "✅" && r.Emoji.Name != "⛔" {
		err := s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
		if err != nil {
			log.Println("("+r.GuildID+") "+"handleGiveawayReactions#s.MessageReactionRemove", err)
		}
		return
	}

	member, err := s.GuildMember(r.GuildID, r.UserID)
	if err != nil {
		log.Println("("+r.GuildID+") "+"handleGiveawayReactions#s.GuildMember", err)
		return
	}

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, r.GuildID)
	if err != nil {
		log.Printf("Could not get server config for guild %s", r.GuildID)
		return
	}

	if isThxMessage {
		if !discord.HasAdminPermissions(s, member, serverConfig.AdminRoleId, r.GuildID) {
			err = s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
			if err != nil {
				log.Println("("+r.GuildID+") "+"handleGiveawayReactions#s.MessageReactionRemove", err)
			}
			return
		}
		// reactionists...
		participant, err := h.GiveawayRepo.GetParticipant(ctx, r.MessageID)
		if err != nil {
			log.Println("handleGiveawayReactions#GetParticipant", err)
			return
		}

		thxNotification, notificationErr := h.GiveawayRepo.GetThxNotification(ctx, r.MessageID)
		if notificationErr != nil && !errors.Is(notificationErr, sql.ErrNoRows) {
			log.Printf("Could not get thx notification for message %s", r.MessageID)
			return
		}

		switch r.Emoji.Name {
		case "✅":
			err := h.GiveawayRepo.UpdateParticipant(ctx, &participant, r.UserID, member.User.Username, true)
			if err != nil {
				log.Println("handleGiveawayReactions#UpdateParticipant", err)
				return
			}
			log.Println(member.User.Username + "(" + member.User.ID + ") zaakceptował udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))

			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, participant.GiveawayId)
			if err != nil {
				log.Println("("+r.GuildID+") Could not get participants", err)
				return
			}

			embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, participant.UserId, r.UserID, "confirm")

			_, err = s.ChannelMessageEditEmbed(r.ChannelID, r.MessageID, embed)
			if err != nil {
				log.Println("("+r.GuildID+") Could not update message", err)
				return
			}

			if errors.Is(notificationErr, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", r.GuildID, r.ChannelID, r.MessageID, participant.UserId, r.UserID, "confirm")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", r.GuildID, err)
					return
				}

				err = h.GiveawayRepo.InsertThxNotification(ctx, r.MessageID, notificationMessageId)
				if err != nil {
					log.Printf("(%s) Could not insert thx notification", r.GuildID)
					return
				}
			} else {
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, r.GuildID, r.ChannelID, r.MessageID, participant.UserId, r.UserID, "confirm")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", r.GuildID, err)
					return
				}
			}

			h.HelperService.CheckHelper(ctx, s, r.GuildID, participant.UserId)
			break
		case "⛔":
			err := h.GiveawayRepo.UpdateParticipant(ctx, &participant, r.UserID, member.User.Username, false)
			if err != nil {
				log.Println("handleGiveawayReactions#UpdateParticipant", err)
				return
			}
			log.Println(member.User.Username + "(" + member.User.ID + ") odrzucił udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))

			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, participant.GiveawayId)
			if err != nil {
				log.Println("("+r.GuildID+") Could not get participants", err)
				return
			}

			embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, participant.UserId, r.UserID, "reject")

			_, err = s.ChannelMessageEditEmbed(r.ChannelID, r.MessageID, embed)
			if err != nil {
				log.Println("("+r.GuildID+") Could not update message", err)
				return
			}

			if errors.Is(notificationErr, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", r.GuildID, r.ChannelID, r.MessageID, participant.UserId, r.UserID, "reject")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", r.GuildID, err)
					return
				}

				err = h.GiveawayRepo.InsertThxNotification(ctx, r.MessageID, notificationMessageId)
				if err != nil {
					log.Printf("(%s) Could not insert thx notification", r.GuildID)
					return
				}
			} else {
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, r.GuildID, r.ChannelID, r.MessageID, participant.UserId, r.UserID, "reject")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", r.GuildID, err)
					return
				}
			}

			h.HelperService.CheckHelper(ctx, s, r.GuildID, participant.UserId)
			break
		}
	} else if isThxmeMessage {
		candidate, err := h.GiveawayRepo.GetParticipantCandidate(ctx, r.MessageID)
		if err != nil {
			log.Println("handleGiveawayReactions#GetParticipant", err)
			return
		}

		if r.UserID != candidate.CandidateApproverId {
			err = s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
			if err != nil {
				log.Println("("+r.GuildID+") "+"handleThxme#s.MessageReactionRemove", err)
			}
			return
		}
		// reactionists...
		switch r.Emoji.Name {
		case "✅":
			err := h.GiveawayRepo.UpdateParticipantCandidate(ctx, &candidate, true)
			if err != nil {
				log.Println("handleGiveawayReactions#UpdateParticipant", err)
				return
			}
			log.Println(candidate.CandidateApproverName + "(" + candidate.CandidateApproverId + ") zaakceptował prosbe o thx uzytkownika " + candidate.CandidateName + "(" + candidate.CandidateId + ")")

			giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, r.GuildID)
			if err != nil {
				log.Println("("+r.GuildID+") Could not get giveaway", err)
				return
			}
			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, giveaway.Id)
			if err != nil {
				log.Println("("+r.GuildID+") Could not get participants", err)
				return
			}

			embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, candidate.CandidateId, "", "wait")

			err = s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)

			_, err = s.ChannelMessageEdit(r.ChannelID, r.MessageID, "Prośba o podziękowanie zaakceptowana przez: "+member.User.Mention())
			if err != nil {
				log.Println("("+r.GuildID+") Could not update message", err)
				return
			}
			_, err = s.ChannelMessageEditEmbed(r.ChannelID, r.MessageID, embed)
			if err != nil {
				log.Println("("+r.GuildID+") Could not update message", err)
				return
			}

			guild, err := s.Guild(r.GuildID)
			if err != nil {
				log.Println("("+r.GuildID+") handleThxCommand#session.Guild", err)
				return
			}

			err = h.GiveawayRepo.InsertParticipant(ctx, giveaway.Id, r.GuildID, guild.Name, candidate.CandidateId, candidate.CandidateName, r.ChannelID, r.MessageID)
			if err != nil {
				log.Println("("+r.GuildID+") Could not insert participant", err)
				_, err = s.ChannelMessageSend(r.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
				return
			}
			log.Println("(" + r.GuildID + ") " + member.User.Username + " has thanked " + candidate.CandidateName)

			thxNotification, err := h.GiveawayRepo.GetThxNotification(ctx, r.MessageID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Could not get thx notification for message %s", r.MessageID)
				return
			}

			if errors.Is(err, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", r.GuildID, r.ChannelID, r.MessageID, candidate.CandidateId, "", "wait")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", r.GuildID, err)
					return
				}

				err = h.GiveawayRepo.InsertThxNotification(ctx, r.MessageID, notificationMessageId)
				if err != nil {
					log.Printf("(%s) Could not insert thx notification", r.GuildID)
					return
				}
			} else {
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, r.GuildID, r.ChannelID, r.MessageID, candidate.CandidateId, "", "wait")
				if err != nil {
					log.Printf("(%s) Could not notify thx on thx info channel: %v", r.GuildID, err)
					return
				}
			}
			break
		case "⛔":
			err := h.GiveawayRepo.UpdateParticipantCandidate(ctx, &candidate, false)
			if err != nil {
				log.Println("handleGiveawayReactions#UpdateParticipant", err)
				return
			}
			log.Println(candidate.CandidateApproverName + "(" + candidate.CandidateApproverId + ") odrzucił prosbe o thx uzytkownika " + candidate.CandidateName + "(" + candidate.CandidateId + ")")

			break
		}
	}

}
