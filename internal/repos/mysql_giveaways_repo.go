package repos

import (
	"context"
	"csrvbot/domain/entities"
	"database/sql"
	"github.com/go-gorp/gorp"
	"time"
)

type GiveawaysRepo struct {
	mysql *gorp.DbMap
}

type SqlGiveaways struct {
	Id            int        `db:"id, primarykey, autoincrement"`
	Type          string     `db:"type, size:255"`
	StartTime     time.Time  `db:"start_time"`
	EndTime       *time.Time `db:"end_time"`
	GuildId       string     `db:"guild_id, size:255"`
	InfoMessageId *string    `db:"info_message_id, size:255"`
	Level         *int       `db:"level"`
}

type SqlGiveawaysParticipant struct {
	Id           int            `db:"id, primarykey, autoincrement"`
	GiveawayId   int            `db:"giveaway_id"`
	GuildId      string         `db:"guild_id,size:255"`
	UserId       string         `db:"user_id, size:255"`
	UserName     string         `db:"user_name, size:255"`
	JoinTime     time.Time      `db:"join_time"`
	UserLevel    *int           `db:"user_level"`
	MessageId    *string        `db:"message_id"`
	ChannelId    *string        `db:"channel_id,size:255"`
	IsAccepted   sql.NullBool   `db:"is_accepted"`
	AcceptTime   *time.Time     `db:"accept_time"`
	AcceptUser   sql.NullString `db:"accept_user,size:255"`
	AcceptUserId sql.NullString `db:"accept_user_id,size:255"`
}

type SqlGiveawaysWinner struct {
	Id         int    `db:"id,primarykey, autoincrement"`
	GiveawayId int    `db:"giveaway_id"`
	UserId     string `db:"user_id, size:255"`
	Code       string `db:"code, size:255"`
}

type SqlThxParticipantCandidate struct {
	Id                    int          `db:"id, primarykey, autoincrement"`
	CandidateId           string       `db:"candidate_id, size:255"`            // User ID
	CandidateName         string       `db:"candidate_name, size:255"`          // User name
	CandidateApproverId   string       `db:"candidate_approver_id, size:255"`   // Approver User ID
	CandidateApproverName string       `db:"candidate_approver_name, size:255"` // Approver User name
	GiveawayId            int          `db:"giveaway_id"`
	GuildId               string       `db:"guild_id, size:255"`
	GuildName             string       `db:"guild_name, size:255"`
	MessageId             string       `db:"message_id, size:255"`
	ChannelId             string       `db:"channel_id, size:255"`
	IsAccepted            sql.NullBool `db:"is_accepted"`
	AcceptTime            *time.Time   `db:"accept_time"`
}

type SqlThxNotification struct {
	Id                    int    `db:"id, primarykey, autoincrement"`
	ThxMessageId          string `db:"thx_message_id, size:255"`
	NotificationMessageId string `db:"notification_message_id, size:255"`
}

type SqlThxParticipantWithThxAmount struct {
	UserId    string `db:"user_id, size:255"`
	ThxAmount int    `db:"amount"`
}

type SqlDailyUserMessages struct {
	Id      int    `db:"id, primarykey, autoincrement"`
	UserId  string `db:"user_id,size:255"`
	Day     string `db:"day"` // fixme should be date
	GuildId string `db:"guild_id,size:255"`
	Count   int    `db:"count"`
}

func FromSqlGiveaways(giveaway *SqlGiveaways) *entities.Giveaway {
	return &entities.Giveaway{
		Id:            giveaway.Id,
		Type:          giveaway.Type,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		InfoMessageId: giveaway.InfoMessageId,
		Level:         giveaway.Level,
	}
}

func ToSqlGiveaways(giveaway *entities.Giveaway) *SqlGiveaways {
	return &SqlGiveaways{
		Id:            giveaway.Id,
		Type:          giveaway.Type,
		StartTime:     giveaway.StartTime,
		EndTime:       giveaway.EndTime,
		GuildId:       giveaway.GuildId,
		InfoMessageId: giveaway.InfoMessageId,
		Level:         giveaway.Level,
	}
}

