package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func ExtractTraceParent(
	ctx context.Context,
	traceparent string,
) context.Context {
	carrier := propagation.MapCarrier{
		"traceparent": traceparent,
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
