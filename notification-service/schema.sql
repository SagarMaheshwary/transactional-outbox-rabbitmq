CREATE TABLE
  processed_messages (
    message_id UUID PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT now ()
  );