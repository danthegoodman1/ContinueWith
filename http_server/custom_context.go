package http_server

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/danthegoodman1/GoAPITemplate/gologger"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type CustomContext struct {
	echo.Context
	RequestID string
	UserID    string
}

func CreateReqContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqID := uuid.NewString()
		ctx := context.WithValue(c.Request().Context(), gologger.ReqIDKey, reqID)
		ctx = logger.WithContext(ctx)
		c.SetRequest(c.Request().WithContext(ctx))
		logger := zerolog.Ctx(ctx)
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("reqID", reqID)
		})
		cc := &CustomContext{
			Context:   c,
			RequestID: reqID,
		}
		return next(cc)
	}
}

// Casts to custom context for the handler, so this doesn't have to be done per handler
func ccHandler(h func(*CustomContext) error) echo.HandlerFunc {
	// TODO: Include the path?
	return func(c echo.Context) error {
		return h(c.(*CustomContext))
	}
}

func (c *CustomContext) internalErrorMessage() string {
	return "internal error, request id: " + c.RequestID
}

func (c *CustomContext) InternalError(err error, msg string) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		zerolog.Ctx(c.Request().Context()).Warn().CallerSkipFrame(1).Msg(err.Error())
	} else {
		zerolog.Ctx(c.Request().Context()).Error().CallerSkipFrame(1).Err(err).Msg(msg)
	}
	return c.String(http.StatusInternalServerError, c.internalErrorMessage())
}

func (c *CustomContext) ReturnAuthorizeRedirectURI(baseURI, code string, state *string) error {
	u, err := url.Parse(baseURI)
	if err != nil {
		return c.InternalError(err, "error in url.Parse")
	}
	q := u.Query()
	q.Set("code", code)
	if state != nil {
		q.Set("state", *state)
	}
	u.RawQuery = q.Encode()
	return c.Redirect(http.StatusFound, u.String())
}

func (c *CustomContext) ReturnErrorResponse(baseURI, errType string, errDescription, errURI, state *string) error {
	u, err := url.Parse(baseURI)
	if err != nil {
		return c.InternalError(err, "error in url.Parse")
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
	return c.Redirect(http.StatusNotFound, u.String())
}
