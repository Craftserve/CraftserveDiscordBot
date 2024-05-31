package repos

import (
	"context"
	"github.com/go-gorp/gorp"
	"time"
)

type UnconditionalGiveawayRepo struct {
	mysql *gorp.DbMap
}

func NewUnconditionalGiveawayRepo(mysql *gorp.DbMap) *UnconditionalGiveawayRepo {
	mysql.AddTableWithName(UnconditionalGiveaway{}, "unconditional_giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(UnconditionalGiveawayWinner{}, "unconditional_giveaway_winners").SetKeys(true, "id")
	mysql.AddTableWithName(UnconditionalGiveawayParticipant{}, "unconditional_participants").SetKeys(true, "id")

	return &UnconditionalGiveawayRepo{mysql: mysql}
}

type UnconditionalGiveaway struct {
	Id            int        `db:"id, primarykey, autoincrement"`
	StartTime     time.Time  `db:"start_time"`
	EndTime       *time.Time `db:"end_time"`
	GuildId       string     `db:"guild_id,size:255"`
	InfoMessageId string     `db:"info_message_id,size:255"`
}

type UnconditionalGiveawayWinner struct {
	Id                      int    `db:"id, primarykey, autoincrement"`
	UnconditionalGiveawayId int    `db:"unconditional_giveaway_id"`
	UserId                  string `db:"user_id,size:255"`
	Code                    string `db:"code,size:255"`
}

type UnconditionalGiveawayParticipant struct {
	Id                      int       `db:"id, primarykey, autoincrement"`
	UnconditionalGiveawayId int       `db:"unconditional_giveaway_id"`
	UserId                  string    `db:"user_id,size:255"`
	UserName                string    `db:"user_name,size:255"`
	JoinTime                time.Time `db:"join_time"`
}

func (repo *UnconditionalGiveawayRepo) GetGiveawayForGuild(ctx context.Context, guildId string) (UnconditionalGiveaway, error) {
	var giveaway UnconditionalGiveaway
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, info_message_id FROM unconditional_giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil {
		return UnconditionalGiveaway{}, err
	}

	return giveaway, nil
}

func (repo *UnconditionalGiveawayRepo) GetUnfinishedGiveaways(ctx context.Context) ([]UnconditionalGiveaway, error) {
	var giveaways []UnconditionalGiveaway
	_, err := repo.mysql.WithContext(ctx).Select(&giveaways, "SELECT id, start_time, end_time, guild_id, info_message_id FROM unconditional_giveaways WHERE end_time IS NULL")
	if err != nil {
		return nil, err
	}

	return giveaways, nil
}

func (repo *UnconditionalGiveawayRepo) GetParticipantsForGiveaway(ctx context.Context, giveawayId int) ([]UnconditionalGiveawayParticipant, error) {
	var participants []UnconditionalGiveawayParticipant
	_, err := repo.mysql.WithContext(ctx).Select(&participants, "SELECT id, unconditional_giveaway_id, user_id, user_name, join_time FROM unconditional_participants WHERE unconditional_giveaway_id = ?", giveawayId)
	if err != nil {
		return nil, err
	}

	return participants, nil
}

func (repo *UnconditionalGiveawayRepo) CountParticipantsForGiveaway(ctx context.Context, giveawayId int) (int, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM unconditional_participants WHERE unconditional_giveaway_id = ?", giveawayId)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (repo *UnconditionalGiveawayRepo) InsertGiveaway(ctx context.Context, guildId, messageId string) error {
	giveaway := &UnconditionalGiveaway{
		StartTime:     time.Now(),
		GuildId:       guildId,
		InfoMessageId: messageId,
	}
	err := repo.mysql.WithContext(ctx).Insert(giveaway)
	if err != nil {
		return err
	}

	return nil
}

func (repo *UnconditionalGiveawayRepo) InsertParticipant(ctx context.Context, giveawayId int, userId string, userName string) error {
	participant := &UnconditionalGiveawayParticipant{
		UnconditionalGiveawayId: giveawayId,
		UserId:                  userId,
		UserName:                userName,
		JoinTime:                time.Now(),
	}
	err := repo.mysql.WithContext(ctx).Insert(participant)
	if err != nil {
		return err
	}

	return nil
}

func (repo *UnconditionalGiveawayRepo) InsertWinner(ctx context.Context, giveawayId int, userId, code string) error {
	winner := &UnconditionalGiveawayWinner{
		UnconditionalGiveawayId: giveawayId,
		UserId:                  userId,
		Code:                    code,
	}
	err := repo.mysql.WithContext(ctx).Insert(winner)
	if err != nil {
		return err
	}

	return nil
}

func (repo *UnconditionalGiveawayRepo) FinishGiveaway(ctx context.Context, giveaway *UnconditionalGiveaway, messageId string) error {
	now := time.Now()
	giveaway.EndTime = &now
	giveaway.InfoMessageId = messageId
	_, err := repo.mysql.WithContext(ctx).Update(giveaway)
	if err != nil {
		return err
	}

	return nil
}

func (repo *UnconditionalGiveawayRepo) HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error) {
	result, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM unconditional_giveaways g JOIN unconditional_giveaway_winners w ON g.id = w.unconditional_giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (repo *UnconditionalGiveawayRepo) GetCodeForInfoMessage(ctx context.Context, messageId, userId string) (string, error) {
	code, err := repo.mysql.WithContext(ctx).SelectStr("SELECT code FROM unconditional_giveaways g JOIN unconditional_giveaway_winners w ON g.id = w.unconditional_giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (repo *UnconditionalGiveawayRepo) GetGiveawayByMessageId(ctx context.Context, messageId string) (*UnconditionalGiveaway, error) {
	var giveaway UnconditionalGiveaway
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, info_message_id FROM unconditional_giveaways WHERE info_message_id = ?", messageId)
	if err != nil {
		return nil, err
	}

	return &giveaway, nil
}
