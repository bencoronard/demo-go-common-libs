package http

import (
	"log/slog"
	"slices"

	"github.com/bencoronard/demo-go-common-libs/validator"
	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

type Config struct {
	EnableAccessLog bool
}

type echoRouterParams struct {
	fx.In
	Cfg        Config
	ErrHandler GlobalErrorHandler
	Logger     *slog.Logger                  `optional:"true"`
	Val        validator.Validator           `optional:"true"`
	Pp         propagation.TextMapPropagator `optional:"true"`
	Tp         *trace.TracerProvider         `optional:"true"`
	Mp         *metric.MeterProvider         `optional:"true"`
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

	middlewares := []echo.MiddlewareFunc{
		middleware.Recover(),
		otelMiddleware(p.Pp, p.Tp, p.Mp),
		accessLogMiddleware(p.Logger, p.Cfg.EnableAccessLog),
	}

	e.Use(compact(middlewares)...)

	return e
}

func compact(mws []echo.MiddlewareFunc) []echo.MiddlewareFunc {
	return slices.DeleteFunc(mws, func(mw echo.MiddlewareFunc) bool {
		return mw == nil
	})
}

func otelMiddleware(pp propagation.TextMapPropagator, tp *trace.TracerProvider, mp *metric.MeterProvider) echo.MiddlewareFunc {
	if tp == nil && mp == nil {
		return nil
	}
	return echootel.NewMiddlewareWithConfig(echootel.Config{
		Propagators:    pp,
		TracerProvider: tp,
		MeterProvider:  mp,
	})
}

func accessLogMiddleware(logger *slog.Logger, enabled bool) echo.MiddlewareFunc {
	if !enabled {
		return nil
	}
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:      true,
		LogProtocol:     true,
		LogRemoteIP:     true,
		LogMethod:       true,
		LogURI:          true,
		LogStatus:       true,
		LogResponseSize: true,
		LogUserAgent:    true,

		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("protocol", v.Protocol),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.Int64("response_size", v.ResponseSize),
				slog.String("user_agent", v.UserAgent),
			}

			level := slog.LevelInfo
			switch {
			case v.Error != nil:
				level = slog.LevelError
				attrs = append(attrs, slog.Any("error", v.Error))
			case v.Status >= 500:
				level = slog.LevelError
			case v.Status >= 400:
				level = slog.LevelWarn
			}

			logger.LogAttrs(c.Request().Context(), level, "request", attrs...)
			return nil
		},
	})
}
