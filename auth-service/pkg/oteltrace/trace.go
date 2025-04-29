package oteltrace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewOTLPExporter(ctx context.Context, otlpEndpoint string) (sdktrace.SpanExporter, error) {
	insecureOpt := otlptracegrpc.WithInsecure()
	endpointOpt := otlptracegrpc.WithEndpoint(otlpEndpoint)

	exporter, err := otlptracegrpc.New(ctx, insecureOpt, endpointOpt)
	if err != nil {
		return nil, fmt.Errorf("observ.NewOTLPExporter: %w", err)
	}

	return exporter, nil
}

func NewTraceProvider(exporter sdktrace.SpanExporter, appName string) (*sdktrace.TracerProvider, error) {
	/*res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(appName),
		),
	)*/

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)

	return tp, nil
}
