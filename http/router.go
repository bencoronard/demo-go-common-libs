package http

import (
	"github.com/bencoronard/demo-go-common-libs/validator"
	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
)

type EchoRouterParams struct {
	fx.In
	ErrHandler GlobalErrorHandler
	Val        validator.Validator `optional:"true"`
}

func NewEchoRouter(p EchoRouterParams) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = p.ErrHandler.GetHandler()
	if p.Val != nil {
		e.Validator = p.Val
	}
	return e
}
