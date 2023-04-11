package repos

import (
	"context"
	"database/sql"
	"github.com/go-gorp/gorp"
	"log"
)

type UserRepo struct {
	mysql *gorp.DbMap
}

func NewUserRepo(mysql *gorp.DbMap) *UserRepo {
	mysql.AddTableWithName(Blacklist{}, "blacklists").SetKeys(true, "id").SetUniqueTogether("guild_id", "user_id")
	mysql.AddTableWithName(MemberRole{}, "member_roles").SetKeys(true, "id")
	mysql.AddTableWithName(HelperBlacklist{}, "helper_blacklists").SetKeys(true, "id").SetUniqueTogether("guild_id", "user_id")

	return &UserRepo{mysql: mysql}
}

type Blacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

type MemberRole struct {
	Id       int    `db:"id,primarykey,autoincrement"`
	GuildId  string `db:"guild_id,size:255"`
	MemberId string `db:"member_id,size:255"`
	RoleId   string `db:"role_id,size:255"`
}

type HelperBlacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

func (repo *UserRepo) GetRolesForMember(ctx context.Context, guildId, memberId string) ([]MemberRole, error) {
	var memberRoles []MemberRole
	_, err := repo.mysql.WithContext(ctx).Select(&memberRoles, "SELECT * FROM member_roles WHERE guild_id = ? AND member_id = ?", guildId, memberId)
	if err != nil {
		return nil, err
	}

	return memberRoles, nil
}

func (repo *UserRepo) AddRoleForMember(ctx context.Context, guildId, memberId, roleId string) error {
	role := MemberRole{GuildId: guildId, RoleId: roleId, MemberId: memberId}
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

func (repo *UserRepo) IsUserHelperBlacklisted(ctx context.Context, userId string, guildId string) (bool, error) {
	var helperBlacklist HelperBlacklist // todo: use another query to check if user is blacklisted
	err := repo.mysql.WithContext(ctx).SelectOne(&helperBlacklist, "SELECT * FROM helper_blacklists WHERE user_id = ? AND guild_id = ?", userId, guildId)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		log.Panicln("("+userId+") checkUserBlacklist#DbMap.SelectOne", err)
	}
	return true, nil
}

func (repo *UserRepo) UpdateMemberSavedRoles(ctx context.Context, memberRoles []string, memberId, guildId string) { //todo: how to properly return errors from this?
	savedRoles, err := repo.GetRolesForMember(ctx, guildId, memberId)
	if err != nil {
		log.Println("("+guildId+") "+"updateMemberSavedRoles Error while getting saved roles", err)
		return
	}
	var savedRolesIds []string
	for _, role := range savedRoles {
		savedRolesIds = append(savedRolesIds, role.RoleId)
	}

	for _, memberRole := range memberRoles {
		found := false
		for i, savedRole := range savedRolesIds {
			if savedRole == memberRole {
				found = true
				savedRolesIds[i] = ""
				break
			}
		}
		if !found {
			err = repo.AddRoleForMember(ctx, guildId, memberId, memberRole)
			if err != nil {
				log.Println("("+guildId+") Error while saving new role info", err)
				continue
			}
		}
	}

	for _, savedRole := range savedRolesIds {
		if savedRole != "" {
			err = repo.RemoveRoleForMember(ctx, guildId, memberId, savedRole)
			if err != nil {
				log.Println("("+guildId+") "+"Error while deleting info about member role", err)
				continue
			}
		}
	}
}

func (repo *UserRepo) IsUserBlacklisted(ctx context.Context, userId string, guildId string) (bool, error) {
	ret, err := repo.mysql.WithContext(ctx).SelectInt("SELECT count(*) FROM blacklists WHERE guild_id = ? AND user_id = ?", guildId, userId)
	if err != nil {
		return false, err
	}
	return ret > 0, nil
}

func (repo *UserRepo) AddBlacklistForUser(ctx context.Context, userId, guildId, blacklisterId string) error {
	blacklist := Blacklist{UserId: userId, GuildId: guildId, BlacklisterId: blacklisterId}
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
	blacklist := HelperBlacklist{UserId: userId, GuildId: guildId, BlacklisterId: blacklisterId}
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