func FromSqlGiveawaysParticipant(participant *SqlGiveawaysParticipant) *entities.GiveawayParticipant {
	return &entities.GiveawayParticipant{
		Id:           participant.Id,
		GiveawayId:   participant.GiveawayId,
		GuildId:      participant.GuildId,
		UserId:       participant.UserId,
		UserName:     participant.UserName,
		JoinTime:     participant.JoinTime,
		UserLevel:    participant.UserLevel,
		MessageId:    participant.MessageId,
		ChannelId:    participant.ChannelId,
		IsAccepted:   participant.IsAccepted,
		AcceptTime:   participant.AcceptTime,
		AcceptUser:   participant.AcceptUser,
		AcceptUserId: participant.AcceptUserId,
	}
}

func ToSqlGiveawaysParticipant(participant *entities.GiveawayParticipant) *SqlGiveawaysParticipant {
	return &SqlGiveawaysParticipant{
		Id:           participant.Id,
		GiveawayId:   participant.GiveawayId,
		GuildId:      participant.GuildId,
		UserId:       participant.UserId,
		UserName:     participant.UserName,
		JoinTime:     participant.JoinTime,
		UserLevel:    participant.UserLevel,
		MessageId:    participant.MessageId,
		ChannelId:    participant.ChannelId,
		IsAccepted:   participant.IsAccepted,
		AcceptTime:   participant.AcceptTime,
		AcceptUser:   participant.AcceptUser,
		AcceptUserId: participant.AcceptUserId,
	}
}

func FromSqlGiveawaysWinner(winner *SqlGiveawaysWinner) *entities.GiveawayWinner {
	return &entities.GiveawayWinner{
		Id:         winner.Id,
		GiveawayId: winner.GiveawayId,
		UserId:     winner.UserId,
		Code:       winner.Code,
	}
}

func ToSqlGiveawaysWinner(winner *entities.GiveawayWinner) *SqlGiveawaysWinner {
	return &SqlGiveawaysWinner{
		Id:         winner.Id,
		GiveawayId: winner.GiveawayId,
		UserId:     winner.UserId,
		Code:       winner.Code,
	}
}

func FromSqlThxParticipantCandidate(candidate *SqlThxParticipantCandidate) *entities.ThxParticipantCandidate {
	return &entities.ThxParticipantCandidate{
		Id:                    candidate.Id,
		CandidateId:           candidate.CandidateId,
		CandidateName:         candidate.CandidateName,
		CandidateApproverId:   candidate.CandidateApproverId,
		CandidateApproverName: candidate.CandidateApproverName,
		GiveawayId:            candidate.GiveawayId,
		GuildId:               candidate.GuildId,
		GuildName:             candidate.GuildName,
		MessageId:             candidate.MessageId,
		ChannelId:             candidate.ChannelId,
		IsAccepted:            candidate.IsAccepted,
		AcceptTime:            candidate.AcceptTime,
	}
}

func ToSqlThxParticipantCandidate(candidate *entities.ThxParticipantCandidate) *SqlThxParticipantCandidate {
	return &SqlThxParticipantCandidate{
		Id:                    candidate.Id,
		CandidateId:           candidate.CandidateId,
		CandidateName:         candidate.CandidateName,
		CandidateApproverId:   candidate.CandidateApproverId,
		CandidateApproverName: candidate.CandidateApproverName,
		GiveawayId:            candidate.GiveawayId,
		GuildId:               candidate.GuildId,
		GuildName:             candidate.GuildName,
		MessageId:             candidate.MessageId,
		ChannelId:             candidate.ChannelId,
		IsAccepted:            candidate.IsAccepted,
		AcceptTime:            candidate.AcceptTime,
	}
}

