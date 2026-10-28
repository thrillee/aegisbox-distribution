-- +goose Up
-- +goose StatementBegin

-- =========================================
-- ROUTE SENDER POOL
-- =========================================

-- Some sender IDs are registered with only one upstream route. A route sender pool lists the
-- sender IDs a connection accepts; for service providers that opt in, the gateway replaces the
-- message's sender with one from the pool of the connection it is sent through, so a failover to
-- another route also switches to that route's sender.
CREATE TABLE IF NOT EXISTS route_senders (
    id SERIAL PRIMARY KEY,
    mno_connection_id INT NOT NULL REFERENCES mno_connections(id) ON DELETE CASCADE,
    sender_id_string VARCHAR(16) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    position INT NOT NULL DEFAULT 0,
    usage_count BIGINT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (mno_connection_id, sender_id_string)
);
CREATE INDEX IF NOT EXISTS idx_route_senders_connection_status ON route_senders(mno_connection_id, status);

-- How a connection picks from its pool:
--   round_robin: the sender used least recently, so senders take turns
--   random:      any active sender
--   ordered:     the first active sender by position; the others are only used when it is disabled
ALTER TABLE mno_connections
    ADD COLUMN IF NOT EXISTS sender_pool_strategy VARCHAR(20) NOT NULL DEFAULT 'round_robin';
ALTER TABLE mno_connections DROP CONSTRAINT IF EXISTS mno_connections_sender_pool_strategy_check;
ALTER TABLE mno_connections
    ADD CONSTRAINT mno_connections_sender_pool_strategy_check
    CHECK (sender_pool_strategy IN ('round_robin', 'random', 'ordered'));

-- Opt in per service provider. When on, the provider's own sender is not checked against its
-- approved sender IDs and is replaced by the route's sender on every send.
ALTER TABLE service_providers
    ADD COLUMN IF NOT EXISTS use_route_sender_pool BOOLEAN NOT NULL DEFAULT false;

-- The sender actually submitted upstream when it differs from original_source_addr.
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS sent_source_addr VARCHAR(21);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE messages DROP COLUMN IF EXISTS sent_source_addr;
ALTER TABLE service_providers DROP COLUMN IF EXISTS use_route_sender_pool;
ALTER TABLE mno_connections DROP CONSTRAINT IF EXISTS mno_connections_sender_pool_strategy_check;
ALTER TABLE mno_connections DROP COLUMN IF EXISTS sender_pool_strategy;
DROP TABLE IF EXISTS route_senders;

-- +goose StatementEnd
