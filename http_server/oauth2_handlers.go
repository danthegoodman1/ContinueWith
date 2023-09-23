package http_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/pg"
	"github.com/danthegoodman1/GoAPITemplate/provider_api"
	"github.com/danthegoodman1/GoAPITemplate/query"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"net/http"
	"strings"
	"time"
)

var (
	ResponseTypeAuthorizationCode = "code"
	ResponseTypeClientCredentials = "client_credentials"

	AuthErrInvalidRequest          = "invalid_request"
	AuthErrUnauthorizedClient      = "unauthorized_client"
	AuthErrAccessDenied            = "access_denied"
	AuthErrUnsupportedResponseType = "unsupported_response_type"
	AuthErrInvalidScope            = "invalid_scope"
	AuthErrServerError             = "server_error"
	AuthErrTemporarilyUnavailable  = "temporarily_unavailable"

	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeRefreshToken      = "refresh_token"

	BearerTokenType = "bearer"
	MacTokenType    = "mac"

	ClientUserID = "_client"
)

type (
	AuthorizeRequest struct {
		ResponseType string  `query:"response_type" validate:"require"`
		ClientID     string  `query:"client_id" validate:"require"`
		RedirectURI  string  `query:"redirect_uri"`
		Scope        string  `query:"scope"`
		State        *string `query:"state"`
	}
)

// The consent screen has provided a result
func (s *HTTPServer) GetAuthorize(c *CustomContext) error {
	logger := zerolog.Ctx(c.Request().Context())
	var reqBody AuthorizeRequest
	if err := ValidateRequest(c, &reqBody); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Update our logger to have the context
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("ClientID", reqBody.ClientID).Str("ResponseType", reqBody.ResponseType).Str("RedirectURI", reqBody.RedirectURI).Str("Scope", reqBody.Scope)
	})

	// Validate response type
	if reqBody.ResponseType != ResponseTypeAuthorizationCode && reqBody.ResponseType != ResponseTypeClientCredentials {
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrUnsupportedResponseType, nil, nil, reqBody.State)
	}

	// Handle flow for response type
	switch reqBody.ResponseType {
	case ResponseTypeAuthorizationCode:
		return s.handleGetAuthorizationCode(c, reqBody)
	case ResponseTypeClientCredentials:
		return s.handleGetClientCredentials(c, reqBody)
	default:
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrUnsupportedResponseType, nil, nil, reqBody.State)
	}
}

func (s *HTTPServer) handleGetAuthorizationCode(c *CustomContext, reqBody AuthorizeRequest) error {
	ctx := c.Request().Context()
	logger := zerolog.Ctx(ctx)

	// Lookup client and get scopes
	var client query.Client
	var scopes []query.Scope
	err := query.ReliableExec(ctx, pg.Pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		client, err = q.SelectClient(ctx, reqBody.ClientID)
		if err != nil {
			return fmt.Errorf("error in SelectClient: %w", err)
		}
		scopes, err = q.ListScopes(ctx)
		if err != nil {
			return fmt.Errorf("error in ListScopes: %w", err)
		}
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrUnauthorizedClient, utils.Ptr("unknown client_id"), nil, reqBody.State)
	}
	if err != nil {
		return c.InternalError(err, "error getting client info")
	}

	// Verify client not suspended
	if client.Suspended {
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrAccessDenied, utils.Ptr("client suspended"), nil, reqBody.State)
	}

	// Validate scopes
	requestedScopes := strings.Split(reqBody.Scope, " ")
	scopeIDs := lo.Map(scopes, func(item query.Scope, index int) string {
		return item.ID
	})

	_, unknownScopes := lo.Difference(scopeIDs, requestedScopes)
	if len(unknownScopes) > 0 {
		// The client provided scopes that we don't know about
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrInvalidScope, utils.Ptr(fmt.Sprintf("invalid scopes: %+v", unknownScopes)), nil, reqBody.State)
	}

	// Forward auth header to provider API and get user info back
	userInfo, err := provider_api.ExchangeAuthForUserInfo(ctx, utils.ProviderAPIUserExchange, utils.ProviderAuthHeader, c.Request().Header.Get(utils.ProviderAuthHeader))
	if err != nil {
		var errType, errDesc string
		if isClientError := errors.Is(err, provider_api.ErrClientError); isClientError {
			errType = AuthErrInvalidRequest
			errDesc = err.Error()
		} else {
			errType = AuthErrServerError
			errDesc = err.Error()
			logger.Error().Err(err).Msg("server error exchanging auth for user info")
		}
		return c.ReturnErrorResponse(reqBody.RedirectURI, errType, utils.Ptr(errDesc), nil, reqBody.State)
	}

	// Insert the authorization code that can be exchanged for a token pair
	authCode := utils.GenRandomIDWithSize("ac_", 10)
	err = query.ReliableExec(ctx, pg.Pool, time.Second*10, func(ctx context.Context, q *query.Queries) error {
		return q.InsertAuthorizationCode(ctx, query.InsertAuthorizationCodeParams{
			ID:       authCode,
			UserID:   userInfo.UserID,
			Scopes:   nil,
			Expires:  time.Now().Add(time.Minute * 10),
			ClientID: client.ID,
		})
	})
	if err != nil {
		logger.Error().Err(err).Msg("error in InsertAuthorizationCode")
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrServerError, utils.Ptr("internal server error"), nil, reqBody.State)
	}

	return c.ReturnAuthorizeRedirectURI(reqBody.RedirectURI, authCode, nil)
}

