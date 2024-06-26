package repos

import (
	"context"
	"csrvbot/domain/entities"
	"github.com/go-gorp/gorp"
	"time"
)

type JoinableGiveawayRepo struct {
	mysql *gorp.DbMap
}

type SqlJoinableGiveaway struct {
	Id            int        `db:"id,primarykey,autoincrement"`
	StartTime     time.Time  `db:"start_time"`
	EndTime       *time.Time `db:"end_time"`
	GuildId       string     `db:"guild_id,size:255"`
	InfoMessageId string     `db:"info_message_id,size:255"`
	Level         *int       `db:"level"`
}

type SqlJoinableGiveawayWinner struct {
	Id                 int    `db:"id,primarykey,autoincrement"`
	JoinableGiveawayId int    `db:"joinable_giveaway_id"`
	UserId             string `db:"user_id,size:255"`
	Code               string `db:"code,size:255"`
}

type SqlJoinableGiveawayParticipant struct {
	Id                 int       `db:"id,primarykey,autoincrement"`
	JoinableGiveawayId int       `db:"joinable_giveaway_id"`
	UserId             string    `db:"user_id,size:255"`
	UserName           string    `db:"user_name,size:255"`
	JoinTime           time.Time `db:"join_time"`
	UserLevel          *int      `db:"user_level"`
}

func FromSqlJoinableGiveaway(giveaway *SqlJoinableGiveaway) *entities.JoinableGiveaway {
	return &entities.JoinableGiveaway{
		Id:            giveaway.Id,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		InfoMessageId: giveaway.InfoMessageId,
		Level:         giveaway.Level,
	}
}

func ToSqlJoinableGiveaway(giveaway *entities.JoinableGiveaway) *SqlJoinableGiveaway {
	return &SqlJoinableGiveaway{
		Id:            giveaway.Id,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		InfoMessageId: giveaway.InfoMessageId,
		Level:         giveaway.Level,
	}
}

func FromSqlJoinableGiveawayWinner(winner *SqlJoinableGiveawayWinner) *entities.JoinableGiveawayWinner {
	return &entities.JoinableGiveawayWinner{
		Id:                 winner.Id,
		JoinableGiveawayId: winner.JoinableGiveawayId,
		UserId:             winner.UserId,
		Code:               winner.Code,
	}
}

func ToSqlJoinableGiveawayWinner(winner *entities.JoinableGiveawayWinner) *SqlJoinableGiveawayWinner {
	return &SqlJoinableGiveawayWinner{
		Id:                 winner.Id,
		JoinableGiveawayId: winner.JoinableGiveawayId,
		UserId:             winner.UserId,
		Code:               winner.Code,
	}
}

func FromSqlJoinableGiveawayParticipant(participant *SqlJoinableGiveawayParticipant) *entities.JoinableGiveawayParticipant {
	return &entities.JoinableGiveawayParticipant{
		Id:                 participant.Id,
		JoinableGiveawayId: participant.JoinableGiveawayId,
		UserId:             participant.UserId,
		UserName:           participant.UserName,
		JoinTime:           participant.JoinTime,
		UserLevel:          participant.UserLevel,
	}
}

func ToSqlJoinableGiveawayParticipant(participant *entities.JoinableGiveawayParticipant) *SqlJoinableGiveawayParticipant {
	return &SqlJoinableGiveawayParticipant{
		Id:                 participant.Id,
		JoinableGiveawayId: participant.JoinableGiveawayId,
		UserId:             participant.UserId,
		UserName:           participant.UserName,
		JoinTime:           participant.JoinTime,
		UserLevel:          participant.UserLevel,
	}
}

