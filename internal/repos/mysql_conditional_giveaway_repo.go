package repos

import (
	"context"
	"github.com/go-gorp/gorp"
	"time"
)

type ConditionalGiveawayRepo struct {
	mysql *gorp.DbMap
}

func NewConditionalGiveawayRepo(mysql *gorp.DbMap) *ConditionalGiveawayRepo {
	mysql.AddTableWithName(ConditionalGiveaway{}, "conditional_giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(ConditionalGiveawayWinner{}, "conditional_giveaway_winners").SetKeys(true, "id")
	mysql.AddTableWithName(ConditionalGiveawayParticipant{}, "conditional_participants").SetKeys(true, "id")

	return &ConditionalGiveawayRepo{mysql: mysql}
}

type ConditionalGiveaway struct {
	Id            int        `db:"id, primarykey, autoincrement"`
	StartTime     time.Time  `db:"start_time"`
	EndTime       *time.Time `db:"end_time"`
	GuildId       string     `db:"guild_id,size:255"`
	InfoMessageId string     `db:"info_message_id,size:255"`
	Level         int        `db:"level,default:0"`
}

type ConditionalGiveawayWinner struct {
	Id                    int    `db:"id, primarykey, autoincrement"`
	ConditionalGiveawayId int    `db:"conditional_giveaway_id"`
	UserId                string `db:"user_id,size:255"`
	Code                  string `db:"code,size:255"`
}

type ConditionalGiveawayParticipant struct {
	Id                    int       `db:"id, primarykey, autoincrement"`
	ConditionalGiveawayId int       `db:"conditional_giveaway_id"`
	UserId                string    `db:"user_id,size:255"`
	UserName              string    `db:"user_name,size:255"`
	JoinTime              time.Time `db:"join_time"`
	UserLevel             int       `db:"user_level,default:0"`
}

func (repo *ConditionalGiveawayRepo) GetGiveawayForGuild(ctx context.Context, guildId string) (ConditionalGiveaway, error) {
	var giveaway ConditionalGiveaway
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, info_message_id, level FROM conditional_giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil {
		return ConditionalGiveaway{}, err
	}

	return giveaway, nil
}

func (repo *ConditionalGiveawayRepo) GetUnfinishedGiveaways(ctx context.Context) ([]ConditionalGiveaway, error) {
	var giveaways []ConditionalGiveaway
	_, err := repo.mysql.WithContext(ctx).Select(&giveaways, "SELECT id, start_time, end_time, guild_id, info_message_id, level FROM conditional_giveaways WHERE end_time IS NULL")
	if err != nil {
		return nil, err
	}

	return giveaways, nil
}

func (repo *ConditionalGiveawayRepo) GetParticipantsForGiveaway(ctx context.Context, giveawayId int) ([]ConditionalGiveawayParticipant, error) {
	var participants []ConditionalGiveawayParticipant
	_, err := repo.mysql.WithContext(ctx).Select(&participants, "SELECT id, conditional_giveaway_id, user_id, user_name, join_time, user_level FROM conditional_participants WHERE conditional_giveaway_id = ?", giveawayId)
	if err != nil {
		return nil, err
	}

	return participants, nil
}

func (repo *ConditionalGiveawayRepo) CountParticipantsForGiveaway(ctx context.Context, giveawayId int) (int, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM conditional_participants WHERE conditional_giveaway_id = ?", giveawayId)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (repo *ConditionalGiveawayRepo) InsertGiveaway(ctx context.Context, guildId, messageId string, level int) error {
	giveaway := &ConditionalGiveaway{
		StartTime:     time.Now(),
		GuildId:       guildId,
		InfoMessageId: messageId,
		Level:         level,
	}
	err := repo.mysql.WithContext(ctx).Insert(giveaway)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ConditionalGiveawayRepo) InsertParticipant(ctx context.Context, giveawayId, level int, userId string, userName string) error {
	participant := &ConditionalGiveawayParticipant{
		ConditionalGiveawayId: giveawayId,
		UserId:                userId,
		UserName:              userName,
		JoinTime:              time.Now(),
		UserLevel:             level,
	}
	err := repo.mysql.WithContext(ctx).Insert(participant)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ConditionalGiveawayRepo) InsertWinner(ctx context.Context, giveawayId int, userId, code string) error {
	winner := &ConditionalGiveawayWinner{
		ConditionalGiveawayId: giveawayId,
		UserId:                userId,
		Code:                  code,
	}
	err := repo.mysql.WithContext(ctx).Insert(winner)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ConditionalGiveawayRepo) UpdateGiveaway(ctx context.Context, giveaway *ConditionalGiveaway, messageId string) error {
	now := time.Now()
	giveaway.EndTime = &now
	giveaway.InfoMessageId = messageId
	_, err := repo.mysql.WithContext(ctx).Update(giveaway)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ConditionalGiveawayRepo) HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error) {
	result, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM conditional_giveaways, conditional_giveaway_winners WHERE info_message_id = ? AND user_id = ? AND conditional_giveaways.id = conditional_giveaway_winners.conditional_giveaway_id", messageId, userId)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (repo *ConditionalGiveawayRepo) GetCodeForInfoMessage(ctx context.Context, messageId, userId string) (string, error) {
	code, err := repo.mysql.WithContext(ctx).SelectStr("SELECT code FROM conditional_giveaways, conditional_giveaway_winners WHERE info_message_id = ? AND user_id = ? AND conditional_giveaways.id = conditional_giveaway_winners.conditional_giveaway_id", messageId, userId)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (repo *ConditionalGiveawayRepo) IsGiveawayFinishedByMessageId(ctx context.Context, messageId string) (bool, error) {
	result, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM conditional_giveaways WHERE info_message_id = ? AND end_time IS NOT NULL", messageId)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}
