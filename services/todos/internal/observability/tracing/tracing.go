package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Init(ctx context.Context, defaultServiceName, serviceVersion, env string) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }

	if os.Getenv("OTEL_ENABLED") != "true" {
		slog.Info("tracing disabled (OTEL_ENABLED != true)")
		return noop, nil
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		slog.Warn("tracing disabled: OTEL_EXPORTER_OTLP_ENDPOINT not set")
		return noop, nil
	}

	serviceName := defaultServiceName
	if s := os.Getenv("OTEL_SERVICE_NAME"); s != "" {
		serviceName = s
	}

	isInsecure := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true"
	slog.Info("tracing init", "endpoint", endpoint, "service", serviceName, "insecure", isInsecure)

	// Build gRPC connection explicitly so connection errors surface at startup.
	grpcOpts := []grpc.DialOption{}
	if isInsecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(endpoint, grpcOpts...)
	if err != nil {
		return noop, fmt.Errorf("tracing: dial otlp endpoint %s: %w", endpoint, err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		_ = conn.Close()
		return noop, fmt.Errorf("tracing: create otlp exporter: %w", err)
	}

	attrs := []attribute.KeyValue{
		attribute.String("service.name", serviceName),
	}
	if serviceVersion != "" {
		attrs = append(attrs, attribute.String("service.version", serviceVersion))
	}
	if env != "" {
		attrs = append(attrs, attribute.String("deployment.environment", env))
	}

	res, err := resource.New(ctx, resource.WithAttributes(attrs...))
	if err != nil || res == nil {
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("tracing ready", "service", serviceName, "endpoint", endpoint)
	return tp.Shutdown, nil
}

// RecordError marks the span as errored and records the error event.
// Safe to call with a nil error (no-op).
func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(otelcodes.Error, err.Error())
}
