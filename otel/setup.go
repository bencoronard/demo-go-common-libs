package otel

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

type OtelParams struct {
	fx.In
	Lifecycle      fx.Lifecycle
	TracerProvider *trace.TracerProvider `optional:"true"`
	MeterProvider  *metric.MeterProvider `optional:"true"`
	LoggerProvider *log.LoggerProvider   `optional:"true"`
}

func InitOtel(p OtelParams) {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTextMapPropagator(prop)

	if p.TracerProvider != nil {
		otel.SetTracerProvider(p.TracerProvider)
	}
	if p.MeterProvider != nil {
		otel.SetMeterProvider(p.MeterProvider)
	}
	if p.LoggerProvider != nil {
		global.SetLoggerProvider(p.LoggerProvider)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			var errs []error

			if p.TracerProvider != nil {
				errs = append(errs, p.TracerProvider.Shutdown(ctx))
			}
			if p.MeterProvider != nil {
				errs = append(errs, p.MeterProvider.Shutdown(ctx))
			}
			if p.LoggerProvider != nil {
				errs = append(errs, p.LoggerProvider.Shutdown(ctx))
			}

			return errors.Join(errs...)
		},
	})
}
