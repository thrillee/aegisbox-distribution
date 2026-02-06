-- +goose Up
-- +goose StatementBegin

-- Consolidated Aegisbox Schema with HTTP MNO Support
-- Combines original schema with talk-thrill-http providers and senders

-- Extensions (Optional)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Service Providers (Your Customers)
CREATE TABLE service_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    default_currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_service_providers_status ON service_providers(status);

-- SP Credentials Table
-- Tracks how SPs connect to YOUR system (SMPP or HTTP)
CREATE TABLE sp_credentials (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    -- SMPP Specific Fields
    system_id VARCHAR(16) UNIQUE,
    password_hash VARCHAR(255),
    bind_type VARCHAR(10) DEFAULT 'trx',
    -- HTTP Specific Fields
    api_key_hash VARCHAR(255),
    api_key_identifier VARCHAR(64) UNIQUE,
    http_config JSONB,
    -- Flexible routing
    scope VARCHAR(50) NOT NULL DEFAULT 'default',
    routing_group_id INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sp_credentials_service_provider_id ON sp_credentials(service_provider_id);
CREATE INDEX idx_sp_credentials_protocol ON sp_credentials(protocol);
CREATE INDEX idx_sp_credentials_system_id ON sp_credentials(system_id) WHERE system_id IS NOT NULL;
CREATE INDEX idx_sp_credentials_scope ON sp_credentials(scope);
CREATE INDEX idx_sp_credentials_routing_group_id ON sp_credentials(routing_group_id);

-- Mobile Network Operators (MNOs)
CREATE TABLE mnos (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    country_code VARCHAR(5) NOT NULL,
    network_code VARCHAR(10),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- MNO Connections (SMPP or HTTP)
CREATE TABLE mno_connections (
    id SERIAL PRIMARY KEY,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT,
    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    priority INT DEFAULT 1,

    -- SMPP Specific Fields
    system_id VARCHAR(16),
    password VARCHAR(255),
    host VARCHAR(255),
    port INT,
    use_tls BOOLEAN DEFAULT false,
    bind_type VARCHAR(10) DEFAULT 'trx',
    system_type VARCHAR(13),
    enquire_link_interval_secs INT DEFAULT 30,
    request_timeout_secs INT DEFAULT 10,
    connect_retry_delay_secs INT DEFAULT 5,
    max_window_size INT DEFAULT 10,
    default_data_coding INT DEFAULT 0,
    source_addr_ton INT DEFAULT 1,
    source_addr_npi INT DEFAULT 1,
    dest_addr_ton INT DEFAULT 1,
    dest_addr_npi INT DEFAULT 1,

    -- HTTP Specific Fields (from talk-thrill-http providers table)
    api_key TEXT,
    base_url VARCHAR(255) NOT NULL DEFAULT '',
    username VARCHAR(255),
    http_password TEXT,
    secret_key TEXT,
    default_sender VARCHAR(100),
    rate_limit INTEGER DEFAULT 100,
    email VARCHAR(255),
    supports_webhook BOOLEAN DEFAULT true,
    webhook_path VARCHAR(255),
    timeout_secs INT DEFAULT 30,

    -- HTTP Provider Configuration (JSONB for provider-specific settings)
    provider_config JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_mno_connections_mno_id ON mno_connections(mno_id);
CREATE INDEX idx_mno_connections_protocol ON mno_connections(protocol);
CREATE INDEX idx_mno_connections_status ON mno_connections(status);

-- Sender IDs (from talk-thrill-http senders table merged with aegisbox sender_ids)
CREATE TABLE sender_ids (
    id SERIAL PRIMARY KEY,
    service_provider_id INT REFERENCES service_providers(id) ON DELETE CASCADE,
    sender_id_string VARCHAR(16) NOT NULL,
    mno_connection_id INT REFERENCES mno_connections(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- HTTP sender-specific fields (from talk-thrill-http senders)
    provider_id INT REFERENCES mno_connections(id) ON DELETE SET NULL,
    rate_limit INTEGER DEFAULT 100,
    otp_template TEXT,
    networks TEXT[] DEFAULT '{}',
    -- OTP sender fields
    max_usage_count INT DEFAULT 10000,
    current_usage_count INT DEFAULT 0,
    reset_interval_hours INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, sender_id_string)
);
CREATE INDEX idx_sender_ids_service_provider_id ON sender_ids(service_provider_id);
CREATE INDEX idx_sender_ids_status ON sender_ids(status);
CREATE INDEX idx_sender_ids_provider_id ON sender_ids(provider_id);
CREATE INDEX idx_sender_ids_mno_connection_id ON sender_ids(mno_connection_id);

-- MSISDN Prefix Groups (for flexible routing)
CREATE TABLE msisdn_prefix_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    reference VARCHAR(100) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE msisdn_prefix_groups IS 'Groups of MSISDN prefixes, e.g., "MTN Nigeria Prefixes".';

CREATE TABLE msisdn_prefix_entries (
    id SERIAL PRIMARY KEY,
    msisdn_prefix_group_id INT NOT NULL REFERENCES msisdn_prefix_groups(id) ON DELETE CASCADE,
    msisdn_prefix VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (msisdn_prefix_group_id, msisdn_prefix),
    UNIQUE (msisdn_prefix)
);
COMMENT ON TABLE msisdn_prefix_entries IS 'Individual MSISDN prefixes belonging to a specific msisdn_prefix_group.';
CREATE INDEX idx_msisdn_prefix_entries_prefix_pattern ON msisdn_prefix_entries (msisdn_prefix varchar_pattern_ops);
CREATE INDEX idx_msisdn_prefix_entries_group_id ON msisdn_prefix_entries(msisdn_prefix_group_id);

-- Routing Groups
CREATE TABLE routing_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    reference VARCHAR(100) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE routing_groups IS 'Defines a collection of routing rules forming a specific routing strategy.';

CREATE TABLE routing_assignments (
    id SERIAL PRIMARY KEY,
    routing_group_id INT NOT NULL REFERENCES routing_groups(id) ON DELETE CASCADE,
    msisdn_prefix_group_id INT NOT NULL REFERENCES msisdn_prefix_groups(id) ON DELETE RESTRICT,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    priority INT NOT NULL DEFAULT 0,
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (routing_group_id, msisdn_prefix_group_id)
);
COMMENT ON TABLE routing_assignments IS 'Assigns an MNO to a specific MSISDN prefix group within a given routing group.';
CREATE INDEX idx_routing_assignments_routing_group_id ON routing_assignments(routing_group_id);
CREATE INDEX idx_routing_assignments_msisdn_prefix_group_id ON routing_assignments(msisdn_prefix_group_id);
CREATE INDEX idx_routing_assignments_mno_id ON routing_assignments(mno_id);
CREATE INDEX idx_routing_assignments_status ON routing_assignments(status);

-- Message Templates
CREATE TABLE templates (
    id SERIAL PRIMARY KEY,
    service_provider_id INT REFERENCES service_providers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, name)
);
CREATE INDEX idx_templates_service_provider_id ON templates(service_provider_id);

-- OTP Alternative Senders
CREATE TABLE otp_alternative_senders (
    id SERIAL PRIMARY KEY,
    sender_id_string VARCHAR(16) NOT NULL UNIQUE,
    service_provider_id INT REFERENCES service_providers(id) ON DELETE CASCADE,
    mno_id INT REFERENCES mnos(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_usage_count INT NOT NULL DEFAULT 0,
    max_usage_count INT NOT NULL DEFAULT 10000,
    reset_interval_hours INT DEFAULT NULL,
    last_reset_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_otp_alternative_senders_mno_id ON otp_alternative_senders(mno_id);
CREATE INDEX idx_otp_alternative_senders_status_usage ON otp_alternative_senders(status, current_usage_count);
CREATE INDEX idx_otp_alternative_senders_last_used ON otp_alternative_senders(last_used_at);

-- OTP Message Templates
CREATE TABLE otp_message_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    content_template TEXT NOT NULL,
    default_brand_name VARCHAR(100) DEFAULT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_otp_message_templates_status ON otp_message_templates(status);

-- OTP Sender Template Assignments
CREATE TABLE otp_sender_template_assignments (
    id SERIAL PRIMARY KEY,
    otp_alternative_sender_id INT NOT NULL REFERENCES otp_alternative_senders(id) ON DELETE CASCADE,
    otp_message_template_id INT NOT NULL REFERENCES otp_message_templates(id) ON DELETE CASCADE,
    mno_id INT REFERENCES mnos(id) ON DELETE SET NULL,
    priority INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    assignment_usage_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (otp_alternative_sender_id, otp_message_template_id, mno_id)
);
CREATE INDEX idx_otp_assignments_sender_template ON otp_sender_template_assignments(otp_alternative_sender_id, otp_message_template_id);
CREATE INDEX idx_otp_assignments_mno_status_priority ON otp_sender_template_assignments(mno_id, status, priority);

-- Core Message Table
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE RESTRICT,
    sp_credential_id INT NOT NULL REFERENCES sp_credentials(id) ON DELETE RESTRICT,
    client_message_id VARCHAR(64),
    client_ref VARCHAR(64),
    original_source_addr VARCHAR(21) NOT NULL,
    original_destination_addr VARCHAR(21) NOT NULL,
    short_message TEXT NOT NULL,
    total_segments INT NOT NULL DEFAULT 1,
    submitted_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_sender_id INT REFERENCES sender_ids(id) ON DELETE SET NULL,
    template_id INT REFERENCES templates(id) ON DELETE SET NULL,
    routed_mno_id INT REFERENCES mnos(id) ON DELETE SET NULL,
    currency_code VARCHAR(3) NOT NULL,
    cost NUMERIC(19, 6) DEFAULT 0.000000,
    processing_status VARCHAR(50) NOT NULL DEFAULT 'received',
    final_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_code VARCHAR(10),
    error_description TEXT,
    processed_for_queue_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    -- Retry tracking
    retry_count INTEGER DEFAULT 0 NOT NULL,
    next_retry_at TIMESTAMPTZ
);
CREATE INDEX idx_messages_service_provider_id ON messages(service_provider_id);
CREATE INDEX idx_messages_received_at ON messages(received_at);
CREATE INDEX idx_messages_processing_status ON messages(processing_status, received_at);
CREATE INDEX idx_messages_final_status ON messages(final_status, completed_at);
CREATE INDEX idx_messages_destination_msisdn_pattern ON messages(original_destination_addr varchar_pattern_ops);
CREATE INDEX idx_messages_submitted_at ON messages(submitted_at);

-- Message Segments
CREATE TABLE message_segments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    segment_seqn INT NOT NULL,
    mno_message_id VARCHAR(64) UNIQUE,
    mno_connection_id INT REFERENCES mno_connections(id) ON DELETE SET NULL,
    sent_to_mno_at TIMESTAMPTZ,
    dlr_status VARCHAR(50),
    dlr_received_at TIMESTAMPTZ,
    error_code VARCHAR(10),
    error_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (message_id, segment_seqn)
);
CREATE INDEX idx_message_segments_message_id ON message_segments(message_id);
CREATE INDEX idx_message_segments_mno_message_id ON message_segments(mno_message_id) WHERE mno_message_id IS NOT NULL;
CREATE INDEX idx_message_segments_dlr_status ON message_segments(dlr_status);
CREATE INDEX idx_message_segments_sent_to_mno_at ON message_segments(sent_to_mno_at);

-- Delivery Reports Raw
CREATE TABLE delivery_reports_raw (
    id BIGSERIAL PRIMARY KEY,
    message_segment_id BIGINT REFERENCES message_segments(id) ON DELETE CASCADE,
    mno_connection_id INT NOT NULL REFERENCES mno_connections(id) ON DELETE RESTRICT,
    raw_payload TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    processing_status VARCHAR(50) NOT NULL DEFAULT 'pending'
);
CREATE INDEX idx_delivery_reports_raw_message_segment_id ON delivery_reports_raw(message_segment_id);
CREATE INDEX idx_delivery_reports_raw_processing_status ON delivery_reports_raw(processing_status, received_at);

-- Wallets
CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL,
    balance NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    low_balance_threshold NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    low_balance_notified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, currency_code)
);
CREATE INDEX idx_wallets_service_provider_id ON wallets(service_provider_id);

