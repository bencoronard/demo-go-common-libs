package actuator

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type ResourceCheck func(ctx context.Context) error

type Actuator interface {
	Liveness() bool
	Readiness() bool
}

type ActuatorParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	DB        *gorm.DB `optional:"true"`
}

type actuatorImpl struct {
	ready  atomic.Bool
	checks []ResourceCheck
}

func NewActuator(p ActuatorParams) (Actuator, error) {
	var checks []ResourceCheck

	// DB check
	if p.DB != nil {
		sqlDB, err := p.DB.DB()
		if err != nil {
			return nil, err
		}
		checks = append(checks, func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		})
	}

	a := &actuatorImpl{
		checks: checks,
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

func (a *actuatorImpl) Liveness() bool {
	return true
}

func (a *actuatorImpl) Readiness() bool {
	return a.ready.Load()
}

func (a *actuatorImpl) monitor(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			ready := a.checkResources(pingCtx)
			cancel()
			a.ready.Store(ready)
		case <-ctx.Done():
			return
		}
	}
}

func (a *actuatorImpl) checkResources(ctx context.Context) bool {
	var wg sync.WaitGroup
	errCh := make(chan error, len(a.checks))

	for _, check := range a.checks {
		wg.Add(1)
		go func(c ResourceCheck) {
			defer wg.Done()
			if err := c(ctx); err != nil {
				errCh <- err
			}
		}(check)
	}

	wg.Wait()
	close(errCh)

	return len(errCh) == 0
}
