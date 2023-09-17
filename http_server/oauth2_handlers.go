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
func (srv *HTTPServer) GetAuthorize(c *CustomContext) error {
	var reqBody AuthorizeRequest
	if err := ValidateRequest(c, &reqBody); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Validate response type
	if reqBody.ResponseType != ResponseTypeAuthorizationCode && reqBody.ResponseType != ResponseTypeClientCredentials {
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrUnsupportedResponseType, nil, nil, reqBody.State)
	}

	// Handle flow for response type
	switch reqBody.ResponseType {
	case ResponseTypeAuthorizationCode:
		return srv.handleGetAuthorizationCode(c, reqBody)
	case ResponseTypeClientCredentials:
		return srv.handleGetClientCredentials(c, reqBody)
	default:
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrUnsupportedResponseType, nil, nil, reqBody.State)
	}
}

func (srv *HTTPServer) handleGetAuthorizationCode(c *CustomContext, reqBody AuthorizeRequest) error {
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
		return c.InternalError(err, "")
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
	authCode := utils.GenRandomShortID()
	err = query.ReliableExec(ctx, pg.Pool, time.Second*10, func(ctx context.Context, q *query.Queries) error {
		return q.InsertAuthorizationCode(ctx, query.InsertAuthorizationCodeParams{
			ID:      authCode,
			UserID:  userInfo.UserID,
			Scopes:  nil,
			Expires: time.Now().Add(time.Minute * 10),
		})
	})

	return c.ReturnAuthorizeRedirectURI(reqBody.RedirectURI, authCode, nil)
}

func (srv *HTTPServer) handleGetClientCredentials(c *CustomContext, reqBody AuthorizeRequest) error {
	// TODO: implement
	return c.NoContent(http.StatusNotImplemented)
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
)

// The client is exchanging an authorizaqtion code for an access token, or refreshing an access token
func (srv *HTTPServer) PostAccessToken(c *CustomContext) error {
	var reqBody AccessTokenRequest
	if err := ValidateRequest(c, &reqBody); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	switch reqBody.GrantType {
	case GrantTypeAuthorizationCode:
		return srv.handleAuthorizationCodeRequest(c, reqBody)
	case GrantTypeRefreshToken:
		return srv.handleRefreshTokenRequest(c, reqBody)
	default:
		return c.ReturnErrorResponse(reqBody.RedirectURI, AuthErrInvalidRequest, utils.Ptr("invalid grant_type"), nil, nil)
	}
}

func (src *HTTPServer) handleAuthorizationCodeRequest(c *CustomContext, request AccessTokenRequest) error {
	// TODO: Lookup code
	// TODO: Generate token pair
	// TODO: Return tokens
}

func (src *HTTPServer) handleRefreshTokenRequest(c *CustomContext, request AccessTokenRequest) error {
	// TODO: Lookup code
	// TODO: Generate token pair (refreshing refresh if expired)
	// TODO: Return tokens
}