func NewJoinableGiveawayRepo(mysql *gorp.DbMap) *JoinableGiveawayRepo {
	mysql.AddTableWithName(SqlJoinableGiveaway{}, "joinable_giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(SqlJoinableGiveawayWinner{}, "joinable_giveaway_winners").SetKeys(true, "id")
	mysql.AddTableWithName(SqlJoinableGiveawayParticipant{}, "joinable_participants").SetKeys(true, "id")

	return &JoinableGiveawayRepo{
		mysql: mysql,
	}
}

func (repo *JoinableGiveawayRepo) GetGiveawayForGuild(ctx context.Context, guildId string, withLevel bool) (result *entities.JoinableGiveaway, err error) {
	var giveaway SqlJoinableGiveaway

	if withLevel {
		err = repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, info_message_id, level FROM joinable_giveaways WHERE guild_id = ? AND end_time IS NULL AND level IS NOT NULL", guildId)
	} else {
		err = repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, info_message_id, level FROM joinable_giveaways WHERE guild_id = ? AND end_time IS NULL AND level IS NULL", guildId)
	}

	if err != nil {
		return nil, err
	}

	return FromSqlJoinableGiveaway(&giveaway), nil
}

func (repo *JoinableGiveawayRepo) GetUnfinishedGiveaways(ctx context.Context, withLevel bool) (result []entities.JoinableGiveaway, err error) {
	var giveaways []SqlJoinableGiveaway

	if withLevel {
		_, err = repo.mysql.WithContext(ctx).Select(&giveaways, "SELECT id, start_time, end_time, guild_id, info_message_id, level FROM joinable_giveaways WHERE end_time IS NULL AND level IS NOT NULL")
	} else {
		_, err = repo.mysql.WithContext(ctx).Select(&giveaways, "SELECT id, start_time, end_time, guild_id, info_message_id, level FROM joinable_giveaways WHERE end_time IS NULL AND level IS NULL")
	}

	if err != nil {
		return nil, err
	}

	for _, giveaway := range giveaways {
		result = append(result, *FromSqlJoinableGiveaway(&giveaway))
	}

	return result, nil
}

func (repo *JoinableGiveawayRepo) GetParticipantsForGiveaway(ctx context.Context, giveawayId int) ([]entities.JoinableGiveawayParticipant, error) {
	var participants []SqlJoinableGiveawayParticipant
	_, err := repo.mysql.WithContext(ctx).Select(&participants, "SELECT id, joinable_giveaway_id, user_id, user_name, join_time, user_level FROM joinable_participants WHERE joinable_giveaway_id = ?", giveawayId)
	if err != nil {
		return nil, err
	}

	var result []entities.JoinableGiveawayParticipant
	for _, participant := range participants {
		result = append(result, *FromSqlJoinableGiveawayParticipant(&participant))
	}

	return result, nil
}

func (repo *JoinableGiveawayRepo) CountParticipantsForGiveaway(ctx context.Context, giveawayId int) (int, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM joinable_participants WHERE joinable_giveaway_id = ?", giveawayId)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (repo *JoinableGiveawayRepo) InsertGiveaway(ctx context.Context, guildId, messageId string, level *int) error {
	giveaway := &SqlJoinableGiveaway{
		StartTime:     time.Now(),
		GuildId:       guildId,
		InfoMessageId: messageId,
		Level:         level,
	}
	if err := repo.mysql.WithContext(ctx).Insert(giveaway); err != nil {
		return err
	}

	return nil
}

func (repo *JoinableGiveawayRepo) InsertParticipant(ctx context.Context, giveawayId, level int, userId string, userName string) error {
	participant := &SqlJoinableGiveawayParticipant{
		JoinableGiveawayId: giveawayId,
		UserId:             userId,
		UserName:           userName,
		JoinTime:           time.Now(),
		UserLevel:          &level,
	}
	if err := repo.mysql.WithContext(ctx).Insert(participant); err != nil {
		return err
	}

	return nil
}

func (repo *JoinableGiveawayRepo) InsertWinner(ctx context.Context, giveawayId int, userId, code string) error {
	winner := &SqlJoinableGiveawayWinner{
		JoinableGiveawayId: giveawayId,
		UserId:             userId,
		Code:               code,
	}
	if err := repo.mysql.WithContext(ctx).Insert(winner); err != nil {
		return err
	}

	return nil
}

func (repo *JoinableGiveawayRepo) FinishGiveaway(ctx context.Context, giveaway *entities.JoinableGiveaway, messageId string) error {
	now := time.Now()
	giveaway.EndTime = &now
	giveaway.InfoMessageId = messageId
	if _, err := repo.mysql.WithContext(ctx).Update(ToSqlJoinableGiveaway(giveaway)); err != nil {
		return err
	}

	return nil
}

func (repo *JoinableGiveawayRepo) HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error) {
	result, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM joinable_giveaways g JOIN joinable_giveaway_winners w ON g.id = w.joinable_giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (repo *JoinableGiveawayRepo) GetCodeForInfoMessage(ctx context.Context, messageId, userId string) (string, error) {
	code, err := repo.mysql.WithContext(ctx).SelectStr("SELECT code FROM joinable_giveaways g JOIN joinable_giveaway_winners w ON g.id = w.joinable_giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (repo *JoinableGiveawayRepo) GetGiveawayByMessageId(ctx context.Context, messageId string) (*entities.JoinableGiveaway, error) {
	var giveaway SqlJoinableGiveaway
	if err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, info_message_id, level FROM joinable_giveaways WHERE info_message_id = ?", messageId); err != nil {
		return nil, err
	}

	return FromSqlJoinableGiveaway(&giveaway), nil
}
