package discord

import (
	"context"
	"csrvbot/domain/entities"
	"csrvbot/pkg/logger"
	"fmt"
	"math/rand"
)

func PickLevelForGiveaway(ctx context.Context, serverRepo entities.ServerRepo, guildId string) (int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGuild(guildId)
	log.Debug("Picking level for giveaway")

	levels, err := serverRepo.GetConditionalGiveawayLevels(ctx, guildId)
	if err != nil {
		log.WithError(err).Error("PickLevelForGiveaway#serverRepo.GetConditionalGiveawayLevels")
		return 0, err
	}

	if len(levels) == 0 {
		return 0, fmt.Errorf("no levels found")
	}

	return levels[rand.Intn(len(levels))], nil
}
