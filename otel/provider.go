package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(ctx context.Context, endpoint string, batchTimeoutInSec int) (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(
			exporter,
			trace.WithBatchTimeout(time.Duration(batchTimeoutInSec)*time.Second)),
	)

	return provider, nil
}

func NewMeterProvider(ctx context.Context, endpoint string, samplingIntervalInSec int) (*metric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(
			exporter,
			metric.WithInterval(time.Duration(samplingIntervalInSec)*time.Second),
		)),
	)

	return provider, nil
}

func NewLoggerProvider(ctx context.Context, endpoint string, exportIntervalInSec int) (*log.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	provider := log.NewLoggerProvider(
		log.WithProcessor(
			log.NewBatchProcessor(
				exporter,
				log.WithExportInterval(time.Duration(exportIntervalInSec)*time.Second),
			),
		),
	)

	return provider, nil
}
