package listeners

import (
	"csrvbot/domain/entities"
	"csrvbot/pkg"
	"csrvbot/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

type MessageCreateListener struct {
	GiveawaysRepo entities.GiveawaysRepo
}

func NewMessageCreateListener(giveawaysRepo entities.GiveawaysRepo) MessageCreateListener {
	return MessageCreateListener{
		GiveawaysRepo: giveawaysRepo,
	}
}

func (h MessageCreateListener) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := pkg.CreateContext()
	log := logger.GetLoggerFromContext(ctx).WithMessage(m.ID).WithUser(m.Author.ID).WithGuild(m.GuildID)

	if m.Author.Bot {
		return
	}

	err := h.GiveawaysRepo.UpdateUserDailyMessageCount(ctx, m.Author.ID, m.GuildID)
	if err != nil {
		log.WithError(err).Error("Could not update user daily message count")
		return
	}
}
