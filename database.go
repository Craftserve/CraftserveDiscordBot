package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v2"
)

type Giveaway struct {
	Id            int            `db:"id, primarykey, autoincrement"`
	StartTime     time.Time      `db:"start_time"`
	EndTime       mysql.NullTime `db:"end_time"`
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
	AcceptTime   mysql.NullTime `db:"accept_time"`
	AcceptUser   sql.NullString `db:"accept_user,size:255"`
	AcceptUserId sql.NullString `db:"accept_user_id,size:255"`
}

type ParticipantCandidate struct {
	Id                    int            `db:"id, primarykey, autoincrement"`
	CandidateId           string         `db:"candidate_id,size:255"`
	CandidateName         string         `db:"candidate_name,size:255"`
	CandidateApproverId   string         `db:"candidate_approver_id,size:255"`
	CandidateApproverName string         `db:"candidate_approver_name,size:255"`
	GuildId               string         `db:"guild_id,size:255"`
	GuildName             string         `db:"guild_name,size:255"`
	MessageId             string         `db:"message_id,size:255"`
	ChannelId             string         `db:"channel_id,size:255"`
	IsAccepted            sql.NullBool   `db:"is_accepted"`
	AcceptTime            mysql.NullTime `db:"accept_time"`
}

type Blacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

type ServerConfig struct {
	Id                    int    `db:"id,primarykey,autoincrement"`
	GuildId               string `db:"guild_id,size:255"`
	AdminRole             string `db:"admin_role,size:255"`
	MainChannel           string `db:"main_channel,size:255"`
	ThxInfoChannel        string `db:"thx_info_channel,size:255"`
	HelperRoleName        string `db:"helper_role_name,size:255"`
	HelperRoleThxesNeeded int    `db:"helper_role_thxes_needed"`
}

type MemberRole struct {
	Id       int    `db:"id,primarykey,autoincrement"`
	GuildId  string `db:"guild_id,size:255"`
	MemberId string `db:"member_id,size:255"`
	RoleId   string `db:"role_id,size:255"`
}

type ThxNotification struct {
	Id                       int    `db:"id,primarykey,autoincrement"`
	MessageId                string `db:"message_id,size:255"`
	ThxNotificationMessageId string `db:"thx_notification_message_id,size:255"`
}

type HelperBlacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

var DbMap gorp.DbMap

func initDatabase() {
	db, err := sql.Open("mysql", config.MysqlString)
	if err != nil {
		log.Panic("initDatabase sql.Open(\"mysql\", "+config.MysqlString+")", err)
	}

	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	DbMap.AddTableWithName(Giveaway{}, "Giveaways").SetKeys(true, "id")
	DbMap.AddTableWithName(Participant{}, "Participants").SetKeys(true, "id")
	DbMap.AddTableWithName(ParticipantCandidate{}, "ParticipantCandidates").SetKeys(true, "id")
	DbMap.AddTableWithName(Blacklist{}, "Blacklists").SetKeys(true, "id")
	DbMap.AddTableWithName(ServerConfig{}, "ServerConfig").SetKeys(true, "id")
	DbMap.AddTableWithName(MemberRole{}, "MemberRoles").SetKeys(true, "id")
	DbMap.AddTableWithName(ThxNotification{}, "ThxNotifications").SetKeys(true, "id")
	DbMap.AddTableWithName(HelperBlacklist{}, "HelperBlacklist").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Panic("initDatabase#DbMap.CreateTablesIfNotExists", err)
	}
}
