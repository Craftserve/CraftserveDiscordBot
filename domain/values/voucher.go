package values

import "time"

const (
	VoucherCharset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	VoucherLength     = 8
	VoucherValuePLN   = 500
	VoucherExpiration = 364 * 24 * time.Hour
)
