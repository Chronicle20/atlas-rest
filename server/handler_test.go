package server_test

import (
	"context"
	"errors"
	"github.com/Chronicle20/atlas-rest/requests"
	"github.com/Chronicle20/atlas-rest/server"
	tenant "github.com/Chronicle20/atlas-tenant"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockSpan struct {
	trace.Span
	spanContext trace.SpanContext
}

func (ms *MockSpan) SpanContext() trace.SpanContext {
	return ms.spanContext
}

func (ms *MockSpan) IsRecording() bool {
	return true
}

func (ms *MockSpan) End(options ...trace.SpanEndOption) {
}

func (ms *MockSpan) RecordError(err error, options ...trace.EventOption) {
	// You can record the error or count calls here
}

type MockTracer struct {
	trace.Tracer
	StartedSpans []*MockSpan
}

func (mt *MockTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SpanID:     trace.SpanID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		TraceFlags: trace.FlagsSampled,
	})
	mockSpan := &MockSpan{spanContext: spanContext}
	return trace.ContextWithSpan(ctx, mockSpan), mockSpan
}

type MockTracerProvider struct {
	trace.TracerProvider
	tracer *MockTracer
}

func (m MockTracerProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	if m.tracer == nil {
		m.tracer = &MockTracer{}
	}
	return m.tracer
}

func TestSpanPropagation(t *testing.T) {
	l, _ := test.NewNullLogger()

	otel.SetTracerProvider(&MockTracerProvider{})
	otel.SetTextMapPropagator(propagation.TraceContext{})

	ictx, ispan := otel.GetTracerProvider().Tracer("atlas-kafka").Start(context.Background(), "test-span")

	req, err := http.NewRequest(http.MethodGet, "www.google.com", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	w := httptest.NewRecorder()

	requests.SpanHeaderDecorator(ictx)(req.Header)

	server.RetrieveSpan(l, "test-handler", func(l logrus.FieldLogger, ctx context.Context) http.HandlerFunc {
		span := trace.SpanFromContext(ctx)
		if !span.SpanContext().TraceID().IsValid() {
			t.Fatalf(errors.New("invalid trace id").Error())
		}
		if span.SpanContext().TraceID() != ispan.SpanContext().TraceID() {
			t.Fatalf(errors.New("invalid trace id").Error())
		}
		return func(w http.ResponseWriter, r *http.Request) {
		}
	})(w, req)
}

func TestTenantPropagation(t *testing.T) {
	l, _ := test.NewNullLogger()
	it, err := tenant.Create(uuid.New(), "GMS", 83, 1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	ictx := tenant.WithContext(context.Background(), it)

	req, err := http.NewRequest(http.MethodGet, "www.google.com", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	w := httptest.NewRecorder()

	requests.TenantHeaderDecorator(ictx)(req.Header)

	server.ParseTenant(l, func(l logrus.FieldLogger, ot tenant.Model) http.HandlerFunc {
		if !it.Is(ot) {
			t.Fatalf(errors.New("invalid tenant").Error())
		}
		return func(w http.ResponseWriter, r *http.Request) {
		}
	})(w, req)
}