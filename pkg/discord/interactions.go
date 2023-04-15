package discord

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func RespondWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		log.Println("("+i.GuildID+") Could not respond to interaction ("+i.ID+")", err)
	}
}

func RespondWithEphemeralMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: message,
		},
	})
	if err != nil {
		log.Println("("+i.GuildID+") Could not respond to interaction ("+i.ID+")", err)
	}
}
