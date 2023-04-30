package services

import (
	"context"
	"csrvbot/pkg/logger"
	"encoding/json"
	"net/http"
)

type CsrvClient struct {
	Secret string
}

func NewCsrvClient(secret string) *CsrvClient {
	return &CsrvClient{Secret: secret}
}

type VoucherResponse struct {
	Code string `json:"code"`
}

func (c *CsrvClient) GetCSRVCode(ctx context.Context) (string, error) {
	log := logger.GetLoggerFromContext(ctx)
	req, err := http.NewRequest("POST", "https://craftserve.pl/api/generate_voucher", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("csrvbot", c.Secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var code VoucherResponse
	err = json.NewDecoder(resp.Body).Decode(&code)
	if err != nil {
		return "", err
	}
	err = resp.Body.Close()
	if err != nil {
		log.WithError(err).Error("getCSRVCode resp.Body.Close()")
	}
	return code.Code, nil
}
