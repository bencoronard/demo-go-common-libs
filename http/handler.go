package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bencoronard/demo-go-common-libs/constant"
	"github.com/bencoronard/demo-go-common-libs/dto"
	"github.com/bencoronard/demo-go-common-libs/jwt"
	"github.com/labstack/echo/v5"
)

type AppErrorHandlerFunc func(err error, pd *dto.ProblemDetail) error

func GlobalErrorHandler(fn AppErrorHandlerFunc) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if resp, err := echo.UnwrapResponse(c.Response()); err == nil && resp.Committed {
			return
		}

		status := http.StatusInternalServerError
		msg := "Unhandled error at server side"

		var sc echo.HTTPStatusCoder
		if errors.As(err, &sc) {
			if tmp := sc.StatusCode(); tmp != 0 {
				status = tmp
			}
		}

		if he, ok := errors.AsType[*echo.HTTPError](err); ok {
			status = he.Code
			msg = he.Message
		}

		pd := dto.ForStatusAndDetail(status, msg)
		pd = pd.WithProperty("timestamp", time.Now())
		pd = pd.WithProperty("trace", "demo-999")

		handled := false

		if fn != nil {
			handled = fn(err, &pd) == nil
		}

		if !handled {
			handleUnhandledError(err, &pd)
		}

		c.JSON(pd.Status, pd)
	}
}

func handleUnhandledError(err error, pd *dto.ProblemDetail) {
	switch {
	case errors.Is(err, constant.ErrInsufficientPermission):
		pd.Status = http.StatusForbidden
		pd.Detail = err.Error()
	case errors.Is(err, ErrMissingRequestHeader):
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()
	case errors.Is(err, jwt.ErrTokenVerificationFail),
		errors.Is(err, jwt.ErrTokenClaimsInvalid),
		errors.Is(err, jwt.ErrTokenMalformed):
		pd.Status = http.StatusUnauthorized
		pd.Detail = err.Error()
	default:
		slog.Error(err.Error())
	}
}
