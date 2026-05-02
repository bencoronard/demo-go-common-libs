package logger

import (
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
)

func NewStdOutLogger() (*slog.Logger, error) {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger, nil
}

type otelLoggerParams struct {
	fx.In
	Lp *log.LoggerProvider
}

func NewOtelLogger(p otelLoggerParams) (*slog.Logger, error) {
	opts := []otelslog.Option{
		otelslog.WithLoggerProvider(p.Lp),
		otelslog.WithSource(true),
	}

	handler := otelslog.NewHandler("", opts...)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger, nil
}
