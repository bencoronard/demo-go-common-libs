package logger

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func New() (*slog.Logger, error) {

	handler := otelslog.NewHandler("")

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger, nil
}
