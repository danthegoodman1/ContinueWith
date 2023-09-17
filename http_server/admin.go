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
	UserID           string
	Created, Expires time.Time
	Scopes           []string
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
		UserID:  accessToken.UserID,
		Created: accessToken.Created,
		Expires: accessToken.Expires,
		Scopes:  accessToken.Scopes,
	})
}
