package actuator

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

func (a *actuator) monitor(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.HealthCheckInterval)
	defer ticker.Stop()

	var cancelCurrent context.CancelFunc

	for {
		if cancelCurrent != nil {
			cancelCurrent()
		}

		checkCtx, cancel := context.WithCancel(ctx)
		cancelCurrent = cancel

		go a.healthCheck(checkCtx)

		select {
		case <-ctx.Done():
			slog.Info("stopping health check monitor")
			return
		case <-ticker.C:
			continue
		}
	}
}

func (a *actuator) healthCheck(ctx context.Context) {
	errCh := make(chan error, len(a.hc))

	var wg sync.WaitGroup
	for _, hc := range a.hc {
		wg.Go(func() {
			pCtx, cancel := context.WithTimeout(ctx, a.cfg.HealthCheckInterval)
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

func (a *actuator) liveness(w http.ResponseWriter, r *http.Request) {
	if a.Liveness() {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}

func (a *actuator) readiness(w http.ResponseWriter, r *http.Request) {
	if a.Readiness() {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}
