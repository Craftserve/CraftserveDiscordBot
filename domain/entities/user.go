package entities

import "context"

type Blacklist struct {
	Id            int    `json:"id"`
	GuildId       string `json:"guildId"`
	UserId        string `json:"userId"`
	BlacklisterId string `json:"blacklisterId"`
}

type MemberRole struct {
	Id       int    `json:"id"`
	GuildId  string `json:"guildId"`
	MemberId string `json:"memberId"`
	RoleId   string `json:"roleId"`
}

type HelperBlacklist struct {
	Id            int    `json:"id"`
	GuildId       string `json:"guildId"`
	UserId        string `json:"userId"`
	BlacklisterId string `json:"blacklisterId"`
}

type UserRepo interface {
	GetRolesForMember(ctx context.Context, guildId, memberId string) ([]MemberRole, error)
	AddRoleForMember(ctx context.Context, guildId, memberId, roleId string) error
	RemoveRoleForMember(ctx context.Context, guildId, memberId, roleId string) error
	IsUserHelperBlacklisted(ctx context.Context, userId, guildId string) (bool, error)
	IsUserBlacklisted(ctx context.Context, userId, guildId string) (bool, error)
	AddBlacklistForUser(ctx context.Context, userId, guildId, blacklisterId string) error
	RemoveBlacklistForUser(ctx context.Context, userId, guildId string) error
	AddHelperBlacklistForUser(ctx context.Context, userId, guildId, blacklisterId string) error
	RemoveHelperBlacklistForUser(ctx context.Context, userId, guildId string) error
}
