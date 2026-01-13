CREATE TABLE
  orders (
    id BIGSERIAL PRIMARY KEY,
    status VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT now ()
  );

CREATE TYPE OutboxEventStatus as ENUM ('pending', 'in_progress', 'published', 'failed');

CREATE TABLE
  outbox_events (
    id TEXT PRIMARY KEY,
    event_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    status OutboxEventStatus NOT NULL,
    retry_count INT DEFAULT 0,
    next_retry_at TIMESTAMP DEFAULT NULL,
    locked_at TIMESTAMP DEFAULT NULL,
    locked_by VARCHAR(128) NULL,
    failure_reason VARCHAR(128) DEFAULT NULL,
    failed_at TIMESTAMP DEFAULT NULL,
    traceparent TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW ()
  );

CREATE INDEX idx_outbox_events_pending_ready ON outbox_events (created_at)
WHERE
  status = 'pending';

CREATE INDEX idx_outbox_events_retryable ON outbox_events (locked_at, next_retry_at, created_at)
WHERE
  status = 'in_progress';