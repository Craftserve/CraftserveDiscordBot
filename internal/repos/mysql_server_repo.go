package repos

import (
	"context"
	"github.com/go-gorp/gorp"
)

type ServerRepo struct {
	mysql *gorp.DbMap
}

func NewServerRepo(mysql *gorp.DbMap) *ServerRepo {
	mysql.AddTableWithName(ServerConfig{}, "server_configs").SetKeys(true, "id")

	return &ServerRepo{mysql: mysql}
}

type ServerConfig struct {
	Id                           int    `db:"id,primarykey,autoincrement"`
	GuildId                      string `db:"guild_id,size:255"`
	AdminRoleId                  string `db:"admin_role_id,size:255"`
	MainChannel                  string `db:"main_channel,size:255"`
	ThxInfoChannel               string `db:"thx_info_channel,size:255"`
	HelperRoleId                 string `db:"helper_role_id,size:255"`
	HelperRoleThxesNeeded        int    `db:"helper_role_thxes_needed"`
	MessageGiveawayWinners       int    `db:"message_giveaway_winners,default:0"`
	UnconditionalGiveawayChannel string `db:"unconditional_giveaway_channel,size:255"`
	UnconditionalGiveawayWinners int    `db:"unconditional_giveaway_winners,default:0"`
}

func (repo *ServerRepo) GetServerConfigForGuild(ctx context.Context, guildId string) (ServerConfig, error) {
	var serverConfig ServerConfig
	err := repo.mysql.WithContext(ctx).SelectOne(&serverConfig, "SELECT id, guild_id, admin_role_id, main_channel, thx_info_channel, helper_role_id, helper_role_thxes_needed, message_giveaway_winners, unconditional_giveaway_channel, unconditional_giveaway_winners FROM server_configs WHERE guild_id = ?", guildId)
	if err != nil {
		return ServerConfig{}, err
	}
	return serverConfig, nil
}

func (repo *ServerRepo) InsertServerConfig(ctx context.Context, guildId, giveawayChannel, adminRole string) error {
	var serverConfig ServerConfig
	serverConfig.GuildId = guildId
	serverConfig.MainChannel = giveawayChannel
	serverConfig.AdminRoleId = adminRole
	serverConfig.HelperRoleThxesNeeded = 0
	err := repo.mysql.WithContext(ctx).Insert(&serverConfig)
	if err != nil {
		return err
	}
	return nil
}

func (repo *ServerRepo) UpdateServerConfig(ctx context.Context, serverConfig *ServerConfig) error {
	_, err := repo.mysql.WithContext(ctx).Update(serverConfig)
	if err != nil {
		return err
	}
	return nil
}

func (repo *ServerRepo) GetAdminRoleForGuild(ctx context.Context, guildId string) (string, error) {
	serverConfig, err := repo.GetServerConfigForGuild(ctx, guildId)
	if err != nil {
		return "", err
	}
	return serverConfig.AdminRoleId, nil
}

func (repo *ServerRepo) GetMainChannelForGuild(ctx context.Context, guildId string) (string, error) {
	str, err := repo.mysql.WithContext(ctx).SelectStr("SELECT main_channel FROM server_configs WHERE guild_id = ?", guildId)
	if err != nil {
		return "", err
	}
	return str, nil
}

func (repo *ServerRepo) GetGuildsWithMessageGiveawaysEnabled(ctx context.Context) ([]string, error) {
	var guilds []string
	_, err := repo.mysql.WithContext(ctx).Select(&guilds, "SELECT guild_id FROM server_configs WHERE message_giveaway_winners > 0")
	if err != nil {
		return nil, err
	}
	return guilds, nil
}
