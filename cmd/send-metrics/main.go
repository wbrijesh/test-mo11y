package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	ctx := context.Background()

	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint("localhost:4318"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	res, _ := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName("mo11y-test")),
	)

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(provider)
	meter := provider.Meter("mo11y-test-meter")

	// Counter (Sum)
	counter, _ := meter.Int64Counter("requests_total",
		metric.WithDescription("Total requests processed"),
	)
	counter.Add(ctx, 42)
	counter.Add(ctx, 13)

	// Gauge
	gauge, _ := meter.Float64Gauge("temperature_celsius",
		metric.WithDescription("Current temperature"),
	)
	gauge.Record(ctx, 23.5)

	// Histogram
	histogram, _ := meter.Float64Histogram("request_duration_ms",
		metric.WithDescription("Request duration in milliseconds"),
	)
	histogram.Record(ctx, 12.5)
	histogram.Record(ctx, 45.2)
	histogram.Record(ctx, 8.1)

	// Force flush and shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	provider.ForceFlush(shutdownCtx)
	if err := provider.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to shutdown: %v", err)
	}

	fmt.Println("✓ Sent metrics to mo11y (counter, gauge, histogram)")
}
