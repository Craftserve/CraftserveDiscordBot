package repos

import (
	"context"
	"database/sql"
	"github.com/go-gorp/gorp"
	"time"
)

type MessageGiveawayRepo struct {
	mysql *gorp.DbMap
}

func NewMessageGiveawayRepo(mysql *gorp.DbMap) *MessageGiveawayRepo {
	mysql.AddTableWithName(MessageGiveaway{}, "message_giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(MessageGiveawayWinner{}, "message_giveaway_winners").SetKeys(true, "id")
	mysql.AddTableWithName(DailyUserMessages{}, "daily_user_messages").SetKeys(true, "id").SetUniqueTogether("day", "user_id", "guild_id")

	return &MessageGiveawayRepo{mysql: mysql}
}

type MessageGiveaway struct {
	Id            int            `db:"id, primarykey, autoincrement"`
	StartTime     time.Time      `db:"start_time"`
	EndTime       *time.Time     `db:"end_time"`
	GuildId       string         `db:"guild_id,size:255"`
	InfoMessageId sql.NullString `db:"info_message_id,size:255"`
}

type MessageGiveawayWinner struct {
	Id                int    `db:"id, primarykey, autoincrement"`
	MessageGiveawayId int    `db:"message_giveaway_id"`
	UserId            string `db:"user_id,size:255"`
	Code              string `db:"code,size:255"`
}

type DailyUserMessages struct {
	Id      int    `db:"id, primarykey, autoincrement"`
	UserId  string `db:"user_id,size:255"`
	Day     string `db:"day"` // fixme should be date
	GuildId string `db:"guild_id,size:255"`
	Count   int    `db:"count"`
}

func (repo *MessageGiveawayRepo) InsertMessageGiveaway(ctx context.Context, guildId string) error {
	giveaway := &MessageGiveaway{
		StartTime: time.Now(),
		GuildId:   guildId,
	}
	err := repo.mysql.WithContext(ctx).Insert(giveaway)
	if err != nil {
		return err
	}
	return nil
}

func (repo *MessageGiveawayRepo) GetMessageGiveaway(ctx context.Context, guildId string) (MessageGiveaway, error) {
	var giveaway MessageGiveaway
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, guild_id, info_message_id FROM message_giveaways WHERE guild_id = ? AND info_message_id IS NULL", guildId)
	if err != nil {
		return MessageGiveaway{}, err
	}
	return giveaway, nil
}

func (repo *MessageGiveawayRepo) UpdateMessageGiveaway(ctx context.Context, messageGiveaway *MessageGiveaway, messageId string) error {
	now := time.Now()
	messageGiveaway.EndTime = &now
	messageGiveaway.InfoMessageId = sql.NullString{String: messageId, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(messageGiveaway)
	if err != nil {
		return err
	}
	return nil
}

func (repo *MessageGiveawayRepo) UpdateUserDailyMessageCount(ctx context.Context, userId string, guildId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("INSERT INTO daily_user_messages (user_id, day, guild_id, count) VALUES (?, date(now()), ?, 1) ON DUPLICATE KEY UPDATE count = count + 1", userId, guildId)
	return err
}

func (repo *MessageGiveawayRepo) GetUsersWithMessagesFromLastDays(ctx context.Context, dayCount int, guildId string) ([]string, error) {
	var users []string
	_, err := repo.mysql.WithContext(ctx).Select(&users, "SELECT DISTINCT user_id FROM daily_user_messages WHERE guild_id = ? AND day > date_sub(now(), INTERVAL ? DAY)", guildId, dayCount)
	return users, err
}

func (repo *MessageGiveawayRepo) InsertMessageGiveawayWinner(ctx context.Context, giveawayId int, userId, code string) error {
	winner := &MessageGiveawayWinner{
		MessageGiveawayId: giveawayId,
		UserId:            userId,
		Code:              code,
	}
	err := repo.mysql.WithContext(ctx).Insert(winner)
	if err != nil {
		return err
	}
	return nil
}
