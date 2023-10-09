package http_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/pg"
	"github.com/danthegoodman1/GoAPITemplate/query"
	"github.com/jackc/pgx/v5"
	"net/http"
	"time"
)

type VerifyAccessTokenResponse struct {
	UserID               string
	CreatedMS, ExpiresMS int64
	Scopes               []string
}

func (s *HTTPServer) CheckAccessToken(c *CustomContext) error {
	ctx := c.Request().Context()
	accessTokenID := c.Param("accessToken")

	var accessToken query.AccessToken
	err := query.ReliableExec(ctx, pg.Pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		accessToken, err = q.SelectValidAccessToken(ctx, accessTokenID)
		if err != nil {
			return fmt.Errorf("error in SelectValidAccessToken: %w", err)
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c.String(http.StatusNotFound, "no code found")
	}
	if err != nil {
		c.InternalError(err, "error getting access token")
	}

	return c.JSON(http.StatusOK, VerifyAccessTokenResponse{
		UserID:    accessToken.UserID,
		CreatedMS: accessToken.Created.UnixMilli(),
		ExpiresMS: accessToken.Expires.UnixMilli(),
		Scopes:    accessToken.Scopes,
	})
}

type ClientResponse struct {
	ID        string
	Suspended bool
	Name      string
	Created   time.Time
	Updated   time.Time
}

func (s *HTTPServer) GetClientFromID(c *CustomContext) error {
	ctx := c.Request().Context()
	clientID := c.Param("clientID")

	var client query.Client
	err := query.ReliableExec(ctx, pg.Pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		client, err = q.SelectClient(ctx, clientID)
		if err != nil {
			return fmt.Errorf("error in SelectClient: %w", err)
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c.String(http.StatusNotFound, "client not found")
	}
	if err != nil {
		c.InternalError(err, "error getting client")
	}

	return c.JSON(http.StatusOK, ClientResponse{
		ID:        client.ID,
		Suspended: client.Suspended,
		Name:      client.Name,
		Created:   client.Created,
		Updated:   client.Updated,
	})
}
