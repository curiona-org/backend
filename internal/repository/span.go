package repository

import (
	"context"

	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

//nolint:spancheck
func spanWithSelectQuery(ctx context.Context, tracer trace.Tracer, method, query string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.DBStatementKey.String(query),
		semconv.DBOperationKey.String("SELECT"))
	return ctx, span
}

//nolint:spancheck
func spanWithInsertQuery(ctx context.Context, tracer trace.Tracer, method, query string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.DBStatementKey.String(query),
		semconv.DBOperationKey.String("INSERT"))
	return ctx, span
}

//nolint:spancheck
func spanWithUpdateQuery(ctx context.Context, tracer trace.Tracer, method, query string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.DBStatementKey.String(query),
		semconv.DBOperationKey.String("UPDATE"))
	return ctx, span
}

//nolint:spancheck
func spanWithDeleteQuery(ctx context.Context, tracer trace.Tracer, method, query string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, method)
	span.SetAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.DBStatementKey.String(query),
		semconv.DBOperationKey.String("DELETE"))
	return ctx, span
}
