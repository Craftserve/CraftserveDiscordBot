package commands

import (
	"context"
	"csrvbot/internal/services"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type DocCommand struct {
	Name         string
	Description  string
	DMPermission bool
	GithubClient services.GithubClient
}

func NewDocCommand(githubClient *services.GithubClient) DocCommand {
	return DocCommand{
		Name:         "doc",
		Description:  "Wysyła link do danego poradnika",
		DMPermission: false,
		GithubClient: *githubClient,
	}
}

func (h DocCommand) Register(ctx context.Context, s *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)
	log.Debug("Registering command")
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "nazwa",
				Description:  "Nazwa poradnika",
				Required:     true,
				Autocomplete: true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "anchor",
				Description: "Nazwa sekcji (nagłówka)",
				Required:    false,
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not register command")
	}
}

func (h DocCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	docName := i.ApplicationCommandData().Options[0].StringValue()

	docExists, err := h.GithubClient.GetDocExists(docName)
	if err != nil {
		log.WithError(err).Error("Could not get doc")
		discord.RespondWithMessage(ctx, s, i, "Wystąpił błąd podczas wyszukiwania poradnika")
		return
	}

	if !docExists {
		discord.RespondWithMessage(ctx, s, i, "Taki poradnik nie istnieje")
		return
	}

	anchor := ""
	if len(i.ApplicationCommandData().Options) == 2 {
		anchor = "#" + i.ApplicationCommandData().Options[1].StringValue()
	}
	discord.RespondWithMessage(ctx, s, i, "<https://github.com/craftserve/docs/blob/master/"+docName+".md"+anchor+">")
}

func (h DocCommand) HandleAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	docs, err := h.GithubClient.GetDocs(ctx, data.Options[0].StringValue())
	if err != nil {
		log.WithError(err).Error("Could not get docs")
		return
	}

	for _, doc := range docs {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  doc,
			Value: doc,
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not respond to interaction")
	}
}
