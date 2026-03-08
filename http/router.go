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

type RouterStartParams struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner
	Router     Router
}

func Start(p RouterStartParams) {
	s := http.Server{
		Addr:    fmt.Sprintf(":%d", p.Router.Port()),
		Handler: p.Router.Handler(),
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			slog.Info(
				"HTTP server starting...",
				"pid", os.Getpid(),
				"port", p.Router.Port(),
			)
			go func() {
				if err := s.ListenAndServe(); err != http.ErrServerClosed {
					slog.Error("Failed to start HTTP server", "error", err)
					p.Shutdowner.Shutdown()
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
