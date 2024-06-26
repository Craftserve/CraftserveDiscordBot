package entities

import (
	"context"
	"database/sql"
	"time"
)

type Giveaway struct {
	Id            int            `json:"id"`
	StartTime     time.Time      `json:"startTime"`
	EndTime       *time.Time     `json:"endTime"`
	GuildId       string         `json:"guildId"`
	GuildName     string         `json:"guildName"`
	WinnerId      sql.NullString `json:"winnerId"`
	WinnerName    sql.NullString `json:"winnerName"`
	InfoMessageId sql.NullString `json:"infoMessageId"`
	Code          sql.NullString `json:"code"`
}

type Participant struct {
	Id           int            `json:"id"`
	UserName     string         `json:"userName"`
	UserId       string         `json:"userId"`
	GiveawayId   int            `json:"giveawayId"`
	CreateTime   time.Time      `json:"createTime"`
	GuildName    string         `json:"guildName"`
	GuildId      string         `json:"guildId"`
	MessageId    string         `json:"messageId"`
	ChannelId    string         `json:"channelId"`
	IsAccepted   sql.NullBool   `json:"isAccepted"`
	AcceptTime   *time.Time     `json:"acceptTime"`
	AcceptUser   sql.NullString `json:"acceptUser"`
	AcceptUserId sql.NullString `json:"acceptUserId"`
}

type ParticipantCandidate struct {
	Id                    int          `json:"id"`
	CandidateId           string       `json:"candidateId"`
	CandidateName         string       `json:"candidateName"`
	CandidateApproverId   string       `json:"candidateApproverId"`
	CandidateApproverName string       `json:"candidateApproverName"`
	GuildId               string       `json:"guildId"`
	GuildName             string       `json:"guildName"`
	MessageId             string       `json:"messageId"`
	ChannelId             string       `json:"channelId"`
	IsAccepted            sql.NullBool `json:"isAccepted"`
	AcceptTime            *time.Time   `json:"acceptTime"`
}

type ThxNotification struct {
	Id                    int    `json:"id"`
	ThxMessageId          string `json:"thxMessageId"`
	NotificationMessageId string `json:"notificationMessageId"`
}

type ParticipantWithThxAmount struct {
	UserId    string `json:"user-id"`
	ThxAmount int    `json:"thx-amount"`
}

type GiveawayRepo interface {
	GetGiveawayForGuild(ctx context.Context, guildId string) (Giveaway, error)
	GetParticipantNamesForGiveaway(ctx context.Context, giveawayId int) ([]string, error)
	InsertGiveaway(ctx context.Context, guildId string, guildName string) error
	InsertParticipant(ctx context.Context, giveawayId int, guildId, guildName, userId, userName, channelId, messageId string) error
	InsertParticipantCandidate(ctx context.Context, guildId, guildName, candidateId, candidateName, approverId, approverName, channelId, messageId string) error
	GetParticipantsWithThxAmount(ctx context.Context, guildId string, minThxAmount int) ([]ParticipantWithThxAmount, error)
	HasThxAmount(ctx context.Context, guildId, memberId string, minThxAmount int) (bool, error)
	GetParticipantsForGiveaway(ctx context.Context, giveawayId int) ([]Participant, error)
	GetThxNotification(ctx context.Context, messageId string) (ThxNotification, error)
	InsertThxNotification(ctx context.Context, thxMessageId, notificationMessageId string) error
	IsThxMessage(ctx context.Context, messageId string) (bool, error)
	IsThxmeMessage(ctx context.Context, messageId string) (bool, error)
	GetParticipant(ctx context.Context, messageId string) (Participant, error)
	UpdateParticipant(ctx context.Context, participant *Participant, acceptUserId, acceptUsername string, isAccepted bool) error
	GetParticipantCandidate(ctx context.Context, messageId string) (ParticipantCandidate, error)
	UpdateParticipantCandidate(ctx context.Context, participantCandidate *ParticipantCandidate, isAccepted bool) error
	FinishGiveaway(ctx context.Context, giveaway *Giveaway, messageId, code, winnerId, winnerName string) error
	GetUnfinishedGiveaways(ctx context.Context) ([]Giveaway, error)
	RemoveAllParticipantEntries(ctx context.Context, giveawayId int, participantId string) error
	HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error)
	IsGiveawayEnded(ctx context.Context, giveawayId int) (bool, error)
	GetGiveawayById(ctx context.Context, giveawayId int) (*Giveaway, error)
	GetCodeForInfoMessage(ctx context.Context, infoMessageId string) (string, error)
	GetLastCodesForUser(ctx context.Context, userId string, limit int) ([]string, error)
}
