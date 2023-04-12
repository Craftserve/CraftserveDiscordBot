package repos

import (
	"context"
	"database/sql"
	"github.com/go-gorp/gorp"
	"time"
)

type GiveawayRepo struct {
	mysql *gorp.DbMap
}

func NewGiveawayRepo(mysql *gorp.DbMap) *GiveawayRepo {
	mysql.AddTableWithName(Giveaway{}, "giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(Participant{}, "participants").SetKeys(true, "id")
	mysql.AddTableWithName(ParticipantCandidate{}, "participant_candidates").SetKeys(true, "id")
	mysql.AddTableWithName(ThxNotification{}, "thx_notifications").SetKeys(true, "id")

	return &GiveawayRepo{mysql: mysql}
}

type Giveaway struct {
	Id            int            `db:"id, primarykey, autoincrement"`
	StartTime     time.Time      `db:"start_time"`
	EndTime       *time.Time     `db:"end_time"`
	GuildId       string         `db:"guild_id,size:255"`
	GuildName     string         `db:"guild_name,size:255"`
	WinnerId      sql.NullString `db:"winner_id,size:255"`
	WinnerName    sql.NullString `db:"winner_name,size:255"`
	InfoMessageId sql.NullString `db:"info_message_id,size:255"`
	Code          sql.NullString `db:"code,size:255"`
}

type Participant struct {
	Id           int            `db:"id, primarykey, autoincrement"`
	UserName     string         `db:"user_name,size:255"`
	UserId       string         `db:"user_id,size:255"`
	GiveawayId   int            `db:"giveaway_id"`
	CreateTime   time.Time      `db:"create_time"`
	GuildName    string         `db:"guild_name,size:255"`
	GuildId      string         `db:"guild_id,size:255"`
	MessageId    string         `db:"message_id,size:255"`
	ChannelId    string         `db:"channel_id,size:255"`
	IsAccepted   sql.NullBool   `db:"is_accepted"`
	AcceptTime   *time.Time     `db:"accept_time"`
	AcceptUser   sql.NullString `db:"accept_user,size:255"`
	AcceptUserId sql.NullString `db:"accept_user_id,size:255"`
}

type ParticipantCandidate struct {
	Id                    int          `db:"id, primarykey, autoincrement"`
	CandidateId           string       `db:"candidate_id,size:255"`
	CandidateName         string       `db:"candidate_name,size:255"`
	CandidateApproverId   string       `db:"candidate_approver_id,size:255"`
	CandidateApproverName string       `db:"candidate_approver_name,size:255"`
	GuildId               string       `db:"guild_id,size:255"`
	GuildName             string       `db:"guild_name,size:255"`
	MessageId             string       `db:"message_id,size:255"`
	ChannelId             string       `db:"channel_id,size:255"`
	IsAccepted            sql.NullBool `db:"is_accepted"`
	AcceptTime            *time.Time   `db:"accept_time"`
}

type ThxNotification struct {
	Id                       int    `db:"id,primarykey,autoincrement"`
	MessageId                string `db:"message_id,size:255"`
	ThxNotificationMessageId string `db:"thx_notification_message_id,size:255"`
}

type ParticipantWithThxAmount struct {
	UserId    string `db:"user_id,size:255"`
	ThxAmount int    `db:"amount"`
}

func (repo *GiveawayRepo) GetGiveawayForGuild(ctx context.Context, guildId string) (Giveaway, error) {
	var giveaway Giveaway
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, guild_name, winner_id, winner_name, info_message_id, code FROM giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil {
		return Giveaway{}, err
	}
	return giveaway, nil
}

func (repo *GiveawayRepo) GetParticipantNamesForGiveaway(ctx context.Context, giveawayId int) ([]string, error) {
	var participants []Participant
	_, err := repo.mysql.WithContext(ctx).Select(&participants, "SELECT user_name FROM participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(participants))
	for i := range participants {
		names[i] = participants[i].UserName
	}
	return names, nil
}

func (repo *GiveawayRepo) InsertGiveaway(ctx context.Context, guildId string, guildName string) error {
	giveaway := &Giveaway{
		StartTime: time.Now(),
		GuildId:   guildId,
		GuildName: guildName,
	}
	err := repo.mysql.WithContext(ctx).Insert(giveaway)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) InsertParticipant(ctx context.Context, giveawayId int, guildId, guildName, userId, userName, channelId, messageId string) error {
	participant := &Participant{
		UserId:     userId,
		GiveawayId: giveawayId,
		CreateTime: time.Now(),
		GuildId:    guildId,
		ChannelId:  channelId,
		GuildName:  guildName,
		UserName:   userName,
		MessageId:  messageId,
	}
	err := repo.mysql.WithContext(ctx).Insert(participant)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) InsertParticipantCandidate(ctx context.Context, guildId, guildName, candidateId, candidateName, approverId, approverName, channelId, messageId string) error {
	participantCandidate := &ParticipantCandidate{
		CandidateName:         candidateName,
		CandidateId:           candidateId,
		CandidateApproverName: approverName,
		CandidateApproverId:   approverId,
		GuildName:             guildName,
		GuildId:               guildId,
		ChannelId:             channelId,
		MessageId:             messageId,
	}
	err := repo.mysql.WithContext(ctx).Insert(participantCandidate)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) GetParticipantsWithThxAmount(ctx context.Context, guildId string, minThxAmount int) ([]ParticipantWithThxAmount, error) {
	var helpers []ParticipantWithThxAmount
	_, err := repo.mysql.WithContext(ctx).Select(&helpers, "SELECT user_id, amount FROM (SELECT user_id, count(*) AS amount FROM participants WHERE guild_id=? AND is_accepted=1 GROUP BY user_id) AS a WHERE amount > ?", guildId, minThxAmount)
	if err != nil {
		return nil, err
	}
	return helpers, nil
}

func (repo *GiveawayRepo) HasThxAmount(ctx context.Context, guildId, memberId string, minThxAmount int) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT count(*) AS amount  FROM participants WHERE guild_id=? AND user_id=? AND is_accepted=1 HAVING amount > ?", guildId, memberId, minThxAmount)
	if err != nil {
		return false, err
	}
	return ret > 0, nil
}

func (repo *GiveawayRepo) GetParticipantsForGiveaway(ctx context.Context, giveawayId int) ([]Participant, error) {
	var participants []Participant
	_, err := repo.mysql.WithContext(ctx).Select(&participants, "SELECT id, user_name, user_id, giveaway_id, create_time, guild_name, guild_id, message_id, channel_id, is_accepted, accept_time, accept_user, accept_user_id FROM participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		return nil, err
	}
	return participants, nil
}

func (repo *GiveawayRepo) GetThxNotification(ctx context.Context, messageId string) (ThxNotification, error) {
	var notification ThxNotification
	err := repo.mysql.WithContext(ctx).SelectOne(&notification, "SELECT id, message_id, thx_notification_message_id FROM thx_notifications WHERE message_id = ?", messageId)
	if err != nil {
		return ThxNotification{}, err
	}
	return notification, nil
}

func (repo *GiveawayRepo) InsertThxNotification(ctx context.Context, thxMessageId string, notificationMessageId string) error {
	notification := &ThxNotification{
		MessageId:                thxMessageId,
		ThxNotificationMessageId: notificationMessageId,
	}
	err := repo.mysql.WithContext(ctx).Insert(notification)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) IsThxMessage(ctx context.Context, messageId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT count(*) FROM participants WHERE message_id = ?", messageId)
	if err != nil {
		return false, err
	}

	return ret > 0, nil
}

func (repo *GiveawayRepo) IsThxmeMessage(ctx context.Context, messageId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT count(*) FROM participant_candidates WHERE message_id = ?", messageId)
	if err != nil {
		return false, err
	}

	return ret > 0, nil
}

func (repo *GiveawayRepo) GetParticipant(ctx context.Context, messageId string) (Participant, error) {
	var participant Participant
	err := repo.mysql.WithContext(ctx).SelectOne(&participant, "SELECT * FROM participants WHERE message_id = ?", messageId)
	if err != nil {
		return Participant{}, err
	}
	return participant, nil
}

func (repo *GiveawayRepo) UpdateParticipant(ctx context.Context, participant *Participant, acceptUserId, acceptUsername string, isAccepted bool) error {
	now := time.Now()
	participant.AcceptTime = &now
	participant.IsAccepted = sql.NullBool{Bool: isAccepted, Valid: true}
	participant.AcceptUserId = sql.NullString{String: acceptUserId, Valid: true}
	participant.AcceptUser = sql.NullString{String: acceptUsername, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(participant)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) GetParticipantCandidate(ctx context.Context, messageId string) (ParticipantCandidate, error) {
	var participantCandidate ParticipantCandidate
	err := repo.mysql.WithContext(ctx).SelectOne(&participantCandidate, "SELECT id, candidate_id, candidate_name, candidate_approver_id, candidate_approver_name, guild_id, guild_name, message_id, channel_id, is_accepted, accept_time FROM participant_candidates WHERE message_id = ?", messageId)
	if err != nil {
		return ParticipantCandidate{}, err
	}
	return participantCandidate, nil
}

func (repo *GiveawayRepo) UpdateParticipantCandidate(ctx context.Context, participantCandidate *ParticipantCandidate, isAccepted bool) error {
	now := time.Now()
	participantCandidate.AcceptTime = &now
	participantCandidate.IsAccepted = sql.NullBool{Bool: isAccepted, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(participantCandidate)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) UpdateGiveaway(ctx context.Context, giveaway *Giveaway, messageId, code, winnerId, winnerName string) error {
	now := time.Now()
	giveaway.EndTime = &now
	giveaway.InfoMessageId = sql.NullString{String: messageId, Valid: true}
	giveaway.Code = sql.NullString{String: code, Valid: true}
	giveaway.WinnerId = sql.NullString{String: winnerId, Valid: true}
	giveaway.WinnerName = sql.NullString{String: winnerName, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(giveaway)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) GetUnfinishedGiveaways(ctx context.Context) ([]Giveaway, error) {
	var giveaways []Giveaway
	_, err := repo.mysql.WithContext(ctx).Select(&giveaways, "SELECT id, start_time, end_time, guild_id, guild_name, winner_id, winner_name, info_message_id, code FROM giveaways WHERE end_time IS NULL")
	if err != nil {
		return nil, err
	}
	return giveaways, nil
}

func (repo *GiveawayRepo) RemoveAllParticipantEntries(ctx context.Context, giveawayId int, participantId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("UPDATE participants SET is_accepted=false WHERE giveaway_id = ? AND user_id = ?", giveawayId, participantId)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT count(*) FROM giveaways WHERE info_message_id = ? AND winner_id = ?", messageId, userId)
	if err != nil {
		return false, err
	}

	return ret > 0, nil
}

func (repo *GiveawayRepo) GetCodeForInfoMessage(ctx context.Context, infoMessageId string) (string, error) {
	var code string
	err := repo.mysql.WithContext(ctx).SelectOne(&code, "SELECT code FROM giveaways WHERE info_message_id = ?", infoMessageId)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (repo *GiveawayRepo) GetLastCodesForUser(ctx context.Context, userId string, limit int) ([]string, error) {
	var codes []string
	_, err := repo.mysql.WithContext(ctx).Select(&codes, "SELECT code FROM giveaways WHERE winner_id = ? ORDER BY end_time DESC LIMIT ?", userId, limit)
	if err != nil {
		return nil, err
	}
	return codes, nil
}
