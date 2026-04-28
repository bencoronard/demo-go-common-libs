package http

import (
	"log/slog"

	"github.com/bencoronard/demo-go-common-libs/validator"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/fx"
)

type echoRouterParams struct {
	fx.In
	ErrHandler GlobalErrorHandler
	Logger     *slog.Logger        `optional:"true"`
	Val        validator.Validator `optional:"true"`
}

func NewEchoRouter(p echoRouterParams) *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = p.ErrHandler.GetHandler()

	if p.Logger != nil {
		e.Logger = p.Logger
	}

	if p.Val != nil {
		e.Validator = p.Val
	}

	e.Use(
		middleware.Recover(),
	)

	return e
}
