package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

const (
	ICON_URL = "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp"
	COLOR    = 0x234d20
)

func ConstructInfoEmbed(url string, participants []string, giveawayHours string, value int) *discordgo.MessageEmbed {
	info := "**Ten bot organizuje giveaway kodów na doładowanie portfela Twojego serwera.**\n" +
		fmt.Sprintf("**Każdy kod doładowuje %d PLN do portfela.**\n", value/100) +
		"Aby wziąć udział pomagaj innym użytkownikom. Jeżeli komuś pomożesz, to poproś tą osobę aby użyła komendy </thx:1107007500659728405> lub sam użyj komendy </thxme:1107007504308769020> - w ten sposób dostaniesz się do loterii. To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
		fmt.Sprintf("**Sponsorem tego bota jest %s - hosting serwerów Minecraft.**\n\n", url) +
		"Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.\n\n" +
		"Uczestnicy: " + strings.Join(participants, ", ") + "\n\nNagrody rozdajemy o " + giveawayHours + ", Powodzenia!"
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    "Informacje o Giveawayu",
			IconURL: ICON_URL,
		},
		Description: info,
		Color:       COLOR,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return embed
}

func ConstructThxEmbed(url string, participants []string, giveawayHours, participantId, confirmerId, state string, voucherValue int) *discordgo.MessageEmbed {
	embed := ConstructInfoEmbed(url, participants, giveawayHours, voucherValue)
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Dodany", Value: "<@" + participantId + ">", Inline: true})

	var status string
	switch state {
	case "wait":
		status = "Oczekiwanie"
		break
	case "confirm":
		status = "Potwierdzono"
		if confirmerId != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Potwierdzający", Value: "<@" + confirmerId + ">", Inline: true})
		}
		break
	case "reject":
		status = "Odrzucono"
		if confirmerId != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Odrzucający", Value: "<@" + confirmerId + ">", Inline: true})
		}
		break
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: status, Inline: true})

	return embed
}

func ConstructThxNotificationEmbed(url, guildId, thxChannelId, thxMessageId, participantId, confirmerId, state string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Timestamp: time.Now().Format(time.RFC3339),
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    "Nowe podziękowanie",
			IconURL: ICON_URL,
		},
		Color: COLOR,
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Dla", Value: "<@" + participantId + ">", Inline: true})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Kanał", Value: "<#" + thxChannelId + ">", Inline: true})

	switch state {
	case "wait":
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Oczekiwanie", Inline: true})
		break
	case "confirm":
		if confirmerId != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Potwierdzono przez <@" + confirmerId + ">", Inline: true})
		}
		break
	case "reject":
		if confirmerId == "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Odrzucono", Inline: true})
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Odrzucono przez <@" + confirmerId + ">", Inline: true})
		}
		break
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Link", Value: "https://discordapp.com/channels/" + guildId + "/" + thxChannelId + "/" + thxMessageId, Inline: false})

	return embed
}

func ConstructWinnerEmbed(url, code string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    "Wygrałeś kod na doładowanie portfela na Craftserve!",
			IconURL: ICON_URL,
		},
		Description: "Gratulacje! W loterii wygrałeś darmowy kod na doładowanie portfela na Craftserve! Możesz go użyć w zakładce *Płatności* pod przyciskiem *Zrealizuj kupon*. Kod jest ważny około rok.",
		Color:       COLOR,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "KOD", Value: code,
			},
		},
	}
	return embed
}

