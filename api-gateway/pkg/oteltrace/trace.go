package oteltrace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func NewOTLPExporter(ctx context.Context, otlpEndpoint string) (sdktrace.SpanExporter, error) {
	insecureOpt := otlptracehttp.WithInsecure()
	endpointOpt := otlptracehttp.WithEndpoint(otlpEndpoint)

	exporter, err := otlptracehttp.New(ctx, insecureOpt, endpointOpt)
	if err != nil {
		return nil, fmt.Errorf("observ.NewOTLPExporter: %w", err)
	}

	return exporter, nil
}

func NewTraceProvider(exporter sdktrace.SpanExporter, appName string) (*sdktrace.TracerProvider, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(appName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("observ.trace.NewTraceProvider: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return tp, nil
}
