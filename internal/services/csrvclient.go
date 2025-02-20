package services

import (
	"bytes"
	"context"
	"csrvbot/domain/entities"
	"csrvbot/pkg/logger"
	"encoding/json"
	"fmt"
	"github.com/Craftserve/monies"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type CsrvClient struct {
	Secret        string
	Environment   string
	CraftserveUrl string
}

func NewCsrvClient(secret, environment, craftserveUrl string) *CsrvClient {
	return &CsrvClient{Secret: secret, Environment: environment, CraftserveUrl: craftserveUrl}
}

func (c *CsrvClient) GetCSRVCode(ctx context.Context) (string, error) {
	log := logger.GetLoggerFromContext(ctx)
	log.Debug("Generating CSRV voucher")

	if c.Environment == "development" {
		return fmt.Sprintf("DEV-%d", rand.Int()), nil
	}

	prefix, group := "discord", "discord-giveaway"
	expires := time.Now().Add(365 * 24 * time.Hour)
	uses := 1
	payload := entities.GenerateVoucherPayload{
		Length:  16,
		Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		Prefix:  &prefix,
		Expires: &expires,
		GroupId: &group,
		Uses:    &uses,
		PerUser: &uses,
		Actions: []entities.VoucherAction{
			{
				WalletTx: map[monies.CurrencyCode]monies.Money{
					monies.PLN: monies.MustNew(500, monies.PLN),
				},
			},
		},
	}

	bodyPayload := new(bytes.Buffer)
	err := json.NewEncoder(bodyPayload).Encode(payload)
	if err != nil {
		return "", fmt.Errorf("GetCSRVCode json.NewEncoder failed: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/admin/voucher/generate", c.CraftserveUrl), bodyPayload)
	if err != nil {
		return "", err
	}

	req.AddCookie(&http.Cookie{Name: "user_access_token", Value: c.Secret})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var voucher entities.VoucherResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bodyBytes, &voucher)
	if err != nil {
		return "", fmt.Errorf("getCSRVCode json.Unmarshal failed: %w with body: %s", err, string(bodyBytes))
	}
	err = resp.Body.Close()
	if err != nil {
		log.WithError(err).Error("getCSRVCode resp.Body.Close()")
	}
	return voucher.Id, nil
}
