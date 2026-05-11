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
	ready          atomic.Bool
	healthCheckers []HealthChecker
	config         Config
}

func (a *actuator) Liveness() bool {
	return true
}

func (a *actuator) Readiness() bool {
	return a.ready.Load()
}

func (a *actuator) monitor(ctx context.Context) {
	ticker := time.NewTicker(a.config.HealthCheckInterval)
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
			return
		case <-ticker.C:
			continue
		}
	}
}

func (a *actuator) healthCheck(ctx context.Context) {
	errCh := make(chan error, len(a.healthCheckers))

	var wg sync.WaitGroup
	for _, hc := range a.healthCheckers {
		wg.Go(func() {
			checkCtx, cancel := context.WithTimeout(ctx, a.config.HealthCheckInterval)
			defer cancel()
			if err := hc.Check(checkCtx); err != nil {
				slog.Error(fmt.Sprintf("healthcheck failed for resource: %s", hc.Name()), "error", err)
				errCh <- err
			}
		})
	}

	wg.Wait()
	close(errCh)

	ready := true
	for range errCh {
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
