package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type ServerParams struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner
}

type HTTPServer interface {
	Instance() *http.Server
	Configure() error
}

type HTTPServerConfig struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

type httpServerParams struct {
	ServerParams
	Server HTTPServer
}

func ServeHTTP(p httpServerParams) error {
	if err := p.Server.Configure(); err != nil {
		return err
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := p.Server.Instance().ListenAndServe(); err != http.ErrServerClosed {
					p.Shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return p.Server.Instance().Shutdown(ctx)
		},
	})

	return nil
}

type GRPCServer interface {
	Instance() *grpc.Server
	Listener() net.Listener
	Configure() error
}

type grpcServerParams struct {
	ServerParams
	Server GRPCServer
}

func ServeGRPC(p grpcServerParams) error {
	if err := p.Server.Configure(); err != nil {
		return err
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := p.Server.Instance().Serve(p.Server.Listener()); err != nil {
					p.Shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopped := make(chan struct{})
			go func() {
				p.Server.Instance().GracefulStop()
				close(stopped)
			}()
			select {
			case <-stopped:
				return nil
			case <-ctx.Done():
				p.Server.Instance().Stop()
				return nil
			}
		},
	})

	return nil
}
