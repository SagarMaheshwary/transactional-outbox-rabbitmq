package tracing

import (
	"context"

	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

type Opts struct {
	Config *config.Tracing
	Logger logger.Logger
}

type TracerService struct {
	Config *config.Tracing
	Logger logger.Logger
	tp     *sdktrace.TracerProvider
}

func NewTracerService(ctx context.Context, opts *Opts) (*TracerService, error) {
	cfg := opts.Config
	Tracer = otel.Tracer(cfg.ServiceName)

	exporter, err := otlptrace.New(
		ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(cfg.CollectorURL),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	opts.Logger.Info("Tracing initialized (exporter: otlptracehttp)",
		logger.Field{Key: "serviceName", Value: cfg.ServiceName},
	)

	return &TracerService{
		Config: cfg,
		Logger: opts.Logger,
		tp:     tp,
	}, nil
}

func (t *TracerService) Shutdown(ctx context.Context) error {
	return t.tp.Shutdown(ctx)
}
