package logger

import (
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
)

func NewStdOutLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}

type otelLoggerParams struct {
	fx.In
	Lp *log.LoggerProvider
}

func NewOtelLogger(p otelLoggerParams) *slog.Logger {
	opts := []otelslog.Option{
		otelslog.WithLoggerProvider(p.Lp),
		otelslog.WithSource(true),
	}

	handler := otelslog.NewHandler("", opts...)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}
