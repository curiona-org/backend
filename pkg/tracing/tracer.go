package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	t "go.opentelemetry.io/otel/trace"
)

type ProviderConfig struct {
	OTLPExporterEndpoint string
	AppName              string
	AppEnv               string
}

func NewProvider(ctx context.Context, cfg ProviderConfig) (*trace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPExporterEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.AppName),          // Service Name
			semconv.DeploymentEnvironmentKey.String(cfg.AppEnv), // Deployment Environment
		)),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	return tp, nil
}

func GetTraceID(ctx context.Context) string {
	spanCtx := t.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID := spanCtx.TraceID()
		return traceID.String()
	}
	return ""
}
