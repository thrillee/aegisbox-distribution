-- +goose Up
-- +goose StatementBegin

-- Extensions (Optional)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Service Providers (Your Customers)
CREATE TABLE service_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL, -- For contact/notifications
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- e.g., pending, active, inactive, suspended
    default_currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN', -- Default currency for billing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_service_providers_status ON service_providers(status);

-- SP Credentials Table (Renamed suggestion: sp_credentials)
-- Tracks how SPs connect to YOUR system (SMPP or HTTP)
CREATE TABLE sp_credentials ( -- Renamed from smpp_credentials for clarity
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp', -- 'smpp' or 'http'
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive
    -- SMPP Specific Fields (NULLable if protocol is not 'smpp')
    system_id VARCHAR(16) UNIQUE, -- Unique Constraint applies only when NOT NULL
    password_hash VARCHAR(255),
    bind_type VARCHAR(10) DEFAULT 'trx',
    -- HTTP Specific Fields (NULLable if protocol is not 'http')
    http_config JSONB, -- Store API keys, callback URLs, IP Whitelists etc.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Index checks added during ALTER TABLE later if needed for UNIQUE constraint handling on nullable fields
CREATE INDEX idx_sp_credentials_service_provider_id ON sp_credentials(service_provider_id);
CREATE INDEX idx_sp_credentials_protocol ON sp_credentials(protocol);
CREATE INDEX idx_sp_credentials_system_id ON sp_credentials(system_id) WHERE system_id IS NOT NULL; -- Index non-null system_ids

-- Mobile Network Operators (MNOs) you connect to
CREATE TABLE mnos (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    country_code VARCHAR(5) NOT NULL,
    network_code VARCHAR(10), -- Specific network identifier if needed
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Your Connections to MNOs (SMPP or HTTP)
CREATE TABLE mno_connections (
    id SERIAL PRIMARY KEY,

    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT,

    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp', -- 'smpp' or 'http'
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, disabled, connecting, bound

    -- SMPP Specific Fields
    system_id VARCHAR(16),
    password VARCHAR(255), -- Store plain password as needed by SMPP lib usually
    host VARCHAR(255),
    port INT,
    use_tls BOOLEAN DEFAULT false,
    bind_type VARCHAR(10) DEFAULT 'trx',

    -- Additional SMPP Config
    system_type VARCHAR(13),
    enquire_link_interval_secs INT DEFAULT 30,
    request_timeout_secs INT DEFAULT 10,
    connect_retry_delay_secs INT DEFAULT 5,
    max_window_size INT DEFAULT 10,
    default_data_coding INT DEFAULT 0, -- Default SMSC alphabet
    source_addr_ton INT DEFAULT 1,     -- Default TON/NPI values
    source_addr_npi INT DEFAULT 1,
    dest_addr_ton INT DEFAULT 1,
    dest_addr_npi INT DEFAULT 1,

    -- HTTP Specific Fields
    http_config JSONB, -- Store MNO API URL, auth method, credentials, rate limits etc.

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for optimization
CREATE INDEX idx_mno_connections_mno_id ON mno_connections(mno_id);
CREATE INDEX idx_mno_connections_protocol ON mno_connections(protocol);
CREATE INDEX idx_mno_connections_status ON mno_connections(status);


-- Routing rules based on MSISDN prefix
CREATE TABLE routing_rules (
    id SERIAL PRIMARY KEY,
    prefix VARCHAR(20) UNIQUE NOT NULL, -- e.g., 234803, 23470
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT,
    priority INT NOT NULL DEFAULT 0, -- Lower number means higher priority if prefixes overlap
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Index for efficient prefix matching (use varchar_pattern_ops for LIKE 'prefix%')
CREATE INDEX idx_routing_rules_prefix_pattern ON routing_rules (prefix varchar_pattern_ops);
CREATE INDEX idx_routing_rules_mno_id ON routing_rules(mno_id);

-- Sender IDs allowed for Service Providers
CREATE TABLE sender_ids (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    sender_id_string VARCHAR(16) NOT NULL, -- Alphanumeric (max 11) or Numeric (max 16)
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, sender_id_string)
);
CREATE INDEX idx_sender_ids_service_provider_id ON sender_ids(service_provider_id);
CREATE INDEX idx_sender_ids_status ON sender_ids(status);

-- Message Templates
CREATE TABLE templates (
    id SERIAL PRIMARY KEY,
    service_provider_id INT REFERENCES service_providers(id) ON DELETE CASCADE, -- NULL means global template
    name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, name) -- SP can't have two templates with same name
);
CREATE INDEX idx_templates_service_provider_id ON templates(service_provider_id);

-- Core Message Table (Tracks the logical message submitted by SP)
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    -- SP Identification
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE RESTRICT,
    sp_credential_id INT NOT NULL REFERENCES sp_credentials(id) ON DELETE RESTRICT, -- Renamed FK
    -- Original Submission Details
    client_message_id VARCHAR(64), -- Optional message ID provided by the client
    client_ref VARCHAR(64),        -- Optional reference for multipart correlation (e.g., from UDH)
    original_source_addr VARCHAR(21) NOT NULL, -- Sender ID used by SP
    original_destination_addr VARCHAR(21) NOT NULL, -- Destination submitted by SP
    short_message TEXT NOT NULL, -- Potentially reassembled content for multipart
    total_segments INT NOT NULL DEFAULT 1, -- Number of segments if multipart
    submitted_at TIMESTAMPTZ, -- When SP submitted to us (may differ slightly from received_at)
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- When our system ingested it
    -- Processing & Routing Details
    approved_sender_id INT REFERENCES sender_ids(id) ON DELETE SET NULL,
    template_id INT REFERENCES templates(id) ON DELETE SET NULL,
    routed_mno_id INT REFERENCES mnos(id) ON DELETE SET NULL,
    -- Billing
    currency_code VARCHAR(3) NOT NULL, -- Currency used for pricing/debiting this message
    -- Status Tracking
    processing_status VARCHAR(50) NOT NULL DEFAULT 'received', -- Stage: received -> validated -> routed -> pricing_checked -> queued_for_send -> send_attempted -> (failed states)
    final_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- Aggregate: pending -> delivered / failed / rejected / unknown / expired
    error_code VARCHAR(10),       -- Store internal/protocol error codes for failures (e.g., routing, validation, MNO rejection)
    error_description TEXT,
    -- Timestamps
    processed_for_queue_at TIMESTAMPTZ, -- When picked for sending queue
    completed_at TIMESTAMPTZ -- When message reached a final terminal state (delivered, failed, rejected)
);
-- Indexes
CREATE INDEX idx_messages_service_provider_id ON messages(service_provider_id);
CREATE INDEX idx_messages_received_at ON messages(received_at);
CREATE INDEX idx_messages_processing_status ON messages(processing_status, received_at); -- For workers picking tasks
CREATE INDEX idx_messages_final_status ON messages(final_status, completed_at); -- For reporting
CREATE INDEX idx_messages_destination_msisdn_pattern ON messages(original_destination_addr varchar_pattern_ops); -- Routing uses original dest
CREATE INDEX idx_messages_submitted_at ON messages(submitted_at);


-- Message Segments Table (Tracks each part sent to MNOs)
CREATE TABLE message_segments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE, -- Link to the logical message
    segment_seqn INT NOT NULL, -- Sequence number (1-based)
    mno_message_id VARCHAR(64) UNIQUE, -- MNO ID is unique per segment sent (Index below ensures non-null uniqueness)
    mno_connection_id INT REFERENCES mno_connections(id) ON DELETE SET NULL, -- Which MNO connection was used
    sent_to_mno_at TIMESTAMPTZ,
    dlr_status VARCHAR(50), -- DLR Status for this segment (DELIVRD, EXPIRED, etc.)
    dlr_received_at TIMESTAMPTZ,
    error_code VARCHAR(10), -- Error code from MNO DLR or submission response for this segment
    error_description TEXT, -- Error details for this segment
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (message_id, segment_seqn) -- Ensure segment sequence is unique per message
);
-- Indexes
CREATE INDEX idx_message_segments_message_id ON message_segments(message_id);
CREATE INDEX idx_message_segments_mno_message_id ON message_segments(mno_message_id) WHERE mno_message_id IS NOT NULL; -- Index non-null MNO IDs for DLR lookup
CREATE INDEX idx_message_segments_dlr_status ON message_segments(dlr_status);
CREATE INDEX idx_message_segments_sent_to_mno_at ON message_segments(sent_to_mno_at);


