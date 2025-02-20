package entities

import (
	"github.com/Craftserve/monies"
	"time"
)

type DurationString string

type VoucherAction struct {
	// Contains targets that this voucher can be used on, also targets[0] is default when creating server from voucher
	//Targets  []VoucherTarget `json:"targets"`
	WalletTx map[monies.CurrencyCode]monies.Money `json:"wallet_tx,omitempty"`
}

type Voucher struct {
	Id        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	Expires   *time.Time      `json:"expires"`
	Data      []VoucherAction `json:"data"`
}

type GenerateVoucherPayload struct {
	Length  int             `json:"length"`
	Charset string          `json:"charset"`
	Prefix  *string         `json:"prefix"`
	Expires *time.Time      `json:"expires"`
	GroupId *string         `json:"group_id"`
	Uses    *int            `json:"uses"`
	PerUser *int            `json:"per_user"`
	Actions []VoucherAction `json:"actions"`
}
