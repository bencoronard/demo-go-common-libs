package http

import (
	"net/http"

	"github.com/bencoronard/demo-go-common-libs/dto"
	xjwt "github.com/bencoronard/demo-go-common-libs/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

type AuthHeaderResolver interface {
	ExtractClaims(r *http.Request) (jwt.MapClaims, error)
}

func NewHttpAuthHeaderResolver(verifier xjwt.Verifier) AuthHeaderResolver {
	return &authHeaderResolver{verifier: verifier}
}

type AppErrorHandler interface {
	Handle(err error, pd dto.ProblemDetail) (dto.ProblemDetail, bool)
}

type GlobalErrorHandler interface {
	GetHandler() func(c *echo.Context, err error)
}

type globalErrorHandlerParams struct {
	fx.In
	AppErrHandler  AppErrorHandler          `optional:"true"`
	TracerProvider *sdktrace.TracerProvider `optional:"true"`
}

func NewGlobalErrorHandler(p globalErrorHandlerParams) GlobalErrorHandler {
	return &globalErrorHandler{
		ah: p.AppErrHandler,
		tp: p.TracerProvider,
	}
}
