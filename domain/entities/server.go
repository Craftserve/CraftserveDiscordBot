package entities

import (
	"context"
	"encoding/json"
)

type ServerConfig struct {
	Id                           int             `json:"id"`
	GuildId                      string          `json:"guildId"`
	AdminRoleId                  string          `json:"adminRoleId"`
	MainChannel                  string          `json:"mainChannel"`
	ThxInfoChannel               string          `json:"thxInfoChannel"`
	HelperRoleId                 string          `json:"helperRoleId"`
	HelperRoleThxesNeeded        int             `json:"helperRoleThxesNeeded"`
	MessageGiveawayWinners       int             `json:"messageGiveawayWinners"`
	UnconditionalGiveawayChannel string          `json:"unconditionalGiveawayChannel"`
	UnconditionalGiveawayWinners int             `json:"unconditionalGiveawayWinners"`
	ConditionalGiveawayChannel   string          `json:"conditionalGiveawayChannel"`
	ConditionalGiveawayWinners   int             `json:"conditionalGiveawayWinners"`
	ConditionalGiveawayLevels    json.RawMessage `json:"conditionalGiveawayLevels"`
}

type ServerRepo interface {
	GetServerConfigForGuild(ctx context.Context, guildId string) (ServerConfig, error)
	InsertServerConfig(ctx context.Context, guildId, giveawayChannel, adminRole string) error
	UpdateServerConfig(ctx context.Context, serverConfig *ServerConfig) error
	GetAdminRoleForGuild(ctx context.Context, guildId string) (string, error)
	GetMainChannelForGuild(ctx context.Context, guildId string) (string, error)
	GetGuildsWithMessageGiveawaysEnabled(ctx context.Context) ([]string, error)
	GetConditionalGiveawayLevels(ctx context.Context, guildId string) ([]int, error)
}
