package commands

import (
	"context"
	"csrvbot/domain/entities"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type StatusCommand struct {
	Name         string
	Description  string
	DMPermission bool
	ServerRepo   entities.ServerRepo
	StatusRepo   entities.StatusRepo
}

func NewStatusCommand(serverRepo entities.ServerRepo, statusRepo entities.StatusRepo) StatusCommand {
	return StatusCommand{
		Name:         "status",
		Description:  "Zarządza kanałem statusowym",
		DMPermission: false,
		ServerRepo:   serverRepo,
		StatusRepo:   statusRepo,
	}
}

func (h StatusCommand) Register(ctx context.Context, s *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)
	log.Debug("Registering command")

	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Ustaw status",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "template",
						Description:  "Wybierz szablon statusu",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "edit",
				Description: "Edytuj szablon",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "name",
						Description:  "Wybierz szablon do edycji",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "create",
				Description: "Utwórz nowy szablon",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Usuń szablon",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "name",
						Description:  "Wybierz szablon do usunięcia",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not register command")
	}
}

func (h StatusCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)

	switch i.ApplicationCommandData().Options[0].Name {
	case "edit":
		id := i.ApplicationCommandData().Options[0].Options[0].StringValue()

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			log.WithError(err).Error("Invalid status ID")
		}

		status, err := h.StatusRepo.GetStatusById(ctx, intId)

		if err != nil {
			log.WithError(err).Error("Could not get status by id")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można znaleźć szablonu statusu o podanym ID.")
			return
		}

		statusModal := discord.ConstructStatusEditOrCreateModalComponent(status)
		discord.RespondWithModal(ctx, s, i, statusModal)
	case "create":
		statusModal := discord.ConstructStatusEditOrCreateModalComponent(nil)
		discord.RespondWithModal(ctx, s, i, statusModal)
	case "remove":
		id := i.ApplicationCommandData().Options[0].Options[0].StringValue()

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			log.WithError(err).Error("Invalid status ID")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nieprawidłowe ID szablonu statusu.")
		}

		err = h.StatusRepo.RemoveStatus(ctx, int(intId))

		if err != nil {
			log.WithError(err).Error("Could not remove status")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można usunąć szablonu statusu.")
			return
		}

		discord.RespondWithEphemeralMessage(ctx, s, i, "Szablon statusu został pomyślnie usunięty.")
	case "set":
		id := i.ApplicationCommandData().Options[0].Options[0].StringValue()

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			log.WithError(err).Error("Invalid status ID")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nieprawidłowe ID szablonu statusu.")
		}

		status, err := h.StatusRepo.GetStatusById(ctx, intId)
		if err != nil {
			log.WithError(err).Error("Could not get status by id")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można znaleźć szablonu statusu o podanym ID.")
			return
		}

		if status.GuildId != i.GuildID {
			log.Error("Status does not belong to this guild")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można znaleźć szablonu statusu o podanym ID.")
			return
		}

		serverSettings, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
		if err != nil {
			log.WithError(err).Error("Could not get server settings")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można pobrać ustawień serwera.")
			return
		}

		var languagesChannels map[string]string
		err = json.Unmarshal([]byte(serverSettings.StatusChannelsId), &languagesChannels)
		if err != nil {
			log.WithError(err).Error("Could not unmarshal status channels")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można przetworzyć ustawień kanału statusowego.")
			return
		}

		var languagesContent map[string]string
		err = json.Unmarshal(status.Content, &languagesContent)
		if err != nil {
			log.WithError(err).Error("Could not unmarshal status content")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można przetworzyć zawartości szablonu statusu.")
			return
		}

		for lang, channelId := range languagesChannels {
			content, exists := languagesContent[lang]
			if !exists {
				continue
			}

			messages, err := s.ChannelMessages(channelId, 10, "", "", "")
			if err != nil {
				log.WithError(err).WithField("channelId", channelId).Error("Could not fetch messages from channel")
				discord.RespondWithEphemeralMessage(ctx, s, i, fmt.Sprintf("Nie można pobrać wiadomości z kanału o ID %s.", channelId))
				return
			}

			for _, message := range messages {
				err := s.ChannelMessageDelete(channelId, message.ID)
				if err != nil {
					log.WithError(err).WithField("channelId", channelId).WithField("messageId", message.ID).Error("Could not delete message from channel")
					discord.RespondWithEphemeralMessage(ctx, s, i, fmt.Sprintf("Nie można usunąć wiadomości z kanału o ID %s.", channelId))
					return
				}
			}

			_, err = s.ChannelMessageSend(channelId, content)
			if err != nil {
				log.WithError(err).WithField("channelId", channelId).Error("Could not send status message to channel")
				discord.RespondWithEphemeralMessage(ctx, s, i, fmt.Sprintf("Nie można wysłać wiadomości statusowej na kanał o ID %s.", channelId))
				return
			}

			var emoji string

			switch status.Type {
			case "OUTAGE":
				emoji = "🔴"
			case "MAINTENANCE":
				emoji = "🟡"
			case "OPERATIONAL":
				emoji = "🟢"
			}

			channelName := fmt.Sprintf("%s┇status-%s", emoji, lang)

			// Discord have rate limits for editing channels (2 times per 10 minutes),
			// so we use a goroutine to avoid blocking response from command
			go func(chID, chName string) {
				_, err := s.ChannelEdit(chID, &discordgo.ChannelEdit{
					Name: chName,
				})
				if err != nil {
					log.WithError(err).WithField("channelId", chID).Error("Could not edit channel name")
					discord.RespondFollowUpMessage(ctx, s, i, fmt.Sprintf("Nie można zaktualizować nazwy kanału <#%s>.", chID))
					return
				}
			}(channelId, channelName)
		}

		discord.RespondWithEphemeralMessage(ctx, s, i, "Kanały statusowe zostały pomyślnie zaktualizowane.")
	default:
		log.Error("Unknown subcommand")
		discord.RespondWithEphemeralMessage(ctx, s, i, "Nieznana podkomenda.")
	}
}

