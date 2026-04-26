package actuator

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/fx"
)

type Actuator interface {
	Liveness() bool
	Readiness() bool
}

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

type Config struct {
	Host                string
	Port                int
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

type params struct {
	fx.In
	Lc  fx.Lifecycle
	Sd  fx.Shutdowner
	Hc  []HealthChecker `group:"healthcheck"`
	Cfg Config
}

func New(p params) (Actuator, error) {
	a := &actuator{
		hc:  p.Hc,
		cfg: p.Cfg,
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.Lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go a.monitor(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})

	if p.Cfg.Host != "" && p.Cfg.Port != 0 {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /actuator/liveness", a.liveness)
		mux.HandleFunc("GET /actuator/readiness", a.readiness)

		server := &http.Server{
			Addr:              net.JoinHostPort(p.Cfg.Host, strconv.Itoa(p.Cfg.Port)),
			Handler:           mux,
			ReadTimeout:       2 * time.Second,
			ReadHeaderTimeout: 1 * time.Second,
			WriteTimeout:      2 * time.Second,
			IdleTimeout:       10 * time.Second,
			MaxHeaderBytes:    4 << 10,
		}

		p.Lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				slog.Info(
					"initiated actuator server startup",
					"pid", os.Getpid(),
					"addr", server.Addr,
				)
				go func() {
					if err := server.ListenAndServe(); err != http.ErrServerClosed {
						slog.Error("failed to start actuator server", "error", err)
						p.Sd.Shutdown()
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				slog.Info("initiated actuator server shutdown")
				return server.Shutdown(ctx)
			},
		})
	}

	return a, nil
}
