package actuator

import (
	"context"
	"time"

	"go.uber.org/fx"
)

type Actuator interface {
	Liveness() bool
	Readiness() bool
	ExposeHTTPEndpoints(p serverParams) error
}

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

type HealthCheckConfig struct {
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

type ServerConfig struct {
	Host string
	Port int
}

type params struct {
	fx.In
	Lifecycle      fx.Lifecycle
	HealthCheckers []HealthChecker `group:"healthcheck"`
	Config         HealthCheckConfig
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

	return a, nil
}
