package http_server

import (
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"net/http"
	"net/url"
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

func generateAuthorizeRedirectURI(baseURI, code string, state *string) (string, error) {
	u, err := url.Parse(baseURI)
	if err != nil {
		return "", fmt.Errorf("error in url.Parse: %w", err)
	}
	q := u.Query()
	q.Set("code", code)
	if state != nil {
		q.Set("state", *state)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func generateErrorResponse(baseURI, errType string, errDescription, errURI, state *string) (string, error) {
	u, err := url.Parse(baseURI)
	if err != nil {
		return "", fmt.Errorf("error in url.Parse: %w", err)
	}
	q := u.Query()
	q.Set("error", errType)
	if errDescription != nil {
		q.Set("error_description", *errDescription)
	}
	if errURI != nil {
		q.Set("err_uri", *errURI)
	}
	if state != nil {
		q.Set("state", *state)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// The consent screen has provided a result
func (srv *HTTPServer) GetAuthorize(c *CustomContext) error {
	var reqBody AuthorizeRequest
	if err := ValidateRequest(c, &reqBody); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// TODO: Validate response type, lookup client, validate scopes

	// TODO: Forward auth header to provider API and get user info back

	// TODO: Handle the flow for the response type
	var redirectURI string
	var err error
	redirectURI, err = generateAuthorizeRedirectURI("", "", nil)
	if err != nil {
		c.InternalError(err, "error in GenerateRedirectURI")
	}

	return c.Redirect(http.StatusFound, redirectURI)
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
		redirectURI, err := generateErrorResponse(reqBody.RedirectURI, AuthErrInvalidRequest, utils.Ptr("invalid grant_type"), nil, nil)
		if err != nil {
			c.InternalError(err, "error in generateErrorResponse")
		}
		return c.Redirect(http.StatusFound, redirectURI)
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