-- Delivery Reports Raw (Optional - uncomment if needed)
CREATE TABLE delivery_reports_raw (
    id BIGSERIAL PRIMARY KEY,
    message_segment_id BIGINT REFERENCES message_segments(id) ON DELETE CASCADE, -- Link to segment if possible
    mno_connection_id INT NOT NULL REFERENCES mno_connections(id) ON DELETE RESTRICT,
    raw_pdu TEXT, -- Or BYTEA
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    processing_status VARCHAR(50) NOT NULL DEFAULT 'pending' -- pending -> processed / failed
);
CREATE INDEX idx_delivery_reports_raw_message_segment_id ON delivery_reports_raw(message_segment_id);
CREATE INDEX idx_delivery_reports_raw_processing_status ON delivery_reports_raw(processing_status, received_at);

-- Wallets (One per Service Provider per Currency)
CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL, -- ISO 4217 code (e.g., NGN, USD, EUR)
    balance NUMERIC(19, 6) NOT NULL DEFAULT 0.000000, -- Increased precision to 6 decimal places
    low_balance_threshold NUMERIC(19, 6) NOT NULL DEFAULT 0.000000, -- Increased precision
    low_balance_notified_at TIMESTAMPTZ, -- Track when the last low balance notification was sent
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, currency_code)
);
CREATE INDEX idx_wallets_service_provider_id ON wallets(service_provider_id);

