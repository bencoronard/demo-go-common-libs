package actuator

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/fx"
)

type actuator struct {
	ready          atomic.Bool
	healthCheckers []HealthChecker
	config         HealthCheckConfig
}

func (a *actuator) Liveness() bool {
	return true
}

func (a *actuator) Readiness() bool {
	return a.ready.Load()
}

type serverParams struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner
	Config     ServerConfig
}

func (a *actuator) ExposeHTTPEndpoints(p serverParams) error {
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
			slog.Info("actuator server started", "pid", os.Getpid(), "addr", server.Addr)
			go func() {
				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					slog.Error("actuator server startup failed", "error", err)
					p.Shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to shudown actuator server: %w", err)
			}
			return nil
		},
	})

	return nil
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
