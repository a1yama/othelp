// Package othelp provides thin helpers for OpenTelemetry instrumentation in Go.
//
// The core feature is the defer end(&err) pattern that automatically records
// errors and sets span status, eliminating the most common instrumentation bug.
package othelp

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// EndFunc is called via defer to end a span and automatically record errors.
// Pass a pointer to the function's named error return value.
type EndFunc func(errPtr *error)

// Tracer wraps an OpenTelemetry Tracer with helper methods.
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new Tracer with the given instrumentation name.
func NewTracer(name string, opts ...trace.TracerOption) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(name, opts...),
	}
}

// NewTracerWithProvider creates a new Tracer using the specified TracerProvider.
// Useful for testing or when not using the global provider.
func NewTracerWithProvider(name string, tp trace.TracerProvider, opts ...trace.TracerOption) *Tracer {
	return &Tracer{
		tracer: tp.Tracer(name, opts...),
	}
}

// Start begins a new span and returns an EndFunc for use with defer.
//
//	func GetUser(ctx context.Context, id string) (user *User, err error) {
//	    ctx, end := tracer.Start(ctx, "GetUser", othelp.Str("user.id", id))
//	    defer end(&err)
//	    // ...
//	}
func (t *Tracer) Start(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	ctx, span := t.tracer.Start(ctx, spanName, trace.WithAttributes(attrs...))
	return ctx, newEndFunc(span)
}

// OTelTracer returns the underlying OpenTelemetry Tracer for escape-hatch usage.
func (t *Tracer) OTelTracer() trace.Tracer {
	return t.tracer
}

// Start begins a new span using the global tracer provider.
// For package-level tracer reuse, prefer NewTracer instead.
func Start(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	ctx, span := otel.Tracer("").Start(ctx, spanName, trace.WithAttributes(attrs...))
	return ctx, newEndFunc(span)
}

func newEndFunc(span trace.Span) EndFunc {
	return func(errPtr *error) {
		if errPtr != nil && *errPtr != nil {
			span.RecordError(*errPtr)
			span.SetStatus(codes.Error, (*errPtr).Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}
}
