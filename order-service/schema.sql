CREATE TABLE
  IF NOT EXISTS orders (
    id bigserial primary key,
    status varchar(10) not null,
    created_at timestamp default now (),
    updated_at timestamp default null
  );

CREATE TABLE
  IF NOT EXISTS outbox_events (
    id text primary key,
    event_key text not null,
    payload jsonb not null,
    status text not null,
    locked_at timestamp default null,
    locked_by varchar(128) null,
    failure_reason varchar(128) default null,
    failed_at timestamp default null,
    created_at timestamp default now ()
  );