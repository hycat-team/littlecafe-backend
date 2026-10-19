-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE IF NOT EXISTS outbox_records (
    id UUID PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox_records (status, created_at);

CREATE TABLE IF NOT EXISTS inbox_messages (
    consumer VARCHAR(255) NOT NULL,
    message_id UUID NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    PRIMARY KEY (consumer, message_id)
);

CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY,
    actor_id UUID,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255) NOT NULL,
    resource_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_log_occurred_at ON audit_log (occurred_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS inbox_messages;
DROP TABLE IF EXISTS outbox_records;

