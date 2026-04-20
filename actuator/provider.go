package actuator

import (
	"context"
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
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

type params struct {
	fx.In
	Lc  fx.Lifecycle
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

	return a, nil
}
