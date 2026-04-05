package actuator

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

type actuator struct {
	ready atomic.Bool
	hc    []HealthChecker
	cfg   Config
}

func (a *actuator) Liveness() bool {
	return true
}

func (a *actuator) Readiness() bool {
	return a.ready.Load()
}

func (a *actuator) healthCheck(ctx context.Context) {
	errCh := make(chan error, len(a.hc))

	var wg sync.WaitGroup
	for _, hc := range a.hc {
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic in health check %s: %v", hc.Name(), r)
				}
			}()

			pCtx, cancel := context.WithTimeout(ctx, a.cfg.HealthCheckTimeoutPerTask)
			defer cancel()

			if err := hc.Check(pCtx); err != nil {
				errCh <- fmt.Errorf("%s: %w", hc.Name(), err)
			}
		})
	}

	wg.Wait()
	close(errCh)

	ready := true
	for err := range errCh {
		slog.Error("health check failed", "error", err)
		ready = false
	}

	a.ready.Store(ready)
}

func (a *actuator) monitor(ctx context.Context) {
	jitter := rand.N(3000 * time.Millisecond)
	ticker := time.NewTicker(a.cfg.HealthCheckTimeout + jitter)
	defer ticker.Stop()

	a.healthCheck(ctx)

	for {
		select {
		case <-ticker.C:
			a.healthCheck(ctx)
		case <-ctx.Done():
			slog.Info("stopping health check monitor")
			return
		}
	}
}
