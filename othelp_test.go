package othelp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/a1yama/othelp"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupTestTracer(t *testing.T) (*othelp.Tracer, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	tracer := othelp.NewTracerWithProvider("test", tp)
	return tracer, exporter
}

func TestStart_Success(t *testing.T) {
	tracer, exporter := setupTestTracer(t)

	var err error
	func() {
		_, end := tracer.Start(context.Background(), "test-span",
			othelp.Str("key", "value"),
		)
		defer end(&err)
	}()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name != "test-span" {
		t.Errorf("expected span name %q, got %q", "test-span", span.Name)
	}
	if span.Status.Code != codes.Ok {
		t.Errorf("expected status Ok, got %v", span.Status.Code)
	}

	found := false
	for _, attr := range span.Attributes {
		if string(attr.Key) == "key" && attr.Value.AsString() == "value" {
			found = true
		}
	}
	if !found {
		t.Error("expected attribute key=value not found")
	}
}

func TestStart_Error(t *testing.T) {
	tracer, exporter := setupTestTracer(t)

	testErr := errors.New("something went wrong")
	var err error
	func() {
		_, end := tracer.Start(context.Background(), "error-span")
		defer end(&err)
		err = testErr
	}()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Status.Code != codes.Error {
		t.Errorf("expected status Error, got %v", span.Status.Code)
	}
	if span.Status.Description != "something went wrong" {
		t.Errorf("expected status description %q, got %q", "something went wrong", span.Status.Description)
	}

	if len(span.Events) == 0 {
		t.Fatal("expected error event to be recorded")
	}

	foundErr := false
	for _, event := range span.Events {
		if event.Name == "exception" {
			foundErr = true
		}
	}
	if !foundErr {
		t.Error("expected exception event not found")
	}
}

func TestStart_NilErrPtr(t *testing.T) {
	tracer, exporter := setupTestTracer(t)

	func() {
		_, end := tracer.Start(context.Background(), "nil-err-span")
		defer end(nil)
	}()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Ok {
		t.Errorf("expected status Ok, got %v", spans[0].Status.Code)
	}
}

func TestGlobalStart(t *testing.T) {
	ctx, end := othelp.Start(context.Background(), "global-span")
	defer end(nil)
	if ctx == nil {
		t.Error("expected non-nil context")
	}
}
