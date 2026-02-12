package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go.uber.org/fx"
)

type Router interface {
	Port() int
	Handler() http.Handler
	RegisterMiddlewares()
	RegisterRoutes()
}

func Start(lc fx.Lifecycle, sd fx.Shutdowner, r Router) {
	s := http.Server{
		Addr:    fmt.Sprintf(":%d", r.Port()),
		Handler: r.Handler(),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			slog.Info(
				"HTTP server starting...",
				"pid", os.Getpid(),
				"port", r.Port(),
			)
			go func() {
				if err := s.ListenAndServe(); err != http.ErrServerClosed {
					slog.Error("Failed to start HTTP server", "error", err)
					sd.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("HTTP server shutting down...")
			return s.Shutdown(ctx)
		},
	})
}
