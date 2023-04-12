package discord

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"time"
)

func FinishGiveaway(ctx context.Context, s *discordgo.Session, serverRepo repos.ServerRepo, giveawayRepo repos.GiveawayRepo, csrvClient services.CsrvClient, guildId string) {
	giveaway, err := giveawayRepo.GetGiveawayForGuild(ctx, guildId)
	if err != nil {
		log.Println("("+guildId+") Could not get giveaway", err)
		return
	}
	_, err = s.Guild(guildId)
	if err != nil {
		log.Println("("+guildId+") finishGiveaway#session.Guild", err)
		return
	}

	giveawayChannelId, err := serverRepo.GetMainChannelForGuild(ctx, guildId)
	if err != nil {
		log.Println("("+guildId+") finishGiveaway#serverRepo.GetMainChannelForGuild", err)
		return
	}

	participants, err := giveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.Println("("+guildId+") finishGiveaway#DbMap.Select", err)
	}

	if participants == nil || len(participants) == 0 {
		message, err := s.ChannelMessageSend(giveawayChannelId, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie był w loterii.")
		if err != nil {
			log.Println("("+guildId+") finishGiveaway#session.ChannelMessageSend", err)
		}
		err = giveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID, "", "", "")
		if err != nil {
			log.Println("("+guildId+") finishGiveaway#DbMap.Update", err)
		}
		log.Println("(" + guildId + ") Giveaway ended without any participants.")
		return
	}

	code, err := csrvClient.GetCSRVCode()
	if err != nil {
		log.Println("("+guildId+") finishGiveaway#getCSRVCode", err)
		_, err = s.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
		if err != nil {
			return
		}
		return
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	winner := participants[rand.Intn(len(participants))]

	member, err := s.GuildMember(guildId, winner.UserId)
	if err != nil {
		log.Println("("+guildId+") finishGiveaway#session.GuildMember", err)
		return
	}
	dmEmbed := ConstructWinnerEmbed(code)
	dm, err := s.UserChannelCreate(winner.UserId)
	if err != nil {
		log.Println("("+guildId+") finishGiveaway#session.UserChannelCreate", err)
	}
	_, err = s.ChannelMessageSendEmbed(dm.ID, dmEmbed)

	mainEmbed := ConstructChannelWinnerEmbed(member.User.Username)
	message, err := s.ChannelMessageSendComplex(giveawayChannelId, &discordgo.MessageSend{
		Embed: mainEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Kliknij tutaj, aby wyświetlić kod",
						Style:    discordgo.SuccessButton,
						CustomID: "winnercode",
						Emoji: discordgo.ComponentEmoji{
							Name: "🎉",
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.Println("("+guildId+") finishGiveaway#session.ChannelMessageSendEmbed", err)
	}

	err = giveawayRepo.UpdateGiveaway(ctx, &giveaway, message.ID, code, winner.UserId, member.User.Username)
	if err != nil {
		log.Println("("+guildId+") finishGiveaway#DbMap.Update", err)
	}
	log.Println("(" + guildId + ") " + member.User.Username + " has won the giveaway (Code: " + code + ").")

}

func FinishGiveaways(ctx context.Context, s *discordgo.Session, giveawayRepo repos.GiveawayRepo, serverRepo repos.ServerRepo, csrvClient services.CsrvClient) {
	giveaways, err := giveawayRepo.GetUnfinishedGiveaways(ctx)
	if err != nil {
		log.Println("finishGiveaways#giveawayRepo.GetUnfinishedGiveaways", err)
		return
	}
	for _, giveaway := range giveaways {
		FinishGiveaway(ctx, s, serverRepo, giveawayRepo, csrvClient, giveaway.GuildId)
		guild, err := s.Guild(giveaway.GuildId)
		if err == nil {
			CreateMissingGiveaways(ctx, s, serverRepo, giveawayRepo, guild)
		}
	}

}

// Copied to commands/CsrvbotCommand.go and listeners/GuildCreateListener.go

func CreateMissingGiveaways(ctx context.Context, s *discordgo.Session, serverRepo repos.ServerRepo, giveawayRepo repos.GiveawayRepo, guild *discordgo.Guild) {
	serverConfig, err := serverRepo.GetServerConfigForGuild(ctx, guild.ID)
	if err != nil {
		log.Printf("(%s) createMissingGiveaways#ServerRepo.GetServerConfigForGuild: %v", guild.ID, err)
		return
	}
	giveawayChannelId := serverConfig.MainChannel
	_, err = s.Channel(giveawayChannelId)
	if err != nil {
		log.Printf("(%s) createMissingGiveaways#Session.Channel: %v", guild.ID, err)
		return
	}
	_, err = giveawayRepo.GetGiveawayForGuild(ctx, guild.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("(%s) createMissingGiveaways#giveawayRepo.GetGiveawayForGuild: %v", guild.ID, err)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		err = giveawayRepo.InsertGiveaway(ctx, guild.ID, guild.Name)
		if err != nil {
			log.Printf("(%s) createMissingGiveaways#giveawayRepo.InsertGiveaway: %v", guild.ID, err)
			return
		}
	}

}