func (s *HTTPServer) handleGetClientCredentials(c *CustomContext, reqBody AuthorizeRequest) error {
	ctx := c.Request().Context()

	// Lookup client
	var client query.Client
	clientAccessTokenID := utils.GenRandomIDWithSize("ca_", 16)
	err := query.ReliableExec(ctx, pg.Pool, time.Second*10, func(ctx context.Context, q *query.Queries) (err error) {
		client, err = q.SelectClient(ctx, reqBody.ClientID)
		if err != nil {
			return fmt.Errorf("error in SelectClient: %w", err)
		}

		// Insert a client credentials access token
		err = q.InsertAccessToken(ctx, query.InsertAccessTokenParams{
			ID:           clientAccessTokenID,
			ClientID:     reqBody.ClientID,
			RefreshToken: nil,
			UserID:       "",
			Scopes:       nil,
			Expires:      time.Now().Add(time.Second * time.Duration(utils.AccessTokenExpireSeconds)),
		})
		if err != nil {
			return fmt.Errorf("error in InsertAccessToken: %w", err)
		}
		return
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrUnauthorizedClient, utils.Ptr("unknown client_id"), nil, reqBody.State)
	}
	if err != nil {
		return c.InternalError(err, "error getting client info")
	}

	// Verify client not suspended
	if client.Suspended {
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrAccessDenied, utils.Ptr("client suspended"), nil, reqBody.State)
	}

	return c.JSON(http.StatusOK, AccessTokenResponse{
		AccessToken:  clientAccessTokenID,
		TokenType:    BearerTokenType,
		ExpiresIn:    int(utils.AccessTokenExpireSeconds),
		RefreshToken: "", // will be omitted
	})
}

type (
	// ClientSecret is actually not required (which makes sense)
	// according to the RFC: https://datatracker.ietf.org/doc/html/rfc6749#section-4.1.3
	AccessTokenRequest struct {
		ClientID string `query:"client_id" validate:"require"`

		RedirectURI string `query:"redirect_uri" validate:"require"`
		GrantType   string `query:"grant_type" validate:"require"`

		RefreshToken *string `query:"refresh_token"`
		Code         *string `query:"refresh_token"`
	}

	AccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		// "bearer" or "mac"
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}
)

// The client is exchanging an authorizaqtion code for an access token, or refreshing an access token
func (s *HTTPServer) PostAccessToken(c *CustomContext) error {
	var reqBody AccessTokenRequest
	if err := ValidateRequest(c, &reqBody); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	switch reqBody.GrantType {
	case GrantTypeAuthorizationCode:
		if reqBody.Code == nil {
			return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrInvalidRequest, utils.Ptr("missing code"), nil, nil)
		}
		return s.handleAuthorizationCodeRequest(c, reqBody)
	case GrantTypeRefreshToken:
		if reqBody.RefreshToken == nil {
			return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrInvalidRequest, utils.Ptr("missing refresh token"), nil, nil)
		}
		return s.handleRefreshTokenRequest(c, reqBody)
	default:
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrInvalidRequest, utils.Ptr("invalid grant_type"), nil, nil)
	}
}

