package discord

import (
	"context"
	"csrvbot/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func RespondLoading(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.WithError(err).Error("Could not respond to interaction")
	}
}

func EditResponseMessage(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	log := logger.GetLoggerFromContext(ctx)
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
	if err != nil {
		log.WithError(err).Error("Could not edit interaction")
	}
}

func RespondWithMessage(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	log := logger.GetLoggerFromContext(ctx)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not respond to interaction")
	}
}

func RespondWithEphemeralMessage(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	log := logger.GetLoggerFromContext(ctx)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: message,
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not respond to interaction")
	}
}

func RespondFollowUpMessage(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	log := logger.GetLoggerFromContext(ctx)
	_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: message,
	})

	if err != nil {
		log.WithError(err).Error("Could not create follow-up message")
	}
}

func RespondWithModal(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, modal discordgo.InteractionResponseData) {
	log := logger.GetLoggerFromContext(ctx)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &modal,
	})
	if err != nil {
		log.WithError(err).Error("Could not respond to interaction with modal")
	}
}
