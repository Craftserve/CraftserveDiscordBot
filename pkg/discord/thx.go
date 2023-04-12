package discord

import (
	"github.com/bwmarrin/discordgo"
)

func NotifyThxOnThxInfoChannel(s *discordgo.Session, thxInfoChannelId, thxNotificationMessageId, guildId, channelId, thxMessageId, participantId, confirmerId, state string) (string, error) {
	embed := ConstructThxNotificationEmbed(guildId, channelId, thxMessageId, participantId, confirmerId, state)

	if thxInfoChannelId == "" {
		return "", nil
	}

	if thxNotificationMessageId == "" {
		message, err := s.ChannelMessageSendEmbed(thxInfoChannelId, embed)
		if err != nil {
			return "", err
		}
		return message.ID, nil
	} else {
		_, err := s.ChannelMessageEditEmbed(thxInfoChannelId, thxNotificationMessageId, embed)
		if err != nil {
			return "", err
		}
		return thxNotificationMessageId, nil
	}
}
