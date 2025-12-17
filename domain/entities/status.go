package entities

import (
	"context"
	"encoding/json"
)

type Status struct {
	Id        int             `json:"id"`
	GuildId   string          `json:"guildId"`
	Type      string          `json:"type"`
	ShortName string          `json:"shortName"`
	Content   json.RawMessage `json:"content"`
}

type StatusRepo interface {
	GetAllStatuses(ctx context.Context, guildId string) ([]Status, error)
	GetStatusById(ctx context.Context, id int64) (*Status, error)
	UpdateStatus(ctx context.Context, status *Status) error
	CreateStatus(ctx context.Context, status *Status) error
	RemoveStatus(ctx context.Context, id int) error
}
