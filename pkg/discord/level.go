package discord

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/pkg/logger"
	"math/rand"
)

func PickLevelForGiveaway(ctx context.Context, serverRepo repos.ServerRepo, guildId string) (*int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	log.Debug("Picking level for giveaway")

	levels, err := serverRepo.GetConditionalGiveawayLevels(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("PickLevelForGiveaway#serverRepo.GetConditionalGiveawayLevels")
		return nil, err
	}

	if len(levels) == 0 {
		return nil, nil
	}

	return &levels[rand.Intn(len(levels))], nil
}