func (h StatusCommand) HandleModalSubmit(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)

	action := strings.Split(i.ModalSubmitData().CustomID, "_")

	contentPl := i.ModalSubmitData().Components[2].(*discordgo.Label).Component.(*discordgo.TextInput).Value
	contentEn := i.ModalSubmitData().Components[3].(*discordgo.Label).Component.(*discordgo.TextInput).Value
	content := map[string]string{
		"pl": contentPl,
		"en": contentEn,
	}

	shortName := i.ModalSubmitData().Components[0].(*discordgo.Label).Component.(*discordgo.TextInput).Value
	statusType := i.ModalSubmitData().Components[1].(*discordgo.Label).Component.(*discordgo.SelectMenu).Values[0]
	contentJson, err := json.Marshal(content)
	if err != nil {
		log.WithError(err).Error("Could not marshal status content")
		discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można przetworzyć zawartości szablonu statusu.")
		return
	}

	status := &entities.Status{
		GuildId:   i.GuildID,
		ShortName: shortName,
		Type:      statusType,
		Content:   contentJson,
	}

	if action[1] == "create" {
		err := h.StatusRepo.CreateStatus(ctx, status)
		if err != nil {
			log.WithError(err).Error("Could not create status")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można utworzyć szablonu statusu.")
			return
		}

		discord.RespondWithEphemeralMessage(ctx, s, i, "Szablon statusu został pomyślnie utworzony.")
	} else if action[1] == "edit" {
		id, err := strconv.ParseInt(action[2], 10, 64)

		if err != nil {
			log.WithError(err).Error("Invalid status ID")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nieprawidłowe ID szablonu statusu.")
		}

		status.Id = int(id)

		err = h.StatusRepo.UpdateStatus(ctx, status)

		if err != nil {
			log.WithError(err).Error("Could not update status")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie można zaktualizować szablonu statusu.")
			return
		}

		discord.RespondWithEphemeralMessage(ctx, s, i, "Szablon statusu został pomyślnie zaktualizowany.")
	}
}

func (h StatusCommand) HandleAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	log.Debug("Getting matching status templates for " + data.Options[0].Options[0].StringValue())

	statuses, err := h.StatusRepo.GetAllStatuses(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("Could not get statuses")
		return
	}

	for _, status := range statuses {
		if status.GuildId != i.GuildID {
			continue
		}

		if strings.Contains(status.ShortName, data.Options[0].Options[0].StringValue()) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  fmt.Sprintf("%s (%s)", status.ShortName, status.Type),
				Value: strconv.Itoa(status.Id),
			})
		}
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
