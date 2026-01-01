package http

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	"go.uber.org/fx"
)

type Router interface {
	Listen(port int) (net.Listener, error)
	Serve(l net.Listener) error
	Shutdown(ctx context.Context) error
	ListeningPort() int
	RegisterRoutes() error
	RegisterMiddlewares() error
}

func Start(lc fx.Lifecycle, sd fx.Shutdowner, r Router) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := r.Listen(r.ListeningPort())
			if err != nil {
				return err
			}

			slog.Info(
				"HTTP server listening",
				"pid", os.Getpid(),
				"port", r.ListeningPort(),
			)

			go func() {
				if err := r.Serve(ln); err != nil && err != http.ErrServerClosed {
					slog.Error("HTTP server failed", "error", err)
					sd.Shutdown()
				}
			}()

			return nil
		},

		OnStop: func(ctx context.Context) error {
			slog.Info("HTTP server shutting down...")
			return r.Shutdown(ctx)
		},
	})
}