-- Wallet Transactions (Ledger) - Append-only for auditability
CREATE TABLE wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id INT NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    message_id BIGINT REFERENCES messages(id) ON DELETE SET NULL, -- Link debit/credit to the specific message
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('debit', 'credit', 'reversal')), -- Added reversal type
    amount NUMERIC(19, 6) NOT NULL, -- Use positive amount, type indicates direction. Increased precision.
    balance_before NUMERIC(19, 6) NOT NULL, -- Increased precision
    balance_after NUMERIC(19, 6) NOT NULL, -- Increased precision
    description TEXT,
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reference_transaction_id BIGINT REFERENCES wallet_transactions(id) ON DELETE SET NULL -- Link reversal back to original debit
);
-- Indexes
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_message_id ON wallet_transactions(message_id);
CREATE INDEX idx_wallet_transactions_transaction_date ON wallet_transactions(transaction_date);
CREATE INDEX idx_wallet_transactions_transaction_type ON wallet_transactions(transaction_type);
CREATE INDEX idx_wallet_transactions_reference_transaction_id ON wallet_transactions(reference_transaction_id);


-- Pricing Rules per Service Provider
CREATE TABLE pricing_rules (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    mno_id INT REFERENCES mnos(id) ON DELETE SET NULL, -- If NULL, it's the default price for this SP/currency
    currency_code VARCHAR(3) NOT NULL, -- Should match a wallet currency for the SP
    price_per_sms NUMERIC(10, 6) NOT NULL, -- Price per SMS segment. Increased precision.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Indexes and Named Constraints for Uniqueness
CREATE INDEX idx_pricing_rules_service_provider_id ON pricing_rules(service_provider_id);
CREATE INDEX idx_pricing_rules_mno_id ON pricing_rules(mno_id);
CREATE UNIQUE INDEX idx_unique_default_price_per_currency ON pricing_rules (service_provider_id, currency_code) WHERE mno_id IS NULL;
CREATE UNIQUE INDEX idx_unique_price_per_mno_currency ON pricing_rules (service_provider_id, mno_id, currency_code) WHERE mno_id IS NOT NULL;

-- DLR Forwarding Queue Table
CREATE TABLE dlr_forwarding_queue (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL, -- No FK here to avoid locking messages table, rely on app logic
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, failed, success
    payload JSONB NOT NULL, -- Store sp.ForwardedDLRInfo + sp.SPDetails + SP ID etc.
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5, -- Example max attempts
    last_attempt_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Fields for SKIP LOCKED pattern
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(100) -- Identifier of the worker instance that locked it
);
CREATE INDEX idx_dlr_forwarding_queue_status_created ON dlr_forwarding_queue(status, created_at);
CREATE INDEX idx_dlr_forwarding_queue_locked ON dlr_forwarding_queue(locked_at); -- Index for finding stale locks

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pricing_rules;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallets;
-- DROP TABLE IF EXISTS delivery_reports_raw; -- Uncomment if used
DROP TABLE IF EXISTS message_segments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS sender_ids;
DROP TABLE IF EXISTS routing_rules;
DROP TABLE IF EXISTS mno_connections;
DROP TABLE IF EXISTS mnos;
DROP TABLE IF EXISTS sp_credentials; -- Use the new name
DROP TABLE IF EXISTS service_providers;
DROP TABLE IF EXISTS dlr_forwarding_queue;

-- +goose StatementEnd
