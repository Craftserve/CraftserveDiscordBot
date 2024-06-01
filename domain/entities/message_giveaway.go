package entities

import (
	"context"
	"database/sql"
	"github.com/go-gorp/gorp"
	"time"
)

type MessageGiveaway struct {
	Id            int            `json:"id"`
	StartTime     time.Time      `json:"startTime"`
	EndTime       *time.Time     `json:"endTime"`
	GuildId       string         `json:"guildId"`
	InfoMessageId sql.NullString `json:"infoMessageId"`
}

type MessageGiveawayWinner struct {
	Id                int    `json:"id"`
	MessageGiveawayId int    `json:"messageGiveawayId"`
	UserId            string `json:"userId"`
	Code              string `json:"code"`
}

type DailyUserMessages struct {
	Id      int    `json:"id"`
	UserId  string `json:"userId"`
	Day     string `json:"day"`
	GuildId string `json:"guildId"`
	Count   int    `json:"count"`
}

type MessageGiveawayRepo interface {
	WithTx(ctx context.Context, tx *gorp.Transaction) (MessageGiveawayRepo, *gorp.Transaction, error)
	InsertMessageGiveaway(ctx context.Context, guildId string) error
	GetMessageGiveaway(ctx context.Context, guildId string) (MessageGiveaway, error)
	FinishMessageGiveaway(ctx context.Context, messageGiveaway *MessageGiveaway, messageId string) error
	UpdateUserDailyMessageCount(ctx context.Context, userId string, guildId string) error
	GetUsersWithMessagesFromLastDays(ctx context.Context, dayCount int, guildId string) ([]string, error)
	InsertMessageGiveawayWinner(ctx context.Context, giveawayId int, userId, code string) error
	HasWonGiveawayByMessageId(ctx context.Context, infoMessageId, userId string) (bool, error)
	GetCodesForInfoMessage(ctx context.Context, messageId, userId string) ([]string, error)
	GetLastCodesForUser(ctx context.Context, userId string, limit int) ([]string, error)
}
