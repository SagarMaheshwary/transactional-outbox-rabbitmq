package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func headersWithTraceContext(ctx context.Context) amqp091.Table {
	headers := amqp091.Table{}
	carrier := propagation.MapCarrier{}

	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		headers[k] = v
	}

	return headers
}
