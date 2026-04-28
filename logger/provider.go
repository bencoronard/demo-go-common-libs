package logger

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
)

type otelLoggerParams struct {
	fx.In
	Lp *log.LoggerProvider
}

func NewOtelLogger(p otelLoggerParams) (*slog.Logger, error) {
	handler := otelslog.NewHandler(
		"",
		otelslog.WithLoggerProvider(p.Lp),
		otelslog.WithSource(true),
	)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger, nil
}
