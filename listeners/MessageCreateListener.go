package listeners

import (
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
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
	log := logger.GetLoggerFromContext(ctx).WithMessage(m.ID).WithUser(m.Author.ID).WithGuild(m.GuildID)

	if m.Author.Bot {
		return
	}

	err := h.MessageGiveawayRepo.UpdateUserDailyMessageCount(ctx, m.Author.ID, m.GuildID)
	if err != nil {
		log.WithError(err).Error("Could not update user daily message count")
		return
	}
}
