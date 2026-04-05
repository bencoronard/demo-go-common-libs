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
	HealthCheckTimeout        time.Duration
	HealthCheckTimeoutPerTask time.Duration
}

type Params struct {
	fx.In
	Lc  fx.Lifecycle
	Hc  []HealthChecker
	Cfg Config
}

func New(p Params) (Actuator, error) {
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
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})

	return a, nil
}
