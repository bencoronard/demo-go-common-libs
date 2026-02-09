package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/bencoronard/demo-go-common-libs/constant"
	"github.com/bencoronard/demo-go-common-libs/dto"
	"github.com/labstack/echo/v5"
)

type ErrorHandlerFunc func(c *echo.Context, err error) error

func GlobalErrorHandler(fn ErrorHandlerFunc) func(c *echo.Context, err error) {
	return func(c *echo.Context, err error) {

		if resp, err := echo.UnwrapResponse(c.Response()); err == nil && resp.Committed {
			return
		}

		if fn != nil && fn(c, err) == nil {
			return
		}

		handleUncaughtError(c, err)
	}
}

func handleUncaughtError(c *echo.Context, err error) {
	pd := dto.ForStatusAndDetail(http.StatusInternalServerError, "Unhandled error at server side")

	switch {
	case errors.Is(err, constant.ErrInsufficientPermission):
		pd.Status = http.StatusForbidden
		pd.Detail = err.Error()
	case errors.Is(err, ErrMissingRequestHeader):
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()

	default:
		slog.Error(err.Error())
	}

	c.JSON(pd.Status, pd)
}
