package discord

import "github.com/bwmarrin/discordgo"

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
