package entities

import (
	"context"
	"database/sql"
	"time"
)

const (
	JoinedGiveawayType  = "joined"
	LevelGiveawayType   = "level"
	MessageGiveawayType = "message"
	ThxGiveawayType     = "thx"
)

type Giveaway struct {
	Id            int        `json:"id"`
	Type          string     `json:"type"`
	StartTime     time.Time  `json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
	GuildId       string     `json:"guildId"`
	InfoMessageId *string    `json:"infoMessageId"`
	Level         *int       `json:"level"`
}

type GiveawayParticipant struct {
	Id           int            `json:"id"`
	GiveawayId   int            `json:"giveawayId"`
	GuildId      string         `json:"guildId"`
	UserId       string         `json:"userId"`
	UserName     string         `json:"userName"`
	JoinTime     time.Time      `json:"joinTime"`
	UserLevel    *int           `json:"userLevel"`
	MessageId    *string        `json:"messageId"`
	ChannelId    *string        `json:"channelId"`
	IsAccepted   sql.NullBool   `json:"isAccepted"`
	AcceptTime   *time.Time     `json:"acceptTime"`
	AcceptUser   sql.NullString `json:"acceptUser"`
	AcceptUserId sql.NullString `json:"acceptUserId"`
}

type GiveawayWinner struct {
	Id         int    `json:"id"`
	GiveawayId int    `json:"giveawayId"`
	UserId     string `json:"userId"`
	Code       string `json:"code"`
}

type ThxParticipantCandidate struct {
	Id                    int          `json:"id"`
	CandidateId           string       `json:"candidateId"`           // User ID
	CandidateName         string       `json:"candidateName"`         // User name
	CandidateApproverId   string       `json:"candidateApproverId"`   // Approver User ID
	CandidateApproverName string       `json:"candidateApproverName"` // Approver User name
	GiveawayId            int          `json:"giveawayId"`
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

type ThxParticipantWithThxAmount struct {
	UserId    string `json:"userId"`
	ThxAmount int    `json:"thxAmount"`
}

type DailyUserMessages struct {
	Id      int    `json:"id"`
	UserId  string `json:"userId"`
	Day     string `json:"day"`
	GuildId string `json:"guildId"`
	Count   int    `json:"count"`
}

type GiveawaysRepo interface {
	// Common
	GetGiveawayForGuild(ctx context.Context, guildId, giveawayType string) (*Giveaway, error)
	GetUnfinishedGiveaways(ctx context.Context, giveawayType string) ([]Giveaway, error)
	GetParticipantsForGiveaway(ctx context.Context, giveawayId int, accepted *bool) ([]GiveawayParticipant, error)
	CountParticipantsForGiveaway(ctx context.Context, giveawayId int) (int, error)
	InsertGiveaway(ctx context.Context, guildId string, messageId *string, giveawayType string, level *int) error
	InsertParticipant(ctx context.Context, giveawayId, level int, guildId, userId, userName string, messageId, channelId *string) error
	InsertWinner(ctx context.Context, giveawayId int, userId, code string) error
	FinishGiveaway(ctx context.Context, giveaway *Giveaway, messageId *string) error
	HasWonGiveawayByMessageId(ctx context.Context, messageId, userId string) (bool, error)
	GetCodeForInfoMessage(ctx context.Context, messageId, userId string) (string, error)
	GetGiveawayByMessageId(ctx context.Context, messageId string) (*Giveaway, error)

	// Thx
	InsertParticipantCandidate(ctx context.Context, guildId, guildName, candidateId, candidateName, approverId, approverName, channelId, messageId string, giveawayId int) error
	GetParticipantsWithThxAmount(ctx context.Context, guildId string, minThxAmount int) ([]ThxParticipantWithThxAmount, error)
	HasThxAmount(ctx context.Context, guildId, memberId string, minThxAmount int) (bool, error)
	GetThxNotification(ctx context.Context, messageId string) (ThxNotification, error)
	InsertThxNotification(ctx context.Context, thxMessageId, notificationMessageId string) error
	IsThxMessage(ctx context.Context, messageId string) (bool, error)
	IsThxmeMessage(ctx context.Context, messageId string) (bool, error)
	GetParticipant(ctx context.Context, messageId string) (*GiveawayParticipant, error)
	GetParticipantCandidate(ctx context.Context, messageId string) (ThxParticipantCandidate, error)
	UpdateParticipantCandidate(ctx context.Context, participantCandidate *ThxParticipantCandidate, isAccepted bool) error
	IsGiveawayEnded(ctx context.Context, giveawayId int) (bool, error)
	GetGiveawayById(ctx context.Context, giveawayId int) (*Giveaway, error)
	GetLastCodesForUser(ctx context.Context, userId, giveawayType string, limit int) ([]string, error)
	RemoveAllThxParticipantEntries(ctx context.Context, giveawayId int, participantId string) error
	UpdateParticipant(ctx context.Context, participant *GiveawayParticipant, acceptUserId, acceptUsername string, isAccepted bool) error

	// Message
	UpdateUserDailyMessageCount(ctx context.Context, userId string, guildId string) error
	GetUsersWithMessagesFromLastDays(ctx context.Context, dayCount int, guildId string) ([]string, error)
	GetCodesForInfoMessage(ctx context.Context, messageId, userId string) ([]string, error)
}
