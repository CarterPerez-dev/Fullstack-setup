// AngelaMos | 2026
// telemetry.go

package core

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/carterperez-dev/templates/go-backend/internal/config"
)

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	Tracer         trace.Tracer
}

func NewTelemetry(
	ctx context.Context,
	otelCfg config.OtelConfig,
	appCfg config.AppConfig,
) (*Telemetry, error) {
	if !otelCfg.Enabled || otelCfg.Endpoint == "" {
		noopProvider := sdktrace.NewTracerProvider()
		return &Telemetry{
			TracerProvider: noopProvider,
			Tracer:         noopProvider.Tracer(otelCfg.ServiceName),
		}, nil
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(otelCfg.Endpoint),
		otlptracegrpc.WithTimeout(5 * time.Second),
	}

	if otelCfg.Insecure {
		opts = append(
			opts,
			otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
		)
	} else {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(otelCfg.ServiceName),
			semconv.ServiceVersion(appCfg.Version),
			attribute.String("environment", appCfg.Environment),
		),
		resource.WithHost(),
		resource.WithProcess(),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	sampleRate := otelCfg.SampleRate
	if sampleRate <= 0 || sampleRate > 1 {
		sampleRate = 0.1
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(sampleRate),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Telemetry{
		TracerProvider: tp,
		Tracer:         tp.Tracer(otelCfg.ServiceName),
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.TracerProvider == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := t.TracerProvider.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown tracer provider: %w", err)
	}

	return nil
}

func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func AddSpanEvent(
	ctx context.Context,
	name string,
	attrs ...attribute.KeyValue,
) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
