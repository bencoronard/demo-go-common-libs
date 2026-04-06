package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type ServerParams struct {
	fx.In
	Lc fx.Lifecycle
	Sd fx.Shutdowner
}

type HTTPServer interface {
	Instance() *http.Server
	Configure() error
}

type HTTPServerParams struct {
	ServerParams
	Srv HTTPServer
}

func ServeHTTP(p HTTPServerParams) error {
	if err := p.Srv.Configure(); err != nil {
		return err
	}

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
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

type GRPCServer interface {
	Instance() *grpc.Server
	Listener() net.Listener
	Configure() error
}

type GRPCServerParams struct {
	ServerParams
	Srv GRPCServer
}

func ServeGRPC(p GRPCServerParams) error {
	if err := p.Srv.Configure(); err != nil {
		return err
	}

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
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
