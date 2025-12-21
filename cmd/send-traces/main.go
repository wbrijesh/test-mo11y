package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	ctx := context.Background()

	// Create OTLP/HTTP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("mo11y-test"),
			attribute.String("environment", "test"),
			attribute.String("version", "0.1.0"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("mo11y-test-tracer")

	// Create spans
	rootCtx, rootSpan := tracer.Start(ctx, "test-root-operation")
	rootSpan.SetAttributes(
		attribute.String("operation.type", "test"),
		attribute.Int("operation.id", 12345),
	)
	rootSpan.AddEvent("root operation started")
	time.Sleep(10 * time.Millisecond)

	_, child1Span := tracer.Start(rootCtx, "test-child-1")
	child1Span.SetAttributes(
		attribute.String("child.name", "first-child"),
		attribute.Float64("child.value", 3.14),
	)
	child1Span.AddEvent("processing data")
	time.Sleep(5 * time.Millisecond)
	child1Span.End()

	child2Ctx, child2Span := tracer.Start(rootCtx, "test-child-2")
	child2Span.SetAttributes(
		attribute.String("child.name", "second-child"),
		attribute.Bool("child.success", true),
	)
	child2Span.AddEvent("validation complete")
	time.Sleep(8 * time.Millisecond)
	child2Span.End()

	_, child3Span := tracer.Start(child2Ctx, "test-child-3")
	child3Span.SetAttributes(
		attribute.String("child.name", "third-child"),
		attribute.Int64("child.timestamp", time.Now().Unix()),
	)
	child3Span.AddEvent("nested operation")
	time.Sleep(3 * time.Millisecond)
	child3Span.End()

	rootSpan.AddEvent("root operation completed")
	rootSpan.End()

	// Shutdown to flush
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tp.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to shutdown: %v", err)
	}

	fmt.Println("✓ Sent 4 spans with 5 events to mo11y")
}
