package discord

import "github.com/bwmarrin/discordgo"

func ConstructThxWinnerCodeButton(disabled bool) *discordgo.Button {
	return &discordgo.Button{
		Label:    "Kliknij tutaj, aby wyświetlić kod",
		Style:    discordgo.SuccessButton,
		CustomID: "thxwinnercode",
		Emoji: &discordgo.ComponentEmoji{
			Name: "🎉",
		},
		Disabled: disabled,
	}
}

func ConstructMessageWinnerCodeButton(disabled bool) *discordgo.Button {
	return &discordgo.Button{
		Label:    "Kliknij tutaj, aby wyświetlić kod",
		Style:    discordgo.SuccessButton,
		CustomID: "msgwinnercode",
		Emoji: &discordgo.ComponentEmoji{
			Name: "🎉",
		},
		Disabled: disabled,
	}
}

func ConstructUnconditionalJoinButton(disabled bool) *discordgo.Button {
	return &discordgo.Button{
		Label:    "Weź udział",
		Style:    discordgo.SuccessButton,
		CustomID: "unconditionalgiveawayjoin",
		Emoji: &discordgo.ComponentEmoji{
			Name: "🙋",
		},
		Disabled: disabled,
	}
}

func ConstructUnconditionalWinnerCodeButton(disabled bool) *discordgo.Button {
	return &discordgo.Button{
		Label:    "Kliknij tutaj, aby wyświetlić kod",
		Style:    discordgo.SuccessButton,
		CustomID: "unconditionalwinnercode",
		Emoji: &discordgo.ComponentEmoji{
			Name: "🎉",
		},
		Disabled: disabled,
	}
}

func ConstructAcceptButton(disabled bool) *discordgo.Button {
	return &discordgo.Button{
		Label:    "",
		Style:    discordgo.SuccessButton,
		CustomID: "accept",
		Emoji: &discordgo.ComponentEmoji{
			Name: "✅",
		},
		Disabled: disabled,
	}
}

func ConstructRejectButton(disabled bool) *discordgo.Button {
	return &discordgo.Button{
		Label:    "",
		Style:    discordgo.DangerButton,
		CustomID: "reject",
		Emoji: &discordgo.ComponentEmoji{
			Name: "⛔",
		},
		Disabled: disabled,
	}
}
