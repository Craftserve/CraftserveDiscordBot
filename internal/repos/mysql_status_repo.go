package repos

import (
	"context"
	"csrvbot/domain/entities"
	"encoding/json"

	"github.com/go-gorp/gorp"
)

type StatusRepo struct {
	mysql *gorp.DbMap
}

func NewStatusRepo(mysql *gorp.DbMap) *StatusRepo {
	mysql.AddTableWithName(SqlStatus{}, "status").SetKeys(true, "id")

	return &StatusRepo{mysql: mysql}
}

type SqlStatus struct {
	Id        int             `db:"id,primarykey,autoincrement"`
	GuildId   string          `db:"guild_id,size:255"`
	Type      string          `db:"type,size:50"`
	ShortName string          `db:"short_name,size:100"`
	Content   json.RawMessage `db:"content"`
}

func parseJsonToMap(content json.RawMessage) map[string]string {
	result := make(map[string]string)

	err := json.Unmarshal(content, &result)

	if err != nil {
		return make(map[string]string)
	}

	return result
}

func parseMapToJson(content map[string]string) json.RawMessage {
	result, err := json.Marshal(content)

	if err != nil {
		return json.RawMessage("{}")
	}

	return result
}

func FromSqlStatus(status *SqlStatus) *entities.Status {
	return &entities.Status{
		Id:        status.Id,
		GuildId:   status.GuildId,
		Type:      status.Type,
		ShortName: status.ShortName,
		Content:   status.Content,
	}
}

func ToSqlStatus(status *entities.Status) *SqlStatus {
	return &SqlStatus{
		Id:        status.Id,
		GuildId:   status.GuildId,
		Type:      status.Type,
		ShortName: status.ShortName,
		Content:   status.Content,
	}
}

func (repo *StatusRepo) GetAllStatuses(ctx context.Context, guildId string) ([]entities.Status, error) {
	var sqlStatuses []SqlStatus

	_, err := repo.mysql.WithContext(ctx).Select(&sqlStatuses, "SELECT id, guild_id, short_name, type, content FROM status WHERE guild_id = ?", guildId)
	if err != nil {
		return nil, err
	}

	var statuses []entities.Status
	for _, sqlStatus := range sqlStatuses {
		statuses = append(statuses, *FromSqlStatus(&sqlStatus))
	}

	return statuses, nil
}

func (repo *StatusRepo) GetStatusById(ctx context.Context, id int64) (*entities.Status, error) {
	var sqlStatus SqlStatus
	err := repo.mysql.WithContext(ctx).SelectOne(&sqlStatus, "SELECT id, guild_id, short_name, type, content FROM status WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	return FromSqlStatus(&sqlStatus), nil
}

func (repo *StatusRepo) UpdateStatus(ctx context.Context, status *entities.Status) error {
	_, err := repo.mysql.WithContext(ctx).Update(ToSqlStatus(status))
	if err != nil {
		return err
	}
	return nil
}

func (repo *StatusRepo) CreateStatus(ctx context.Context, status *entities.Status) error {
	err := repo.mysql.WithContext(ctx).Insert(ToSqlStatus(status))
	if err != nil {
		return err
	}
	return nil
}

func (repo *StatusRepo) RemoveStatus(ctx context.Context, id int) error {
	_, err := repo.mysql.WithContext(ctx).Delete(&SqlStatus{Id: id})
	if err != nil {
		return err
	}
	return nil
}
