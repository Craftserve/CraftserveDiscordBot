package discord

import (
	"encoding/json"
	"strings"
)

type DiscordGoError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func EqualError(err error, code int) bool {
	start := strings.Index(err.Error(), "{")
	if start == -1 {
		return false
	}

	jsonError := err.Error()[start:]

	var discordGoError DiscordGoError
	err = json.Unmarshal([]byte(jsonError), &discordGoError)
	if err != nil {
		return false
	}

	// Compare codes
	return discordGoError.Code == code
}