-- Wallet Transactions
CREATE TABLE wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id INT NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    message_id BIGINT REFERENCES messages(id) ON DELETE SET NULL,
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('debit', 'credit', 'reversal')),
    amount NUMERIC(19, 6) NOT NULL,
    balance_before NUMERIC(19, 6) NOT NULL,
    balance_after NUMERIC(19, 6) NOT NULL,
    description TEXT,
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reference_transaction_id BIGINT REFERENCES wallet_transactions(id) ON DELETE SET NULL
);
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_message_id ON wallet_transactions(message_id);
CREATE INDEX idx_wallet_transactions_transaction_date ON wallet_transactions(transaction_date);
CREATE INDEX idx_wallet_transactions_transaction_type ON wallet_transactions(transaction_type);

-- Pricing Rules
CREATE TABLE pricing_rules (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    mno_id INT REFERENCES mnos(id) ON DELETE SET NULL,
    currency_code VARCHAR(3) NOT NULL,
    price_per_sms NUMERIC(10, 6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_pricing_rules_service_provider_id ON pricing_rules(service_provider_id);
CREATE INDEX idx_pricing_rules_mno_id ON pricing_rules(mno_id);
CREATE UNIQUE INDEX idx_unique_default_price_per_currency ON pricing_rules (service_provider_id, currency_code) WHERE mno_id IS NULL;
CREATE UNIQUE INDEX idx_unique_price_per_mno_currency ON pricing_rules (service_provider_id, mno_id, currency_code) WHERE mno_id IS NOT NULL;

-- DLR Forwarding Queue
CREATE TABLE dlr_forwarding_queue (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payload JSONB NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    last_attempt_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(100)
);
CREATE INDEX idx_dlr_forwarding_queue_status_created ON dlr_forwarding_queue(status, created_at);
CREATE INDEX idx_dlr_forwarding_queue_locked ON dlr_forwarding_queue(locked_at);
CREATE INDEX idx_dlr_forward_queue_next_retry ON dlr_forwarding_queue(next_retry_at) WHERE status = 'pending';

-- Dead Letter Queue
CREATE TABLE dead_letter_queue (
    id SERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    original_status VARCHAR(50),
    error_code VARCHAR(50),
    error_description TEXT,
    retry_count INTEGER DEFAULT 0,
    failed_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB
);
CREATE INDEX idx_dead_letter_queue_message_id ON dead_letter_queue(message_id);
CREATE INDEX idx_dead_letter_queue_failed_at ON dead_letter_queue(failed_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS dead_letter_queue;
DROP TABLE IF EXISTS dlr_forwarding_queue;
DROP TABLE IF EXISTS pricing_rules;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS delivery_reports_raw;
DROP TABLE IF EXISTS message_segments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS otp_sender_template_assignments;
DROP TABLE IF EXISTS otp_message_templates;
DROP TABLE IF EXISTS otp_alternative_senders;
DROP TABLE IF EXISTS sender_ids;
DROP TABLE IF EXISTS routing_assignments;
DROP TABLE IF EXISTS routing_groups;
DROP TABLE IF EXISTS msisdn_prefix_entries;
DROP TABLE IF EXISTS msisdn_prefix_groups;
DROP TABLE IF EXISTS mno_connections;
DROP TABLE IF EXISTS mnos;
DROP TABLE IF EXISTS sp_credentials;
DROP TABLE IF EXISTS service_providers;

-- +goose StatementEnd
