package listeners

import (
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"github.com/bwmarrin/discordgo"
	"log"
)

type MessageCreateListener struct {
	MessageGiveawayRepo repos.MessageGiveawayRepo
}

func NewMessageCreateListener(messageGiveawayRepo *repos.MessageGiveawayRepo) MessageCreateListener {
	return MessageCreateListener{
		MessageGiveawayRepo: *messageGiveawayRepo,
	}
}

func (h MessageCreateListener) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := pkg.CreateContext()

	if m.Author.Bot {
		return
	}

	err := h.MessageGiveawayRepo.UpdateUserDailyMessageCount(ctx, m.Author.ID, m.GuildID)
	if err != nil {
		log.Printf("(%s) Could not update user daily message count: %v", m.GuildID, err)
		return
	}
}
