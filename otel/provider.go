package otel

import (
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

func NewTracerProvider(lc fx.Lifecycle) (*trace.TracerProvider, error) {
	provider := trace.NewTracerProvider()
	return provider, nil
}

func NewMeterProvider(lc fx.Lifecycle) (*metric.MeterProvider, error) {
	provider := metric.NewMeterProvider()
	return provider, nil
}

func NewLoggerProvider(lc fx.Lifecycle) (*log.LoggerProvider, error) {
	provider := log.NewLoggerProvider()
	return provider, nil
}
