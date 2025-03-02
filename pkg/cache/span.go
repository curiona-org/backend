package cache

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

//nolint:spancheck
func spanWithKey(ctx context.Context, tracer trace.Tracer, method, key string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(
		attribute.String("db.operation.parameter", key),
	)
	return ctx, span
}

//nolint:spancheck
func spanWithOperationKey(ctx context.Context, tracer trace.Tracer, method, operation, key string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(
		semconv.DBSystemRedis,
		semconv.DBOperationKey.String(operation),
		attribute.String("db.operation.parameter", key),
	)
	return ctx, span
}