func ConstructMessageWinnerEmbed(url string, codes []string) *discordgo.MessageEmbed {
	var description string
	var author string
	var name string
	if len(codes) > 1 {
		description = "Gratulacje! W loterii wygrałeś darmowe kody na doładowanie portfela na Craftserve! Możesz je użyć w zakładce *Płatności* pod przyciskiem *Zrealizuj kupon*. Kody są ważne około rok."
		author = "Wygrałeś kody na doładowanie portfela na Craftserve!"
		name = "KODY"
	} else {
		description = "Gratulacje! W loterii wygrałeś darmowy kod na doładowanie portfela na Craftserve! Możesz go użyć w zakładce *Płatności* pod przyciskiem *Zrealizuj kod podarunkowy*. Kod jest ważny około rok."
		author = "Wygrałeś kod na doładowanie serwera na Craftserve!"
		name = "KOD"
	}
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    author,
			IconURL: ICON_URL,
		},
		Color:       COLOR,
		Description: description,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: name, Value: strings.Join(codes, "\n"),
			},
		},
	}
	return embed
}

func ConstructChannelWinnerEmbed(url, username string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    "Wyniki giveaway!",
			IconURL: ICON_URL,
		},
		Color:       COLOR,
		Description: username + " wygrał kod. Gratulacje!",
	}
	return embed
}

func ConstructChannelMessageWinnerEmbed(url string, usernames []string) *discordgo.MessageEmbed {
	var description string
	if len(usernames) > 1 {
		description = "W loterii za aktywność wygraliście kody na doładowanie portfela na Craftserve!"
		description += "\n\n" + strings.Join(usernames, "\n")
	} else {
		description = "W loterii za aktywność " + usernames[0] + " wygrał kod na doładowanie portfela na Craftserve. Gratulacje!"
	}
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    "Wyniki giveaway!",
			IconURL: ICON_URL,
		},
		Color:       COLOR,
		Description: description,
	}
	return embed
}

func ConstructResendEmbed(url string, codes []string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    "Twoje ostatnie wygrane kody",
			IconURL: ICON_URL,
		},
		Description: strings.Join(codes, "\n"),
		Color:       COLOR,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return embed
}

func ConstructJoinableGiveawayEmbed(url string, participantsCount int, levelRoleId *string) *discordgo.MessageEmbed {
	var title, description string
	if levelRoleId != nil {
		title = "Dołącz do poziomowego giveaway już teraz!"
		description = fmt.Sprintf("Właśnie startuje poziomowy giveaway, do którego mogą dołączyć użytkownicy z rolą **<@&%s>** lub wyższą! Wystarczy, że klikniesz w przycisk poniżej i już jesteś w grze o darmowy kod na doładowanie portfela na Craftserve! Powodzenia!", *levelRoleId)
		if participantsCount > 0 {
			description += fmt.Sprintf("\n\n**Liczba uczestników:** %d", participantsCount)
		}
	} else {
		title = "Dołącz do giveaway już teraz!"
		description = "Właśnie startuje bezwarunkowy giveaway, do którego może dołączyć totalnie każdy! Wystarczy, że klikniesz w przycisk poniżej i już jesteś w grze o darmowy kod na doładowanie portfela na Craftserve! Powodzenia!"
		if participantsCount > 0 {
			description += fmt.Sprintf("\n\n**Liczba uczestników:** %d", participantsCount)
		}
	}

	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    title,
			IconURL: ICON_URL,
		},
		Color:       COLOR,
		Description: description,
	}
}

func ConstructJoinableWinnersEmbed(url string, participantsIds []string, levelRoleId *string) *discordgo.MessageEmbed {
	var title, description string
	if levelRoleId != nil {
		title = "Zakończono poziomowy giveaway!"
		description = fmt.Sprintf("Oto zwycięzcy warunkowego giveawaya dla użytkowników z rolą **<@&%s>** lub wyższą! Gratulacje!", *levelRoleId)
	} else {
		title = "Zakończono bezwarunkowy giveaway!"
		description = "Oto zwycięzcy bezwarunkowego giveawaya! Gratulacje!"
	}

	for _, id := range participantsIds {
		description += "\n- <@" + id + ">"
	}

	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     url,
			Name:    title,
			IconURL: ICON_URL,
		},
		Color:       COLOR,
		Description: description,
	}
}
