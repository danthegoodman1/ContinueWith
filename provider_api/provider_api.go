package provider_api

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/samber/lo"
	"io"
	"net/http"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrClientError = errors.New("client error (4xx)")
	ErrServerError = errors.New("server error (5xx)")
)

type ExchangeAuthForUserResponse struct {
	UserID string
}

func ExchangeAuthForUserInfo(ctx context.Context, targetURL, authHeaderKey, authHeaderVal string) (*ExchangeAuthForUserResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error in http.NewRequestWithContext: %w", err)
	}

	req.Header.Set(authHeaderKey, authHeaderVal)
	req.Header.Set("x-continuewith-auth", utils.AdminKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error in http.DefaultClient.Do: %w", err)
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("error in io.ReadAll: %w", err)
	}

	if res.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("%d - %s -- %w", res.StatusCode, resBytes, lo.Ternary(res.StatusCode >= 500, ErrServerError, ErrClientError))
	}

	var resBody ExchangeAuthForUserResponse
	err = sonic.Unmarshal(resBytes, &res)
	if err != nil {
		return nil, fmt.Errorf("error in sonic.Unmarshal: %w", err)
	}

	return &resBody, nil
}
