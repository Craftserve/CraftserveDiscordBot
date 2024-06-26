package repos

import (
	"context"
	"csrvbot/domain/entities"
	"github.com/go-gorp/gorp"
)

type UserRepo struct {
	mysql *gorp.DbMap
}

func NewUserRepo(mysql *gorp.DbMap) *UserRepo {
	mysql.AddTableWithName(SqlBlacklist{}, "blacklists").SetKeys(true, "id").SetUniqueTogether("guild_id", "user_id")
	mysql.AddTableWithName(SqlMemberRole{}, "member_roles").SetKeys(true, "id")
	mysql.AddTableWithName(SqlHelperBlacklist{}, "helper_blacklists").SetKeys(true, "id").SetUniqueTogether("guild_id", "user_id")

	return &UserRepo{mysql: mysql}
}

type SqlBlacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

type SqlMemberRole struct {
	Id       int    `db:"id,primarykey,autoincrement"`
	GuildId  string `db:"guild_id,size:255"`
	MemberId string `db:"member_id,size:255"`
	RoleId   string `db:"role_id,size:255"`
}

type SqlHelperBlacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

func FromSqlBlacklist(blacklist *SqlBlacklist) *entities.Blacklist {
	return &entities.Blacklist{
		Id:            blacklist.Id,
		GuildId:       blacklist.GuildId,
		UserId:        blacklist.UserId,
		BlacklisterId: blacklist.BlacklisterId,
	}
}

func ToSqlBlacklist(blacklist *entities.Blacklist) *SqlBlacklist {
	return &SqlBlacklist{
		Id:            blacklist.Id,
		GuildId:       blacklist.GuildId,
		UserId:        blacklist.UserId,
		BlacklisterId: blacklist.BlacklisterId,
	}
}

func FromSqlMemberRole(memberRole *SqlMemberRole) *entities.MemberRole {
	return &entities.MemberRole{
		Id:       memberRole.Id,
		GuildId:  memberRole.GuildId,
		MemberId: memberRole.MemberId,
		RoleId:   memberRole.RoleId,
	}
}

func ToSqlMemberRole(memberRole *entities.MemberRole) *SqlMemberRole {
	return &SqlMemberRole{
		Id:       memberRole.Id,
		GuildId:  memberRole.GuildId,
		MemberId: memberRole.MemberId,
		RoleId:   memberRole.RoleId,
	}
}

func FromSqlHelperBlacklist(helperBlacklist *SqlHelperBlacklist) *entities.HelperBlacklist {
	return &entities.HelperBlacklist{
		Id:            helperBlacklist.Id,
		GuildId:       helperBlacklist.GuildId,
		UserId:        helperBlacklist.UserId,
		BlacklisterId: helperBlacklist.BlacklisterId,
	}
}

func ToSqlHelperBlacklist(helperBlacklist *entities.HelperBlacklist) *SqlHelperBlacklist {
	return &SqlHelperBlacklist{
		Id:            helperBlacklist.Id,
		GuildId:       helperBlacklist.GuildId,
		UserId:        helperBlacklist.UserId,
		BlacklisterId: helperBlacklist.BlacklisterId,
	}
}

func (repo *UserRepo) GetRolesForMember(ctx context.Context, guildId, memberId string) ([]entities.MemberRole, error) {
	var memberRoles []SqlMemberRole
	_, err := repo.mysql.WithContext(ctx).Select(&memberRoles, "SELECT id, guild_id, member_id, role_id FROM member_roles WHERE guild_id = ? AND member_id = ?", guildId, memberId)
	if err != nil {
		return nil, err
	}

	var result []entities.MemberRole
	for _, memberRole := range memberRoles {
		result = append(result, *FromSqlMemberRole(&memberRole))
	}

	return result, nil
}

func (repo *UserRepo) AddRoleForMember(ctx context.Context, guildId, memberId, roleId string) error {
	role := SqlMemberRole{GuildId: guildId, RoleId: roleId, MemberId: memberId}
	err := repo.mysql.WithContext(ctx).Insert(&role)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepo) RemoveRoleForMember(ctx context.Context, guildId, memberId, roleId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("DELETE FROM member_roles WHERE guild_id = ? AND member_id = ? AND role_id = ?", guildId, memberId, roleId)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepo) IsUserHelperBlacklisted(ctx context.Context, userId, guildId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM helper_blacklists WHERE guild_id = ? AND user_id = ?", guildId, userId)
	if err != nil {
		return false, err
	}
	return ret > 0, nil
}

func (repo *UserRepo) IsUserBlacklisted(ctx context.Context, userId string, guildId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT COUNT(*) FROM blacklists WHERE guild_id = ? AND user_id = ?", guildId, userId)
	if err != nil {
		return false, err
	}
	return ret > 0, nil
}

func (repo *UserRepo) AddBlacklistForUser(ctx context.Context, userId, guildId, blacklisterId string) error {
	blacklist := SqlBlacklist{UserId: userId, GuildId: guildId, BlacklisterId: blacklisterId}
	err := repo.mysql.WithContext(ctx).Insert(&blacklist)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepo) RemoveBlacklistForUser(ctx context.Context, userId, guildId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("DELETE FROM blacklists WHERE guild_id = ? AND user_id = ?", guildId, userId)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepo) AddHelperBlacklistForUser(ctx context.Context, userId, guildId, blacklisterId string) error {
	blacklist := SqlHelperBlacklist{UserId: userId, GuildId: guildId, BlacklisterId: blacklisterId}
	err := repo.mysql.WithContext(ctx).Insert(&blacklist)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepo) RemoveHelperBlacklistForUser(ctx context.Context, userId, guildId string) error {
	_, err := repo.mysql.WithContext(ctx).Exec("DELETE FROM helper_blacklists WHERE guild_id = ? AND user_id = ?", guildId, userId)
	if err != nil {
		return err
	}
	return nil
}
