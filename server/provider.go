package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bencoronard/demo-go-common-libs/actuator"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type ServerParams struct {
	fx.In
	Lc  fx.Lifecycle
	Sd  fx.Shutdowner
	Act actuator.Actuator `optional:"true"`
}

type HttpServer interface {
	Instance() *http.Server
	Configure() error
}

type HttpServerConfig struct {
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
	Srv HttpServer
}

func ServeHttp(p httpServerParams) error {
	if err := p.Srv.Configure(); err != nil {
		return err
	}

	p.Lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			slog.Info(
				"initiated HTTP server startup",
				"pid", os.Getpid(),
				"addr", p.Srv.Instance().Addr,
			)
			go func() {
				if err := p.Srv.Instance().ListenAndServe(); err != http.ErrServerClosed {
					slog.Error("failed to start HTTP server", "error", err)
					p.Sd.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("initiated HTTP server shutdown")
			return p.Srv.Instance().Shutdown(ctx)
		},
	})

	return nil
}

type GrpcServer interface {
	Instance() *grpc.Server
	Listener() net.Listener
	Configure() error
}

type grpcServerParams struct {
	ServerParams
	Srv GrpcServer
}

func ServeGrpc(p grpcServerParams) error {
	if err := p.Srv.Configure(); err != nil {
		return err
	}

	p.Lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			slog.Info(
				"initiated gRPC server startup",
				"pid", os.Getpid(),
				"addr", p.Srv.Listener().Addr(),
			)
			go func() {
				if err := p.Srv.Instance().Serve(p.Srv.Listener()); err != nil {
					slog.Error("failed to start gRPC server", "error", err)
					p.Sd.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("initiated gRPC server shutdown")
			stopped := make(chan struct{})
			go func() {
				p.Srv.Instance().GracefulStop()
				close(stopped)
			}()
			select {
			case <-stopped:
				return nil
			case <-ctx.Done():
				p.Srv.Instance().Stop()
				return nil
			}
		},
	})

	return nil
}
