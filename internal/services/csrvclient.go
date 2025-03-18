package services

import (
	"bytes"
	"context"
	"csrvbot/domain/entities"
	"csrvbot/domain/values"
	"csrvbot/dtos"
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
	Secret         string
	Environment    string
	CraftserveUrl  string
	ValuePLN       int
	ExpirationDays int
}

func NewCsrvClient(secret, environment, craftserveUrl string, valuePLN, expirationDays int) *CsrvClient {
	return &CsrvClient{Secret: secret, Environment: environment, CraftserveUrl: craftserveUrl, ValuePLN: valuePLN, ExpirationDays: expirationDays}
}

func (c *CsrvClient) GetCSRVCode(ctx context.Context) (string, error) {
	log := logger.GetLoggerFromContext(ctx)
	log.Debug("Generating CSRV voucher")

	if c.Environment == "development" {
		return fmt.Sprintf("DEV-%d", rand.Int()), nil
	}

	prefix, group := "discord", "discord-giveaway"
	expires := time.Now().Add(24 * time.Duration(c.ExpirationDays) * time.Hour)
	uses := 1
	payload := dtos.GenerateVoucherPayload{
		Length:  values.VoucherLength,
		Charset: values.VoucherCharset,
		Prefix:  &prefix,
		Expires: &expires,
		GroupId: &group,
		Uses:    &uses,
		PerUser: &uses,
		Actions: []entities.VoucherAction{
			{
				WalletTx: map[monies.CurrencyCode]monies.Money{
					monies.PLN: monies.MustNew(int64(c.ValuePLN), monies.PLN),
				},
			},
		},
	}

	bodyPayload := new(bytes.Buffer)
	err := json.NewEncoder(bodyPayload).Encode(payload)
	if err != nil {
		return "", fmt.Errorf("GetCSRVCode json.NewEncoder failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/admin/voucher/generate", c.CraftserveUrl), bodyPayload)
	if err != nil {
		return "", err
	}

	req.AddCookie(&http.Cookie{Name: "user_access_token", Value: c.Secret})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.WithError(cerr).Error("GetCSRVCode failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GetCSRVCode failed with status: %d", resp.StatusCode)
	}

	var voucher entities.Voucher
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("GetCSRVCode io.ReadAll failed: %w", err)
	}

	err = json.Unmarshal(bodyBytes, &voucher)
	if err != nil {
		return "", fmt.Errorf("GetCSRVCode json.Unmarshal failed: %w with body: %s", err, string(bodyBytes))
	}

	return voucher.Id, nil
}
