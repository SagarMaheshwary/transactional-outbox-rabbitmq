package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func contextWithOtelHeaders(ctx context.Context, headers amqp091.Table) context.Context {
	carrier := make(propagation.MapCarrier)
	for k, v := range headers {
		if str, ok := v.(string); ok {
			carrier[k] = str
		}
	}

	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
