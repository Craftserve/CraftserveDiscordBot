package repos

import (
	"context"
	"csrvbot/domain/entities"
	"database/sql"
	"github.com/go-gorp/gorp"
	"time"
)

type MessageGiveawayRepo struct {
	mysql *gorp.DbMap
	tx    *gorp.Transaction
}

func NewMessageGiveawayRepo(mysql *gorp.DbMap) *MessageGiveawayRepo {
	mysql.AddTableWithName(SqlMessageGiveaway{}, "message_giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(SqlMessageGiveawayWinner{}, "message_giveaway_winners").SetKeys(true, "id")
	mysql.AddTableWithName(SqlDailyUserMessages{}, "daily_user_messages").SetKeys(true, "id").SetUniqueTogether("day", "user_id", "guild_id")

	return &MessageGiveawayRepo{mysql: mysql}
}

type SqlMessageGiveaway struct {
	Id            int            `db:"id, primarykey, autoincrement"`
	StartTime     time.Time      `db:"start_time"`
	EndTime       *time.Time     `db:"end_time"`
	GuildId       string         `db:"guild_id,size:255"`
	InfoMessageId sql.NullString `db:"info_message_id,size:255"`
}

type SqlMessageGiveawayWinner struct {
	Id                int    `db:"id, primarykey, autoincrement"`
	MessageGiveawayId int    `db:"message_giveaway_id"`
	UserId            string `db:"user_id,size:255"`
	Code              string `db:"code,size:255"`
}

type SqlDailyUserMessages struct {
	Id      int    `db:"id, primarykey, autoincrement"`
	UserId  string `db:"user_id,size:255"`
	Day     string `db:"day"` // fixme should be date
	GuildId string `db:"guild_id,size:255"`
	Count   int    `db:"count"`
}

func FromSqlMessageGiveaway(giveaway *SqlMessageGiveaway) *entities.MessageGiveaway {
	return &entities.MessageGiveaway{
		Id:            giveaway.Id,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		InfoMessageId: giveaway.InfoMessageId,
	}
}

func ToSqlMessageGiveaway(giveaway *entities.MessageGiveaway) *SqlMessageGiveaway {
	return &SqlMessageGiveaway{
		Id:            giveaway.Id,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		InfoMessageId: giveaway.InfoMessageId,
	}
}

func FromSqlMessageGiveawayWinner(winner *SqlMessageGiveawayWinner) *entities.MessageGiveawayWinner {
	return &entities.MessageGiveawayWinner{
		Id:                winner.Id,
		MessageGiveawayId: winner.MessageGiveawayId,
		UserId:            winner.UserId,
		Code:              winner.Code,
	}
}

func ToSqlMessageGiveawayWinner(winner *entities.MessageGiveawayWinner) *SqlMessageGiveawayWinner {
	return &SqlMessageGiveawayWinner{
		Id:                winner.Id,
		MessageGiveawayId: winner.MessageGiveawayId,
		UserId:            winner.UserId,
		Code:              winner.Code,
	}
}

func FromSqlDailyUserMessages(dailyUserMessages *SqlDailyUserMessages) *entities.DailyUserMessages {
	return &entities.DailyUserMessages{
		Id:      dailyUserMessages.Id,
		UserId:  dailyUserMessages.UserId,
		Day:     dailyUserMessages.Day,
		GuildId: dailyUserMessages.GuildId,
		Count:   dailyUserMessages.Count,
	}
}

func ToSqlDailyUserMessages(dailyUserMessages *entities.DailyUserMessages) *SqlDailyUserMessages {
	return &SqlDailyUserMessages{
		Id:      dailyUserMessages.Id,
		UserId:  dailyUserMessages.UserId,
		Day:     dailyUserMessages.Day,
		GuildId: dailyUserMessages.GuildId,
		Count:   dailyUserMessages.Count,
	}
}

func (repo *MessageGiveawayRepo) WithTx(ctx context.Context, tx *gorp.Transaction) (entities.MessageGiveawayRepo, *gorp.Transaction, error) {
	if tx == nil {
		var err error
		tx, err = repo.mysql.Begin()
		if err != nil {
			return repo, nil, err
		}

		tx = tx.WithContext(ctx).(*gorp.Transaction)
	}

	newRepo := *repo
	newRepo.tx = tx

	return &newRepo, tx, nil
}

func (repo *MessageGiveawayRepo) InsertMessageGiveaway(ctx context.Context, guildId string) error {
	giveaway := &SqlMessageGiveaway{
		StartTime: time.Now(),
		GuildId:   guildId,
	}
	sqlExecutor := repo.mysql.WithContext(ctx)

	if repo.tx != nil {
		sqlExecutor = repo.tx
	}

	err := sqlExecutor.Insert(giveaway)
	if err != nil {
		return err
	}
	return nil
}

func (repo *MessageGiveawayRepo) GetMessageGiveaway(ctx context.Context, guildId string) (entities.MessageGiveaway, error) {
	var giveaway SqlMessageGiveaway
	sqlExecutor := repo.mysql.WithContext(ctx)

	query := "SELECT id, start_time, guild_id, info_message_id FROM message_giveaways WHERE guild_id = ? AND info_message_id IS NULL"
	if repo.tx != nil {
		sqlExecutor = repo.tx
		query += " FOR UPDATE"
	}

	err := sqlExecutor.SelectOne(&giveaway, query, guildId)
	if err != nil {
		return entities.MessageGiveaway{}, err
	}
	return *FromSqlMessageGiveaway(&giveaway), nil
}

func (repo *MessageGiveawayRepo) FinishMessageGiveaway(ctx context.Context, messageGiveaway *entities.MessageGiveaway, messageId string) error {
	now := time.Now()
	messageGiveaway.EndTime = &now
	messageGiveaway.InfoMessageId = sql.NullString{String: messageId, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(ToSqlMessageGiveaway(messageGiveaway))
	if err != nil {
		return err
	}
	return nil
}

func (repo *MessageGiveawayRepo) UpdateUserDailyMessageCount(ctx context.Context, userId string, guildId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("INSERT INTO daily_user_messages (user_id, day, guild_id, count) VALUES (?, DATE(now()), ?, 1) ON DUPLICATE KEY UPDATE COUNT = COUNT + 1", userId, guildId)
	return err
}

func (repo *MessageGiveawayRepo) GetUsersWithMessagesFromLastDays(ctx context.Context, dayCount int, guildId string) ([]string, error) {
	var users []string
	_, err := repo.mysql.WithContext(ctx).Select(&users, "SELECT DISTINCT user_id FROM daily_user_messages WHERE guild_id = ? AND DAY > date_sub(now(), INTERVAL ? DAY)", guildId, dayCount)
	return users, err
}

func (repo *MessageGiveawayRepo) InsertMessageGiveawayWinner(ctx context.Context, giveawayId int, userId, code string) error {
	winner := &SqlMessageGiveawayWinner{
		MessageGiveawayId: giveawayId,
		UserId:            userId,
		Code:              code,
	}
	sqlExecutor := repo.mysql.WithContext(ctx)

	if repo.tx != nil {
		sqlExecutor = repo.tx
	}

	err := sqlExecutor.Insert(winner)
	if err != nil {
		return err
	}
	return nil
}

func (repo *MessageGiveawayRepo) HasWonGiveawayByMessageId(ctx context.Context, infoMessageId, userId string) (bool, error) {
	result, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM message_giveaways g JOIN message_giveaway_winners w ON g.id = w.message_giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", infoMessageId, userId)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (repo *MessageGiveawayRepo) GetCodesForInfoMessage(ctx context.Context, messageId, userId string) ([]string, error) {
	var codes []string
	_, err := repo.mysql.WithContext(ctx).Select(&codes, "SELECT code FROM message_giveaways g JOIN message_giveaway_winners w ON g.id = w.message_giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	return codes, err
}

func (repo *MessageGiveawayRepo) GetLastCodesForUser(ctx context.Context, userId string, limit int) ([]string, error) {
	var codes []string
	_, err := repo.mysql.WithContext(ctx).Select(&codes, "SELECT code FROM message_giveaway_winners WHERE user_id = ? ORDER BY id DESC LIMIT ?", userId, limit)
	if err != nil {
		return nil, err
	}
	return codes, nil
}
