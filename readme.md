# Transactional Outbox RabbitMQ

This repository accompanies **Article 1** of a series on the **Transactional Outbox pattern**, demonstrating how to reliably publish domain events from a database-backed service using **Go**, **PostgreSQL**, **RabbitMQ**, and **Docker Compose**.

The focus is on **correctness and observability**, not frameworks or abstractions.

**Read the article:**
_Transactional Outbox with RabbitMQ: Building Reliable Event Publishing in Microservices_ (coming soon)

## What This Repository Demonstrates

- Transactional outbox pattern with PostgreSQL
- Reliable event publishing without dual writes
- Worker-based outbox polling and leasing
- Consumer-side idempotency
- Minimal but high-signal Prometheus metrics
- End-to-end trace propagation across async boundaries

This is **not** a production-ready framework, it’s a **reference implementation** meant to explain _why_ things are done a certain way.

## Architecture Overview

The system consists of two services and a shared broker:

- **Order Service**

  - Handles HTTP requests
  - Writes domain data and outbox events atomically
  - Publishes events asynchronously via an outbox worker

- **Notification Service**

  - Consumes events from RabbitMQ
  - Ensures idempotent processing using a local table

**Each service owns its own database schema. There is no shared database.**

![Outbox Highlevel Diagram](image.png)

## Running Locally with Docker Compose

### Prerequisites

- [Docker](https://docs.docker.com/engine/install)

### Start all services

```bash
docker compose up --build
```

This starts:

- Order Service
- Notification Service
- Two isolated PostgreSQL databases
- RabbitMQ
- Prometheus
- Grafana
- Jaeger

All services wait for their dependencies via health checks before starting.

### Create an order

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"product_id":"sku123","quantity":1}' \
  http://localhost:4000/orders | jq
```

Example response:

```json
{
  "order": {
    "id": 1,
    "status": "pending",
    "created_at": "2026-01-10T15:57:41.309777771Z"
  }
}
```

### Expected logs

You should see the request flow across the entire system:

```bash
# HTTP request arrives
order-service         | {"level":"info","message":"Create Order Request Arrived"}

# Outbox worker picks up the event
order-service         | {"level":"info","count":1,"message":"Fetched outbox events"}
order-service         | {"level":"info","worker_id":2,"event_id":"a6f6d2df-f2f9-4a0f-98b7-73c99e96f75b","event_key":"order.created","message":"Worker processing event"}
order-service         | {"level":"info","routing_key":"order.created","message":"RabbitMQ message published"}

# Notification service consumes the message
notification-service  | {"level":"info","message":"Broker Message Arrived"}
notification-service  | {"level":"info","payload":{"id":5,"product_id":"sku123","quantity":1},"message":"Order email sent to customer"}
```

This confirms the full path:

**HTTP → DB → Outbox → Broker → Consumer**

## Observability

### Metrics (Prometheus + Grafana)

The Order Service exposes a minimal set of **high-signal outbox metrics**, including:

- Outbox backlog
- Event publish success / failure counts
- End-to-end publish latency

**Grafana dashboard screenshot (from the article):**

![Outbox Grafana Dashboard](./assets/outbox-grafana-dashboard-article-1.png)

Grafana is available at:

```
http://localhost:3000
```

(username/password: admin)

### Tracing (Jaeger)

Distributed tracing is wired end-to-end:

- HTTP request span
- Outbox event creation
- Asynchronous publish span
- RabbitMQ consume span

**Jaeger trace example (from the article):**

![Create Order Jaeger Trace](./assets/create-order-jaeger-trace.png)

Jaeger UI:

```
http://localhost:16686
```

## Repository Structure

```bash
.
├── order-service
│   ├── cmd
│   └── internal
│       ├── database
│       ├── outbox
│       ├── rabbitmq
│       └── service
├── notification-service
│   ├── cmd
│   └── internal
│       ├── database
│       ├── rabbitmq
│       └── service
├── docker-compose.yaml
└── README.md
```

Each service is intentionally structured similarly to reduce cognitive load.

## What’s Covered in Article 1

- Atomic writes using the outbox pattern
- Worker-based publishing model
- Failure handling (success vs failed states)
- Consumer-side idempotency
- Minimal metrics and tracing

## What’s Coming Next (Article 2)

- Retrying outbox publishing and designing Dead Letter Queues (DLQs)
- Consumer-side retries using dead-letter exchanges

## Support & Contributions

If you find this repository useful, consider giving it a ⭐ — it helps others discover it.

Feedback, corrections, and discussions are very welcome.
Feel free to open an issue or submit a PR.