func FromSqlThxNotification(notification *SqlThxNotification) *entities.ThxNotification {
	return &entities.ThxNotification{
		Id:                    notification.Id,
		ThxMessageId:          notification.ThxMessageId,
		NotificationMessageId: notification.NotificationMessageId,
	}
}

func ToSqlThxNotification(notification *entities.ThxNotification) *SqlThxNotification {
	return &SqlThxNotification{
		Id:                    notification.Id,
		ThxMessageId:          notification.ThxMessageId,
		NotificationMessageId: notification.NotificationMessageId,
	}
}

func ToSqlThxParticipantWithThxAmount(participant *entities.ThxParticipantWithThxAmount) *SqlThxParticipantWithThxAmount {
	return &SqlThxParticipantWithThxAmount{
		UserId:    participant.UserId,
		ThxAmount: participant.ThxAmount,
	}
}

func FromSqlThxParticipantWithThxAmount(participant *SqlThxParticipantWithThxAmount) *entities.ThxParticipantWithThxAmount {
	return &entities.ThxParticipantWithThxAmount{
		UserId:    participant.UserId,
		ThxAmount: participant.ThxAmount,
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

func FromSqlDailyUserMessages(dailyUserMessages *SqlDailyUserMessages) *entities.DailyUserMessages {
	return &entities.DailyUserMessages{
		Id:      dailyUserMessages.Id,
		UserId:  dailyUserMessages.UserId,
		Day:     dailyUserMessages.Day,
		GuildId: dailyUserMessages.GuildId,
		Count:   dailyUserMessages.Count,
	}
}

func NewGiveawaysRepo(mysql *gorp.DbMap) *GiveawaysRepo {
	mysql.AddTableWithName(SqlGiveaways{}, "giveaways").SetKeys(true, "id")
	mysql.AddTableWithName(SqlGiveawaysParticipant{}, "giveaway_participants").SetKeys(true, "id")
	mysql.AddTableWithName(SqlThxParticipantCandidate{}, "thx_participant_candidates").SetKeys(true, "id")
	mysql.AddTableWithName(SqlGiveawaysWinner{}, "giveaway_winners").SetKeys(true, "id")
	mysql.AddTableWithName(SqlThxNotification{}, "thx_notifications").SetKeys(true, "id")
	mysql.AddTableWithName(SqlDailyUserMessages{}, "daily_user_messages").SetKeys(true, "id").SetUniqueTogether("day", "user_id", "guild_id")

	return &GiveawaysRepo{mysql: mysql}
}

func (repo GiveawaysRepo) GetGiveawayForGuild(ctx context.Context, guildId, giveawayType string) (*entities.Giveaway, error) {
	var giveaway SqlGiveaways
	err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, type, start_time, end_time, guild_id, info_message_id, level FROM giveaways WHERE end_time IS NULL AND guild_id = ? AND type = ?", guildId, giveawayType)
	if err != nil {
		return nil, err
	}

	return FromSqlGiveaways(&giveaway), nil
}

func (repo GiveawaysRepo) GetUnfinishedGiveaways(ctx context.Context, giveawayType string) (result []entities.Giveaway, err error) {
	var giveaways []SqlGiveaways
	_, err = repo.mysql.WithContext(ctx).Select(&giveaways, "SELECT id, type, start_time, end_time, guild_id, info_message_id, level FROM giveaways WHERE end_time IS NULL AND type = ?", giveawayType)
	if err != nil {
		return nil, err
	}

	for _, giveaway := range giveaways {
		result = append(result, *FromSqlGiveaways(&giveaway))
	}

	return result, nil
}

func (repo GiveawaysRepo) GetParticipantsForGiveaway(ctx context.Context, giveawayId int, accepted *bool) (result []entities.GiveawayParticipant, err error) {
	var participants []SqlGiveawaysParticipant
	if accepted == nil {
		_, err = repo.mysql.WithContext(ctx).Select(&participants, "SELECT id, giveaway_id, guild_id, user_id, user_name, join_time, user_level, message_id, channel_id, is_accepted, accept_time, accept_user, accept_user_id FROM giveaway_participants WHERE giveaway_id = ?", giveawayId)
	} else {
		_, err = repo.mysql.WithContext(ctx).Select(&participants, "SELECT id, giveaway_id, guild_id, user_id, user_name, join_time, user_level, message_id, channel_id, is_accepted, accept_time, accept_user, accept_user_id FROM giveaway_participants WHERE giveaway_id = ? AND is_accepted = ?", giveawayId, *accepted)
	}

	if err != nil {
		return nil, err
	}

	for _, participant := range participants {
		result = append(result, *FromSqlGiveawaysParticipant(&participant))
	}

	return result, nil
}

func (repo GiveawaysRepo) CountParticipantsForGiveaway(ctx context.Context, giveawayId int) (int, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM giveaway_participants WHERE giveaway_id = ?", giveawayId)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (repo GiveawaysRepo) InsertGiveaway(ctx context.Context, guildId string, messageId *string, giveawayType string, level *int) error {
	giveaway := &SqlGiveaways{
		Type:          giveawayType,
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

func (repo GiveawaysRepo) InsertParticipant(ctx context.Context, giveawayId, level int, guildId, userId, userName string, messageId, channelId *string) error {
	participant := &SqlGiveawaysParticipant{
		GiveawayId: giveawayId,
		GuildId:    guildId,
		UserId:     userId,
		UserName:   userName,
		JoinTime:   time.Now(),
		UserLevel:  &level,
		MessageId:  messageId,
		ChannelId:  channelId,
	}
	if err := repo.mysql.WithContext(ctx).Insert(participant); err != nil {
		return err
	}

	return nil
}

func (repo GiveawaysRepo) InsertWinner(ctx context.Context, giveawayId int, userId, code string) error {
	winner := &SqlGiveawaysWinner{
		GiveawayId: giveawayId,
		UserId:     userId,
		Code:       code,
	}
	if err := repo.mysql.WithContext(ctx).Insert(winner); err != nil {
		return err
	}

	return nil
}

func (repo GiveawaysRepo) FinishGiveaway(ctx context.Context, giveaway *entities.Giveaway, messageId *string) error {
	now := time.Now()
	giveaway.EndTime = &now
	giveaway.InfoMessageId = messageId
	if _, err := repo.mysql.WithContext(ctx).Update(ToSqlGiveaways(giveaway)); err != nil {
		return err
	}

	return nil
}

func (repo GiveawaysRepo) HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM giveaways g JOIN giveaway_winners w ON g.id = w.giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo GiveawaysRepo) GetCodeForInfoMessage(ctx context.Context, messageId, userId string) (string, error) {
	code, err := repo.mysql.WithContext(ctx).SelectStr("SELECT code FROM giveaways g JOIN giveaway_winners w ON g.id = w.giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (repo GiveawaysRepo) GetGiveawayByMessageId(ctx context.Context, messageId string) (*entities.Giveaway, error) {
	var giveaway SqlGiveaways
	if err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, type, start_time, end_time, guild_id, info_message_id, level FROM giveaways WHERE info_message_id = ?", messageId); err != nil {
		return nil, err
	}

	return FromSqlGiveaways(&giveaway), nil
}

func (repo GiveawaysRepo) InsertParticipantCandidate(ctx context.Context, guildId, guildName, candidateId, candidateName, approverId, approverName, channelId, messageId string, giveawayId int) error {
	candidate := &SqlThxParticipantCandidate{
		CandidateId:           candidateId,
		CandidateName:         candidateName,
		CandidateApproverId:   approverId,
		CandidateApproverName: approverName,
		GiveawayId:            giveawayId,
		GuildId:               guildId,
		GuildName:             guildName,
		MessageId:             messageId,
		ChannelId:             channelId,
	}
	if err := repo.mysql.WithContext(ctx).Insert(candidate); err != nil {
		return err
	}

	return nil
}

func (repo GiveawaysRepo) GetParticipantsWithThxAmount(ctx context.Context, guildId string, minThxAmount int) (result []entities.ThxParticipantWithThxAmount, err error) {
	var helpers []SqlThxParticipantWithThxAmount
	_, err = repo.mysql.WithContext(ctx).Select(&helpers, "SELECT user_id, amount FROM (SELECT user_id, COUNT(*) AS amount FROM giveaway_participants WHERE guild_id=? AND is_accepted=1 GROUP BY user_id) AS a WHERE amount > ?", guildId, minThxAmount)
	if err != nil {
		return nil, err
	}

	for _, helper := range helpers {
		result = append(result, *FromSqlThxParticipantWithThxAmount(&helper))
	}

	return result, nil
}

func (repo GiveawaysRepo) HasThxAmount(ctx context.Context, guildId, memberId string, minThxAmount int) (bool, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) AS amount  FROM giveaway_participants WHERE guild_id=? AND user_id=? AND is_accepted=1 HAVING amount > ?", guildId, memberId, minThxAmount)
	if err != nil {
		return false, err
	}

	return count > 1, nil
}

func (repo GiveawaysRepo) GetThxNotification(ctx context.Context, messageId string) (entities.ThxNotification, error) {
	var notification SqlThxNotification
	if err := repo.mysql.WithContext(ctx).SelectOne(&notification, "SELECT id, thx_message_id, notification_message_id FROM thx_notifications WHERE notification_message_id = ?", messageId); err != nil {
		return entities.ThxNotification{}, err
	}

	return *FromSqlThxNotification(&notification), nil
}

func (repo GiveawaysRepo) InsertThxNotification(ctx context.Context, thxMessageId, notificationMessageId string) error {
	notification := &SqlThxNotification{
		ThxMessageId:          thxMessageId,
		NotificationMessageId: notificationMessageId,
	}
	if err := repo.mysql.WithContext(ctx).Insert(notification); err != nil {
		return err
	}

	return nil
}

func (repo GiveawaysRepo) IsThxMessage(ctx context.Context, messageId string) (bool, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM giveaway_participants p JOIN giveaways g ON p.giveaway_id = g.id WHERE g.type = ? AND p.message_id = ?", entities.ThxGiveawayType, messageId)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo GiveawaysRepo) IsThxmeMessage(ctx context.Context, messageId string) (bool, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM thx_participant_candidates WHERE message_id = ?", messageId)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo GiveawaysRepo) GetParticipant(ctx context.Context, messageId string) (*entities.GiveawayParticipant, error) {
	var participant SqlGiveawaysParticipant
	if err := repo.mysql.WithContext(ctx).SelectOne(&participant, "SELECT id, giveaway_id, guild_id, user_id, user_name, join_time, user_level, message_id, channel_id, is_accepted, accept_time, accept_user, accept_user_id FROM giveaway_participants WHERE message_id = ?", messageId); err != nil {
		return nil, err
	}

	return FromSqlGiveawaysParticipant(&participant), nil
}

func (repo GiveawaysRepo) GetParticipantCandidate(ctx context.Context, messageId string) (entities.ThxParticipantCandidate, error) {
	var candidate SqlThxParticipantCandidate
	if err := repo.mysql.WithContext(ctx).SelectOne(&candidate, "SELECT id, candidate_id, candidate_name, candidate_approver_id, candidate_approver_name, giveaway_id, guild_id, guild_name, message_id, channel_id, is_accepted, accept_time FROM thx_participant_candidates WHERE message_id = ?", messageId); err != nil {
		return entities.ThxParticipantCandidate{}, err
	}

	return *FromSqlThxParticipantCandidate(&candidate), nil
}

func (repo GiveawaysRepo) UpdateParticipantCandidate(ctx context.Context, participantCandidate *entities.ThxParticipantCandidate, isAccepted bool) error {
	now := time.Now()
	participantCandidate.AcceptTime = &now
	participantCandidate.IsAccepted = sql.NullBool{Bool: isAccepted, Valid: true}
	if _, err := repo.mysql.WithContext(ctx).Update(ToSqlThxParticipantCandidate(participantCandidate)); err != nil {
		return err
	}

	return nil
}

func (repo GiveawaysRepo) IsGiveawayEnded(ctx context.Context, giveawayId int) (bool, error) {
	count, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM giveaways WHERE id = ? AND end_time IS NOT NULL", giveawayId)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo GiveawaysRepo) GetGiveawayById(ctx context.Context, giveawayId int) (*entities.Giveaway, error) {
	var giveaway SqlGiveaways
	if err := repo.mysql.WithContext(ctx).SelectOne(&giveaway, "SELECT id, type, start_time, end_time, guild_id, info_message_id, level FROM giveaways WHERE id = ?", giveawayId); err != nil {
		return nil, err
	}

	return FromSqlGiveaways(&giveaway), nil
}

func (repo GiveawaysRepo) GetLastCodesForUser(ctx context.Context, userId, giveawayType string, limit int) ([]string, error) {
	var codes []string
	_, err := repo.mysql.WithContext(ctx).Select(&codes, "SELECT code FROM giveaway_winners w JOIN giveaways g ON w.giveaway_id = g.id WHERE w.user_id = ? AND g.type = ? ORDER BY g.end_time DESC LIMIT ?", userId, giveawayType, limit)
	if err != nil {
		return nil, err
	}

	return codes, nil
}

func (repo GiveawaysRepo) RemoveAllThxParticipantEntries(ctx context.Context, giveawayId int, participantId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("UPDATE thx_participant_candidates SET is_accepted = FALSE WHERE candidate_id = ? AND giveawayId = ?", participantId, giveawayId)
	return err
}

func (repo GiveawaysRepo) UpdateParticipant(ctx context.Context, participant *entities.GiveawayParticipant, acceptUserId, acceptUsername string, isAccepted bool) error {
	now := time.Now()
	participant.AcceptTime = &now
	participant.IsAccepted = sql.NullBool{Bool: isAccepted, Valid: true}
	participant.AcceptUserId = sql.NullString{String: acceptUserId, Valid: true}
	participant.AcceptUser = sql.NullString{String: acceptUsername, Valid: true}
	_, err := repo.mysql.WithContext(ctx).Update(ToSqlGiveawaysParticipant(participant))
	if err != nil {
		return err
	}
	return nil
}

func (repo GiveawaysRepo) UpdateUserDailyMessageCount(ctx context.Context, userId string, guildId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("INSERT INTO daily_user_messages (user_id, day, guild_id, count) VALUES (?, DATE(now()), ?, 1) ON DUPLICATE KEY UPDATE COUNT = COUNT + 1", userId, guildId)
	return err
}

func (repo GiveawaysRepo) GetUsersWithMessagesFromLastDays(ctx context.Context, dayCount int, guildId string) ([]string, error) {
	var users []string
	_, err := repo.mysql.WithContext(ctx).Select(&users, "SELECT DISTINCT user_id FROM daily_user_messages WHERE guild_id = ? AND DAY > date_sub(now(), INTERVAL ? DAY)", guildId, dayCount)
	if err != nil {
		return nil, err
	}

	return users, err
}

func (repo GiveawaysRepo) GetCodesForInfoMessage(ctx context.Context, messageId, userId string) ([]string, error) {
	var codes []string
	_, err := repo.mysql.WithContext(ctx).Select(&codes, "SELECT code FROM giveaways g JOIN giveaway_winners w ON g.id = w.giveaway_id WHERE g.info_message_id = ? AND w.user_id = ?", messageId, userId)
	if err != nil {
		return nil, err
	}

	return codes, nil
}
