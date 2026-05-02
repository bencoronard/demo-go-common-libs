package actuator

import (
	"context"
	"net"
	"net/http"
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
	Lifecycle      fx.Lifecycle
	Shutdowner     fx.Shutdowner
	HealthCheckers []HealthChecker `group:"healthcheck"`
	Config         Config
}

func New(p params) (Actuator, error) {
	a := &actuator{
		healthCheckers: p.HealthCheckers,
		config:         p.Config,
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go a.monitor(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})

	if p.Config.Host != "" && p.Config.Port != 0 {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /actuator/liveness", a.liveness)
		mux.HandleFunc("GET /actuator/readiness", a.readiness)

		server := &http.Server{
			Addr:              net.JoinHostPort(p.Config.Host, strconv.Itoa(p.Config.Port)),
			Handler:           mux,
			ReadTimeout:       2 * time.Second,
			ReadHeaderTimeout: 1 * time.Second,
			WriteTimeout:      2 * time.Second,
			IdleTimeout:       10 * time.Second,
			MaxHeaderBytes:    4 << 10,
		}

		p.Lifecycle.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err := server.ListenAndServe(); err != http.ErrServerClosed {
						p.Shutdowner.Shutdown()
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return server.Shutdown(ctx)
			},
		})
	}

	return a, nil
}