func (s *HTTPServer) handleAuthorizationCodeRequest(c *CustomContext, request AccessTokenRequest) error {
	ctx := c.Request().Context()
	logger := zerolog.Ctx(ctx)

	// Generate token pair
	refreshTokenID := utils.GenRandomIDWithSize("r_", 16)
	accessTokenID := utils.GenRandomIDWithSize("a_", 16)
	err := query.ReliableExecInTx(ctx, pg.Pool, time.Second*20, func(ctx context.Context, q *query.Queries) error {
		if utils.IsPostgres {
			err := q.SetIsolationLevel(ctx, query.Serializable)
			if err != nil {
				return fmt.Errorf("error in SetIsolationLevel: %w", err)
			}
		}
		code, err := q.DeleteAuthorizationCode(ctx, *request.Code)
		if err != nil {
			return fmt.Errorf("error in SelectAuthorizationCode: %w", err)
		}

		// Insert the tokens
		err = q.InsertRefreshToken(ctx, query.InsertRefreshTokenParams{
			ID:       refreshTokenID,
			ClientID: code.ClientID,
			UserID:   code.UserID,
			Scopes:   code.Scopes,
			Expires:  time.Now().Add(time.Second * time.Duration(utils.RefreshTokenExpireSeconds)),
		})
		if err != nil {
			return fmt.Errorf("error in InsertRefreshToken: %w", err)
		}
		err = q.InsertAccessToken(ctx, query.InsertAccessTokenParams{
			ID:           accessTokenID,
			ClientID:     code.ClientID,
			UserID:       code.UserID,
			Scopes:       code.Scopes,
			Expires:      time.Now().Add(time.Second * time.Duration(utils.AccessTokenExpireSeconds)),
			RefreshToken: utils.Ptr(refreshTokenID),
		})
		if err != nil {
			return fmt.Errorf("error in InsertAccessToken: %w", err)
		}

		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c.ReturnErrorResponse(request.RedirectURI, AuthErrInvalidRequest, utils.Ptr("code not found"), nil, nil)
	}
	if err != nil {
		logger.Error().Err(err).Msg("error exchanging auth code for tokens in DB")
		return c.ReturnErrorResponse(request.RedirectURI, AuthErrInvalidRequest, utils.Ptr("internal server error"), nil, nil)
	}

	return c.JSON(http.StatusOK, AccessTokenResponse{
		AccessToken:  accessTokenID,
		TokenType:    BearerTokenType,
		ExpiresIn:    int(utils.AccessTokenExpireSeconds),
		RefreshToken: refreshTokenID,
	})
}

func (s *HTTPServer) handleRefreshTokenRequest(c *CustomContext, request AccessTokenRequest) error {
	ctx := c.Request().Context()
	logger := zerolog.Ctx(ctx)
	start := time.Now()

	// Lookup token
	newRefreshToken := ""
	newAccessToken := utils.GenRandomIDWithSize("a_", 16)
	err := query.ReliableExecInTx(ctx, pg.Pool, time.Second*20, func(ctx context.Context, q *query.Queries) error {
		if utils.IsPostgres {
			err := q.SetIsolationLevel(ctx, query.Serializable)
			if err != nil {
				return fmt.Errorf("error in SetIsolationLevel: %w", err)
			}
		}

		refreshToken, err := q.SelectValidRefreshToken(ctx, *request.RefreshToken)
		if err != nil {
			return fmt.Errorf("error in SelectValidRefreshToken: %w", err)
		}

		expired := start.After(refreshToken.Expires)
		if expired {
			// If expired, we need to make a new one
			newRefreshToken = utils.GenRandomIDWithSize("r_", 16)
			err = q.RevokeRefreshToken(ctx, refreshToken.ID)
			if err != nil {
				return fmt.Errorf("error in RevokeRefreshToken: %w", err)
			}

			err = q.InsertRefreshToken(ctx, query.InsertRefreshTokenParams{
				ID:       newRefreshToken,
				ClientID: refreshToken.ClientID,
				UserID:   refreshToken.UserID,
				Scopes:   refreshToken.Scopes,
				Expires:  time.Now().Add(time.Second * time.Duration(utils.RefreshTokenExpireSeconds)),
			})
			if err != nil {
				return fmt.Errorf("error in InsertRefreshToken: %w", err)
			}
		}

		// Insert the new access token
		err = q.InsertAccessToken(ctx, query.InsertAccessTokenParams{
			ID:           newAccessToken,
			ClientID:     refreshToken.ClientID,
			UserID:       refreshToken.UserID,
			Scopes:       refreshToken.Scopes,
			Expires:      time.Now().Add(time.Second * time.Duration(utils.AccessTokenExpireSeconds)),
			RefreshToken: utils.Ptr(lo.Ternary(expired, newRefreshToken, refreshToken.ID)),
		})
		if err != nil {
			return fmt.Errorf("error in InsertAccessToken: %w", err)
		}

		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c.ReturnErrorResponse(request.RedirectURI, AuthErrInvalidRequest, utils.Ptr("refresh token not found"), nil, nil)
	}
	if err != nil {
		logger.Error().Err(err).Msg("error exchanging auth code for tokens in DB")
		return c.ReturnErrorResponse(request.RedirectURI, AuthErrInvalidRequest, utils.Ptr("internal server error"), nil, nil)
	}

	return c.JSON(http.StatusOK, AccessTokenResponse{
		AccessToken:  newAccessToken,
		TokenType:    BearerTokenType,
		ExpiresIn:    int(utils.AccessTokenExpireSeconds),
		RefreshToken: newRefreshToken, // omitempty, will only be included if old expired
	})
}
