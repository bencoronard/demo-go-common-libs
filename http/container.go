package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.uber.org/fx"
)

type Container interface {
	ServeHTTP() error
}

func InitContainer(lc fx.Lifecycle, sd fx.Shutdowner, c Container) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			errChan := make(chan error, 1)

			go func() {
				if err := c.ServeHTTP(); err != nil && err != http.ErrServerClosed {
					errChan <- err
				}
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errChan:
				return err
			case <-time.After(100 * time.Millisecond):
				go func() {
					if err := <-errChan; err != nil {
						slog.Error(err.Error())
						sd.Shutdown()
					}
				}()
				return nil
			}
		},
	})
}
