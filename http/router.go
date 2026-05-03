package http

import (
	"log/slog"

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
	Config         Config
	ErrHandler     GlobalErrorHandler
	Logger         *slog.Logger                  `optional:"true"`
	Validator      validator.Validator           `optional:"true"`
	Propagator     propagation.TextMapPropagator `optional:"true"`
	TracerProvider *trace.TracerProvider         `optional:"true"`
	MeterProvider  *metric.MeterProvider         `optional:"true"`
}

func NewEchoRouter(p echoRouterParams) *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = p.ErrHandler.GetHandler()

	logger := slog.Default()
	if p.Logger != nil {
		logger = p.Logger
	}
	e.Logger = logger

	if p.Validator != nil {
		e.Validator = p.Validator
	}

	middlewares := []echo.MiddlewareFunc{middleware.Recover()}

	if p.Propagator != nil || p.TracerProvider != nil || p.MeterProvider != nil {
		middlewares = append(middlewares, otelMiddleware(p.Propagator, p.TracerProvider, p.MeterProvider))
	}

	if p.Config.EnableAccessLog {
		middlewares = append(middlewares, accessLogMiddleware(logger))
	}

	e.Use(middlewares...)

	return e
}

func otelMiddleware(pp propagation.TextMapPropagator, tp *trace.TracerProvider, mp *metric.MeterProvider) echo.MiddlewareFunc {
	return echootel.NewMiddlewareWithConfig(echootel.Config{
		Propagators:    pp,
		TracerProvider: tp,
		MeterProvider:  mp,
	})
}

func accessLogMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:     true,
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
			if v.Error != nil {
				attrs = append(attrs, slog.Any("error", v.Error))
			}
			logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "ACCESS", attrs...)
			return nil
		},
	})
}
