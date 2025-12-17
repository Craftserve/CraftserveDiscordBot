package discord

import (
	"csrvbot/domain/entities"
	"encoding/json"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func ConstructThxWinnerComponents(disabled bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "Kliknij tutaj, aby wyświetlić kod",
					Style:    discordgo.SuccessButton,
					CustomID: "thxwinnercode",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎉",
					},
					Disabled: disabled,
				},
			},
		},
	}
}

func ConstructMessageWinnerComponents(disabled bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "Kliknij tutaj, aby wyświetlić kod",
					Style:    discordgo.SuccessButton,
					CustomID: "msgwinnercode",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎉",
					},
					Disabled: disabled,
				},
			},
		},
	}
}

func ConstructAcceptRejectComponents(disabled bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "",
					Style:    discordgo.SuccessButton,
					CustomID: "accept",
					Emoji: &discordgo.ComponentEmoji{
						Name: "✅",
					},
					Disabled: disabled,
				},
				&discordgo.Button{
					Label:    "",
					Style:    discordgo.DangerButton,
					CustomID: "reject",
					Emoji: &discordgo.ComponentEmoji{
						Name: "⛔",
					},
					Disabled: disabled,
				},
			},
		},
	}
}

func ConstructJoinComponents(disabled bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "Weź udział",
					Style:    discordgo.SuccessButton,
					CustomID: "giveawayjoin",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🙋",
					},
					Disabled: disabled,
				},
			},
		},
	}
}

func ConstructJoinableGiveawayWinnerComponents(disabled bool) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "Kliknij tutaj, aby wyświetlić kod",
					Style:    discordgo.SuccessButton,
					CustomID: "joinablewinnercode",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎉",
					},
					Disabled: disabled,
				},
			},
		},
	}
}

func ConstructStatusEditOrCreateModalComponent(data *entities.Status, action string) discordgo.InteractionResponseData {

	var customId string
	var title string

	switch action {
	case "create":
		customId = "status_create"
		title = "Utwórz nowy szablon statusu"
	case "edit":
		title = "Edytuj szablon statusu"
		customId = "status_edit_" + strconv.Itoa(data.Id)
	case "set":
		title = "Ustaw status serwera"
		customId = "status_set_" + strconv.Itoa(data.Id)
	}

	var content map[string]string

	err := json.Unmarshal(data.Content, &content)
	if err != nil {
		content = map[string]string{
			"pl": "",
			"en": "",
		}
	}

	var currentType string
	if data != nil {
		currentType = data.Type
	} else {
		currentType = "OPERATIONAL"
	}

	return discordgo.InteractionResponseData{
		CustomID: customId,
		Title:    title,
		Flags:    discordgo.MessageFlagsIsComponentsV2,
		Components: []discordgo.MessageComponent{
			discordgo.Label{
				Label:       "Nazwa szablonu",
				Description: "Wpisz nazwę szablonu statusu",
				Component: discordgo.TextInput{
					Style:       discordgo.TextInputShort,
					CustomID:    "short_name",
					Placeholder: "Wpisz nazwę szablonu tutaj...",
					Required:    true,
					Value: func() string {
						if data != nil {
							return data.ShortName
						} else {
							return ""
						}
					}(),
				},
			},
			discordgo.Label{
				Label:       "Typ statusu",
				Description: "Wybierz typ statusu",
				Component: discordgo.SelectMenu{
					MenuType: discordgo.StringSelectMenu,
					CustomID: "type",
					Options: []discordgo.SelectMenuOption{
						{Label: "Awaria", Value: "OUTAGE", Default: currentType == "OUTAGE"},
						{Label: "Konserwacja", Value: "MAINTENANCE", Default: currentType == "MAINTENANCE"},
						{Label: "Brak Awarii", Value: "OPERATIONAL", Default: currentType == "OPERATIONAL"},
					},
				},
			},
			discordgo.Label{
				Label:       "Treść statusu (PL)",
				Description: "Wpisz treść statusu w języku polskim",
				Component: discordgo.TextInput{
					Style:       discordgo.TextInputParagraph,
					CustomID:    "content_pl",
					Placeholder: "Wpisz treść statusu tutaj...",
					Required:    true,
					Value: func() string {
						if data != nil {
							return content["pl"]
						} else {
							return ""
						}
					}(),
				},
			},
			discordgo.Label{
				Label:       "Treść statusu (EN)",
				Description: "Wpisz treść statusu w języku angielskim",
				Component: discordgo.TextInput{
					Style:       discordgo.TextInputParagraph,
					CustomID:    "content_en",
					Placeholder: "Wpisz treść statusu tutaj...",
					Required:    true,
					Value: func() string {
						if data != nil {
							return content["en"]
						} else {
							return ""
						}
					}(),
				},
			},
		},
	}
}
