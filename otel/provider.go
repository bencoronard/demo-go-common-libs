package otel

import (
	"context"

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

type OtelParams struct {
	fx.In
	Lc   fx.Lifecycle
	Prop propagation.TextMapPropagator
	Res  *resource.Resource
}

func NewPropagator() (propagation.TextMapPropagator, error) {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	return prop, nil
}

func NewTracerProvider(p OtelParams) (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		return nil, err
	}

	provider := trace.NewTracerProvider(
		trace.WithResource(p.Res),
		trace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(provider)

	p.Lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return provider.Shutdown(ctx)
		},
	})

	return provider, nil
}

func NewMeterProvider(p OtelParams) (*metric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(context.Background())
	if err != nil {
		return nil, err
	}

	reader := metric.NewPeriodicReader(exporter)

	provider := metric.NewMeterProvider(
		metric.WithResource(p.Res),
		metric.WithReader(reader),
	)

	otel.SetMeterProvider(provider)

	p.Lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return provider.Shutdown(ctx)
		},
	})

	return provider, nil
}

func NewLoggerProvider(p OtelParams) (*log.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(context.Background())
	if err != nil {
		return nil, err
	}

	processor := log.NewBatchProcessor(exporter)

	provider := log.NewLoggerProvider(
		log.WithResource(p.Res),
		log.WithProcessor(processor),
	)

	global.SetLoggerProvider(provider)

	p.Lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return provider.Shutdown(ctx)
		},
	})

	return provider, nil
}
