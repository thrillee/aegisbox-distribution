-- +goose Up
-- +goose StatementBegin

-- For generating UUIDs if needed, although SERIALs are often fine for PKs
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Service Providers (Your Customers)
CREATE TABLE service_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_service_providers_status ON service_providers(status);

-- SMPP Credentials for Service Providers connecting to your system
CREATE TABLE smpp_credentials (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    system_id VARCHAR(16) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    bind_type VARCHAR(10) NOT NULL DEFAULT 'trx',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_smpp_credentials_service_provider_id ON smpp_credentials(service_provider_id);

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

-- SMPP Connections to MNOs
CREATE TABLE mno_connections (
    id SERIAL PRIMARY KEY,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT,
    system_id VARCHAR(16) NOT NULL,
    password VARCHAR(255) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    use_tls BOOLEAN NOT NULL DEFAULT false,
    bind_type VARCHAR(10) NOT NULL DEFAULT 'trx',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_mno_connections_mno_id ON mno_connections(mno_id);
CREATE INDEX idx_mno_connections_status ON mno_connections(status);

-- Routing rules
CREATE TABLE routing_rules (
    id SERIAL PRIMARY KEY,
    prefix VARCHAR(20) UNIQUE NOT NULL,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT,
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_routing_rules_prefix_pattern ON routing_rules (prefix varchar_pattern_ops);
CREATE INDEX idx_routing_rules_mno_id ON routing_rules(mno_id);

-- Sender IDs
CREATE TABLE sender_ids (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    sender_id_string VARCHAR(16) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, sender_id_string)
);
CREATE INDEX idx_sender_ids_service_provider_id ON sender_ids(service_provider_id);
CREATE INDEX idx_sender_ids_status ON sender_ids(status);

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

-- Messages
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE RESTRICT,
    smpp_credential_id INT NOT NULL REFERENCES smpp_credentials(id) ON DELETE RESTRICT,
    client_message_id VARCHAR(64),
    sender_id_used VARCHAR(16) NOT NULL,
    destination_msisdn VARCHAR(20) NOT NULL,
    short_message TEXT NOT NULL,
    template_id INT REFERENCES templates(id) ON DELETE SET NULL,
    approved_sender_id INT REFERENCES sender_ids(id) ON DELETE SET NULL,
    routed_mno_id INT REFERENCES mnos(id) ON DELETE SET NULL,
    -- Updated design
    currency_code VARCHAR(3),
    total_segments INT NOT NULL DEFAULT 1,
    client_ref VARCHAR(64),
    submitted_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_for_queue_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    processing_status VARCHAR(50) NOT NULL DEFAULT 'received',
    send_status VARCHAR(50),
    final_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_code INT,
    error_description TEXT
);
CREATE INDEX idx_messages_service_provider_id ON messages(service_provider_id);
CREATE INDEX idx_messages_received_at ON messages(received_at);
CREATE INDEX idx_messages_processing_status ON messages(processing_status);
CREATE INDEX idx_messages_final_status ON messages(final_status);
CREATE INDEX idx_messages_destination_msisdn_pattern ON messages(destination_msisdn varchar_pattern_ops);

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
    error_code INT,
    error_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (message_id, segment_seqn)
);
CREATE INDEX idx_message_segments_message_id ON message_segments(message_id);
CREATE INDEX idx_message_segments_mno_message_id ON message_segments(mno_message_id);
CREATE INDEX idx_message_segments_dlr_status ON message_segments(dlr_status);

-- Delivery Reports Raw
CREATE TABLE delivery_reports_raw (
     id BIGSERIAL PRIMARY KEY,
     message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
     mno_connection_id INT NOT NULL REFERENCES mno_connections(id) ON DELETE RESTRICT,
     raw_pdu TEXT,
     received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     processed_at TIMESTAMPTZ,
     processing_status VARCHAR(50) NOT NULL DEFAULT 'pending'
);
CREATE INDEX idx_delivery_reports_raw_message_id ON delivery_reports_raw(message_id);
CREATE INDEX idx_delivery_reports_raw_processing_status ON delivery_reports_raw(processing_status, received_at);

-- Wallets
CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL,
    balance DECIMAL(19, 4) NOT NULL DEFAULT 0.00,
    low_balance_threshold DECIMAL(19, 4) NOT NULL DEFAULT 0.00,
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
    transaction_type VARCHAR(10) NOT NULL,
    amount DECIMAL(19, 4) NOT NULL,
    balance_before DECIMAL(19, 4) NOT NULL,
    balance_after DECIMAL(19, 4) NOT NULL,
    description TEXT,
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
    price_per_sms DECIMAL(10, 6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_pricing_rules_service_provider_id ON pricing_rules(service_provider_id);
CREATE INDEX idx_pricing_rules_mno_id ON pricing_rules(mno_id);
CREATE UNIQUE INDEX unique_default_price_per_currency ON pricing_rules (service_provider_id, currency_code) WHERE mno_id IS NULL;
CREATE UNIQUE INDEX unique_price_per_mno_currency ON pricing_rules (service_provider_id, mno_id, currency_code) WHERE mno_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pricing_rules;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS delivery_reports_raw;
DROP TABLE IF EXISTS message_segments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS sender_ids;
DROP TABLE IF EXISTS routing_rules;
DROP TABLE IF EXISTS mno_connections;
DROP TABLE IF EXISTS mnos;
DROP TABLE IF EXISTS smpp_credentials;
DROP TABLE IF EXISTS service_providers;

-- +goose StatementEnd

