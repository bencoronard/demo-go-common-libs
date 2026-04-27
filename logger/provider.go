package logger

import "log/slog"

func New() (*slog.Logger, error) {

	logger := slog.Default()

	slog.SetDefault(logger)

	return logger, nil
}
