package dtos

import (
	"csrvbot/domain/entities"
	"time"
)

type GenerateVoucherPayload struct {
	Length   int                      `json:"length"`
	Charset  string                   `json:"charset"`
	Prefix   *string                  `json:"prefix"`
	Expires  *time.Time               `json:"expires"`
	GroupId  *string                  `json:"group_id"`
	Uses     *int                     `json:"uses"`
	PerUser  *int                     `json:"per_user"`
	Actions  []entities.VoucherAction `json:"actions"`
	Quantity *int                     `json:"quantity"`
}
