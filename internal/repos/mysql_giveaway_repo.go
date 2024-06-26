package repos

import (
	"context"
	"csrvbot/domain/entities"
	"database/sql"
	"github.com/go-gorp/gorp"
	"time"
)

type GiveawayRepo struct {
	mysql *gorp.DbMap
}

func NewGiveawayRepo(mysql *gorp.DbMap) *GiveawayRepo {
	mysql.AddTableWithName(SqlGiveaway{}, "giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(SqlParticipant{}, "participants").SetKeys(true, "id")
	mysql.AddTableWithName(SqlParticipantCandidate{}, "participant_candidates").SetKeys(true, "id")
	mysql.AddTableWithName(SqlThxNotification{}, "thx_notifications").SetKeys(true, "id")

	return &GiveawayRepo{mysql: mysql}
}

type SqlGiveaway struct {
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

type SqlParticipant struct {
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

type SqlParticipantCandidate struct {
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

type SqlThxNotification struct {
	Id                    int    `db:"id,primarykey,autoincrement"`
	ThxMessageId          string `db:"thx_message_id,size:255"`
	NotificationMessageId string `db:"notification_message_id,size:255"`
}

type SqlParticipantWithThxAmount struct {
	UserId    string `db:"user_id,size:255"`
	ThxAmount int    `db:"amount"`
}

func FromSqlGiveaway(giveaway *SqlGiveaway) *entities.Giveaway {
	return &entities.Giveaway{
		Id:            giveaway.Id,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		GuildName:     giveaway.GuildName,
		WinnerId:      giveaway.WinnerId,
		WinnerName:    giveaway.WinnerName,
		InfoMessageId: giveaway.InfoMessageId,
		Code:          giveaway.Code,
	}
}

func ToSqlGiveaway(giveaway *entities.Giveaway) *SqlGiveaway {
	return &SqlGiveaway{
		Id:            giveaway.Id,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		GuildName:     giveaway.GuildName,
		WinnerId:      giveaway.WinnerId,
		WinnerName:    giveaway.WinnerName,
		InfoMessageId: giveaway.InfoMessageId,
		Code:          giveaway.Code,
	}
}

func FromSqlParticipant(participant *SqlParticipant) *entities.Participant {
	return &entities.Participant{
		Id:           participant.Id,
		UserName:     participant.UserName,
		UserId:       participant.UserId,
		GiveawayId:   participant.GiveawayId,
		CreateTime:   participant.CreateTime,
		GuildName:    participant.GuildName,
		GuildId:      participant.GuildId,
		MessageId:    participant.MessageId,
		ChannelId:    participant.ChannelId,
		IsAccepted:   participant.IsAccepted,
		AcceptTime:   participant.AcceptTime,
		AcceptUser:   participant.AcceptUser,
		AcceptUserId: participant.AcceptUserId,
	}
}

func ToSqlParticipant(participant *entities.Participant) *SqlParticipant {
	return &SqlParticipant{
		Id:           participant.Id,
		UserName:     participant.UserName,
		UserId:       participant.UserId,
		GiveawayId:   participant.GiveawayId,
		CreateTime:   participant.CreateTime,
		GuildName:    participant.GuildName,
		GuildId:      participant.GuildId,
		MessageId:    participant.MessageId,
		ChannelId:    participant.ChannelId,
		IsAccepted:   participant.IsAccepted,
		AcceptTime:   participant.AcceptTime,
		AcceptUser:   participant.AcceptUser,
		AcceptUserId: participant.AcceptUserId,
	}
}

func FromSqlParticipantCandidate(participantCandidate *SqlParticipantCandidate) *entities.ParticipantCandidate {
	return &entities.ParticipantCandidate{
		Id:                    participantCandidate.Id,
		CandidateId:           participantCandidate.CandidateId,
		CandidateName:         participantCandidate.CandidateName,
		CandidateApproverId:   participantCandidate.CandidateApproverId,
		CandidateApproverName: participantCandidate.CandidateApproverName,
		GuildId:               participantCandidate.GuildId,
		GuildName:             participantCandidate.GuildName,
		MessageId:             participantCandidate.MessageId,
		ChannelId:             participantCandidate.ChannelId,
		IsAccepted:            participantCandidate.IsAccepted,
		AcceptTime:            participantCandidate.AcceptTime,
	}
}

func ToSqlParticipantCandidate(participantCandidate *entities.ParticipantCandidate) *SqlParticipantCandidate {
	return &SqlParticipantCandidate{
		Id:                    participantCandidate.Id,
		CandidateId:           participantCandidate.CandidateId,
		CandidateName:         participantCandidate.CandidateName,
		CandidateApproverId:   participantCandidate.CandidateApproverId,
		CandidateApproverName: participantCandidate.CandidateApproverName,
		GuildId:               participantCandidate.GuildId,
		GuildName:             participantCandidate.GuildName,
		MessageId:             participantCandidate.MessageId,
		ChannelId:             participantCandidate.ChannelId,
		IsAccepted:            participantCandidate.IsAccepted,
		AcceptTime:            participantCandidate.AcceptTime,
	}
}

func FromSqlThxNotification(thxNotification *SqlThxNotification) *entities.ThxNotification {
	return &entities.ThxNotification{
		Id:                    thxNotification.Id,
		ThxMessageId:          thxNotification.ThxMessageId,
		NotificationMessageId: thxNotification.NotificationMessageId,
	}
}

func ToSqlThxNotification(thxNotification *entities.ThxNotification) *SqlThxNotification {
	return &SqlThxNotification{
		Id:                    thxNotification.Id,
		ThxMessageId:          thxNotification.ThxMessageId,
		NotificationMessageId: thxNotification.NotificationMessageId,
	}
}

func FromSqlParticipantWithThxAmount(participantWithThxAmount *SqlParticipantWithThxAmount) *entities.ParticipantWithThxAmount {
	return &entities.ParticipantWithThxAmount{
		UserId:    participantWithThxAmount.UserId,
		ThxAmount: participantWithThxAmount.ThxAmount,
	}
}

func ToSqlParticipantWithThxAmount(participantWithThxAmount *entities.ParticipantWithThxAmount) *SqlParticipantWithThxAmount {
	return &SqlParticipantWithThxAmount{
		UserId:    participantWithThxAmount.UserId,
		ThxAmount: participantWithThxAmount.ThxAmount,
	}
}

func (repo *GiveawayRepo) GetGiveawayForGuild(ctx context.Context, guildId string) (entities.Giveaway, error) {
	var giveaway SqlGiveaway
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, guild_name, winner_id, winner_name, info_message_id, code FROM giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil {
		return entities.Giveaway{}, err
	}
	return *FromSqlGiveaway(&giveaway), nil
}

func (repo *GiveawayRepo) GetParticipantNamesForGiveaway(ctx context.Context, giveawayId int) ([]string, error) {
	var participants []SqlParticipant
	_, err := repo.mysql.WithContext(ctx).Select(&participants, "SELECT user_name FROM participants WHERE giveaway_id = ? AND is_accepted = TRUE", giveawayId)
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
	giveaway := &SqlGiveaway{
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
	participant := &SqlParticipant{
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
	participantCandidate := &SqlParticipantCandidate{
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

func (repo *GiveawayRepo) GetParticipantsWithThxAmount(ctx context.Context, guildId string, minThxAmount int) ([]entities.ParticipantWithThxAmount, error) {
	var helpers []SqlParticipantWithThxAmount
	_, err := repo.mysql.WithContext(ctx).Select(&helpers, "SELECT user_id, amount FROM (SELECT user_id, COUNT(*) AS amount FROM participants WHERE guild_id=? AND is_accepted=1 GROUP BY user_id) AS a WHERE amount > ?", guildId, minThxAmount)
	if err != nil {
		return nil, err
	}

	var result []entities.ParticipantWithThxAmount
	for _, helper := range helpers {
		result = append(result, *FromSqlParticipantWithThxAmount(&helper))
	}

	return result, nil
}

func (repo *GiveawayRepo) HasThxAmount(ctx context.Context, guildId, memberId string, minThxAmount int) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) AS amount  FROM participants WHERE guild_id=? AND user_id=? AND is_accepted=1 HAVING amount > ?", guildId, memberId, minThxAmount)
	if err != nil {
		return false, err
	}
	return ret > 0, nil
}

func (repo *GiveawayRepo) GetParticipantsForGiveaway(ctx context.Context, giveawayId int) ([]entities.Participant, error) {
	var participants []SqlParticipant
	_, err := repo.mysql.WithContext(ctx).Select(&participants, "SELECT id, user_name, user_id, giveaway_id, create_time, guild_name, guild_id, message_id, channel_id, is_accepted, accept_time, accept_user, accept_user_id FROM participants WHERE giveaway_id = ? AND is_accepted = TRUE", giveawayId)
	if err != nil {
		return nil, err
	}

	var result []entities.Participant
	for _, participant := range participants {
		result = append(result, *FromSqlParticipant(&participant))
	}

	return result, nil
}

func (repo *GiveawayRepo) GetThxNotification(ctx context.Context, messageId string) (entities.ThxNotification, error) {
	var notification SqlThxNotification
	err := repo.mysql.WithContext(ctx).SelectOne(&notification, "SELECT id, thx_message_id, notification_message_id FROM thx_notifications WHERE thx_message_id = ?", messageId)
	if err != nil {
		return entities.ThxNotification{}, err
	}
	return *FromSqlThxNotification(&notification), nil
}

func (repo *GiveawayRepo) InsertThxNotification(ctx context.Context, thxMessageId, notificationMessageId string) error {
	notification := &SqlThxNotification{
		ThxMessageId:          thxMessageId,
		NotificationMessageId: notificationMessageId,
	}
	err := repo.mysql.WithContext(ctx).Insert(notification)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) IsThxMessage(ctx context.Context, messageId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM participants WHERE message_id = ?", messageId)
	if err != nil {
		return false, err
	}

	return ret > 0, nil
}

func (repo *GiveawayRepo) IsThxmeMessage(ctx context.Context, messageId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM participant_candidates WHERE message_id = ?", messageId)
	if err != nil {
		return false, err
	}

	return ret > 0, nil
}

func (repo *GiveawayRepo) GetParticipant(ctx context.Context, messageId string) (entities.Participant, error) {
	var participant SqlParticipant
	err := repo.mysql.WithContext(ctx).SelectOne(&participant, "SELECT * FROM participants WHERE message_id = ?", messageId)
	if err != nil {
		return entities.Participant{}, err
	}
	return *FromSqlParticipant(&participant), nil
}

func (repo *GiveawayRepo) UpdateParticipant(ctx context.Context, participant *entities.Participant, acceptUserId, acceptUsername string, isAccepted bool) error {
	now := time.Now()
	participant.AcceptTime = &now
	participant.IsAccepted = sql.NullBool{Bool: isAccepted, Valid: true}
	participant.AcceptUserId = sql.NullString{String: acceptUserId, Valid: true}
	participant.AcceptUser = sql.NullString{String: acceptUsername, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(ToSqlParticipant(participant))
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) GetParticipantCandidate(ctx context.Context, messageId string) (entities.ParticipantCandidate, error) {
	var participantCandidate SqlParticipantCandidate
	err := repo.mysql.WithContext(ctx).SelectOne(&participantCandidate, "SELECT id, candidate_id, candidate_name, candidate_approver_id, candidate_approver_name, guild_id, guild_name, message_id, channel_id, is_accepted, accept_time FROM participant_candidates WHERE message_id = ?", messageId)
	if err != nil {
		return entities.ParticipantCandidate{}, err
	}
	return *FromSqlParticipantCandidate(&participantCandidate), nil
}

func (repo *GiveawayRepo) UpdateParticipantCandidate(ctx context.Context, participantCandidate *entities.ParticipantCandidate, isAccepted bool) error {
	now := time.Now()
	participantCandidate.AcceptTime = &now
	participantCandidate.IsAccepted = sql.NullBool{Bool: isAccepted, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(ToSqlParticipantCandidate(participantCandidate))
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) FinishGiveaway(ctx context.Context, giveaway *entities.Giveaway, messageId, code, winnerId, winnerName string) error {
	now := time.Now()
	giveaway.EndTime = &now
	giveaway.InfoMessageId = sql.NullString{String: messageId, Valid: true}
	giveaway.Code = sql.NullString{String: code, Valid: true}
	giveaway.WinnerId = sql.NullString{String: winnerId, Valid: true}
	giveaway.WinnerName = sql.NullString{String: winnerName, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(ToSqlGiveaway(giveaway))
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) GetUnfinishedGiveaways(ctx context.Context) ([]entities.Giveaway, error) {
	var giveaways []SqlGiveaway
	_, err := repo.mysql.WithContext(ctx).Select(&giveaways, "SELECT id, start_time, end_time, guild_id, guild_name, winner_id, winner_name, info_message_id, code FROM giveaways WHERE end_time IS NULL")
	if err != nil {
		return nil, err
	}

	var result []entities.Giveaway
	for _, giveaway := range giveaways {
		result = append(result, *FromSqlGiveaway(&giveaway))

	}

	return result, nil
}

func (repo *GiveawayRepo) RemoveAllParticipantEntries(ctx context.Context, giveawayId int, participantId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("UPDATE participants SET is_accepted=FALSE WHERE giveaway_id = ? AND user_id = ?", giveawayId, participantId)
	if err != nil {
		return err
	}
	return nil
}

func (repo *GiveawayRepo) HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM giveaways WHERE info_message_id = ? AND winner_id = ?", messageId, userId)
	if err != nil {
		return false, err
	}

	return ret > 0, nil
}

func (repo *GiveawayRepo) IsGiveawayEnded(ctx context.Context, giveawayId int) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM giveaways WHERE id = ? AND end_time IS NOT NULL", giveawayId)
	if err != nil {
		return false, err
	}

	return ret > 0, nil
}

func (repo *GiveawayRepo) GetGiveawayById(ctx context.Context, giveawayId int) (*entities.Giveaway, error) {
	var giveaway SqlGiveaway
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, start_time, end_time, guild_id, guild_name, winner_id, winner_name, info_message_id, code FROM giveaways WHERE id = ?", giveawayId)
	if err != nil {
		return nil, err
	}

	return FromSqlGiveaway(&giveaway), nil
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
