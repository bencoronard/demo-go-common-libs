package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

func NewResource() *resource.Resource {
	return resource.Environment()
}

func NewPropagator() propagation.TextMapPropagator {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTextMapPropagator(prop)

	return prop
}

type params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Resource  *resource.Resource
}

func NewTracerProvider(p params) (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithResource(p.Resource),
		trace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(provider)

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := provider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to shutdown provider: %w", err)
			}
			return nil
		},
	})

	return provider, nil
}

func NewMeterProvider(p params) (*metric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	reader := metric.NewPeriodicReader(exporter)

	provider := metric.NewMeterProvider(
		metric.WithResource(p.Resource),
		metric.WithReader(reader),
	)

	otel.SetMeterProvider(provider)

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := provider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to shutdown provider: %w", err)
			}
			return nil
		},
	})

	return provider, nil
}

func NewLoggerProvider(p params) (*log.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	processor := log.NewBatchProcessor(exporter)

	provider := log.NewLoggerProvider(
		log.WithResource(p.Resource),
		log.WithProcessor(processor),
	)

	global.SetLoggerProvider(provider)

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := provider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to shutdown provider: %w", err)
			}
			return nil
		},
	})

	return provider, nil
}
