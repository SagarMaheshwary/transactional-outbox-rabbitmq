CREATE TABLE
  processed_messages (
    message_id TEXT PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT now ()
  );