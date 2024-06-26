package entities

import (
	"context"
	"time"
)

type JoinableGiveaway struct {
	Id            int        `json:"id"`
	StartTime     time.Time  `json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
	GuildId       string     `json:"guildId"`
	InfoMessageId string     `json:"infoMessageId"`
	Level         *int       `json:"level"`
}

type JoinableGiveawayWinner struct {
	Id                 int    `bson:"id"`
	JoinableGiveawayId int    `bson:"joinableGiveawayId"`
	UserId             string `bson:"userId"`
	Code               string `bson:"code"`
}

type JoinableGiveawayParticipant struct {
	Id                 int       `json:"id"`
	JoinableGiveawayId int       `json:"joinableGiveawayId"`
	UserId             string    `json:"userId"`
	UserName           string    `json:"userName"`
	JoinTime           time.Time `json:"joinTime"`
	UserLevel          *int      `json:"userLevel"`
}

type JoinableGiveawayRepo interface {
	GetGiveawayForGuild(ctx context.Context, guildId string, withLevel bool) (*JoinableGiveaway, error)
	GetUnfinishedGiveaways(ctx context.Context, withLevel bool) ([]JoinableGiveaway, error)
	GetParticipantsForGiveaway(ctx context.Context, giveawayId int) ([]JoinableGiveawayParticipant, error)
	CountParticipantsForGiveaway(ctx context.Context, giveawayId int) (int, error)
	InsertGiveaway(ctx context.Context, guildId, messageId string, level *int) error
	InsertParticipant(ctx context.Context, giveawayId, level int, userId string, userName string) error
	InsertWinner(ctx context.Context, giveawayId int, userId, code string) error
	FinishGiveaway(ctx context.Context, giveaway *JoinableGiveaway, messageId string) error
	HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error)
	GetCodeForInfoMessage(ctx context.Context, messageId, userId string) (string, error)
	GetGiveawayByMessageId(ctx context.Context, messageId string) (*JoinableGiveaway, error)
}
