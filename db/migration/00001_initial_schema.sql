-- +goose Up
-- +goose StatementBegin

-- For generating UUIDs if needed, although SERIALs are often fine for PKs
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Service Providers (Your Customers)
CREATE TABLE service_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL, -- For contact/notifications
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- e.g., pending, active, inactive, suspended
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_service_providers_status ON service_providers(status);

-- SMPP Credentials for Service Providers connecting to your system
CREATE TABLE smpp_credentials (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    system_id VARCHAR(16) UNIQUE NOT NULL, -- SMPP system_id
    password_hash VARCHAR(255) NOT NULL, -- Store hashed passwords
    bind_type VARCHAR(10) NOT NULL DEFAULT 'trx', -- tx, rx, trx
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_smpp_credentials_service_provider_id ON smpp_credentials(service_provider_id);

-- Mobile Network Operators (MNOs) you connect to
CREATE TABLE mnos (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    country_code VARCHAR(5) NOT NULL, -- e.g., 234
    network_code VARCHAR(10), -- Specific network identifier if needed
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Your SMPP Client Connections to MNOs
CREATE TABLE mno_connections (
    id SERIAL PRIMARY KEY,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT,
    system_id VARCHAR(16) NOT NULL,
    password VARCHAR(255) NOT NULL, -- Store plain password as needed by SMPP lib usually
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    use_tls BOOLEAN NOT NULL DEFAULT false,
    bind_type VARCHAR(10) NOT NULL DEFAULT 'trx', -- tx, rx, trx
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, disabled
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_mno_connections_mno_id ON mno_connections(mno_id);
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

-- Core Message Table (Incoming from SPs -> Outgoing to MNOs)
-- This table will be VERY high volume. Consider partitioning later if needed.
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    -- Who sent it?
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE RESTRICT,
    smpp_credential_id INT NOT NULL REFERENCES smpp_credentials(id) ON DELETE RESTRICT,
    -- Message Details
    client_message_id VARCHAR(64), -- Optional message ID provided by the client
    sender_id_used VARCHAR(16) NOT NULL,
    destination_msisdn VARCHAR(20) NOT NULL,
    short_message TEXT NOT NULL,
    -- Optional validation links
    template_id INT REFERENCES templates(id) ON DELETE SET NULL,
    approved_sender_id INT REFERENCES sender_ids(id) ON DELETE SET NULL, -- Link to the validated SenderID record
    -- Routing & MNO Info
    routed_mno_id INT REFERENCES mnos(id) ON DELETE SET NULL,
    mno_connection_id INT REFERENCES mno_connections(id) ON DELETE SET NULL,
    mno_message_id VARCHAR(64), -- Message ID received from the MNO after submit_sm_resp
    -- Timestamps
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- When we received from SP
    processed_for_queue_at TIMESTAMPTZ,            -- When marked/picked for sending queue
    sent_to_mno_at TIMESTAMPTZ,                     -- When successfully sent to MNO
    dlr_received_at TIMESTAMPTZ,                  -- When DLR was received from MNO
    completed_at TIMESTAMPTZ,                     -- When DLR processing finished or message failed terminally
    -- Status Tracking
    -- Stage 1: Processing before sending
    processing_status VARCHAR(50) NOT NULL DEFAULT 'received',
        -- received -> validated -> routed -> pricing_checked -> queued_for_send
        -- -> failed_validation / failed_routing / failed_pricing / failed_queueing
    -- Stage 2: Status after sending attempt
    send_status VARCHAR(50),
        -- sent -> send_failed
    -- Stage 3: Delivery Status from DLR
    dlr_status VARCHAR(50),
        -- DELIVRD, EXPIRED, DELETED, UNDELIV, ACCEPTD, UNKNOWN, REJECTD (standard SMPP states)
    -- Final conclusive status
    final_status VARCHAR(50) NOT NULL DEFAULT 'pending',
        -- pending -> delivered / failed / rejected
    -- Error details
    error_code INT,
    error_description TEXT
);
-- Indexes for efficient querying and processing
CREATE INDEX idx_messages_service_provider_id ON messages(service_provider_id);
CREATE INDEX idx_messages_received_at ON messages(received_at);
CREATE INDEX idx_messages_processing_status ON messages(processing_status); -- Crucial for picking messages to process
CREATE INDEX idx_messages_sent_to_mno_at ON messages(sent_to_mno_at);
CREATE INDEX idx_messages_mno_message_id ON messages(mno_message_id); -- Crucial for matching DLRs
CREATE INDEX idx_messages_final_status ON messages(final_status); -- For reporting
CREATE INDEX idx_messages_destination_msisdn_pattern ON messages(destination_msisdn varchar_pattern_ops); -- For routing lookup if needed here


-- Delivery Reports raw data (optional, if you need to store the raw PDU)
-- Usually, updating the main 'messages' table is sufficient.
-- If storing raw DLRs:
CREATE TABLE delivery_reports_raw (
     id BIGSERIAL PRIMARY KEY,
     message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
     mno_connection_id INT NOT NULL REFERENCES mno_connections(id) ON DELETE RESTRICT,
     raw_pdu TEXT, -- Or BYTEA
     received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     processed_at TIMESTAMPTZ,
     processing_status VARCHAR(50) NOT NULL DEFAULT 'pending' -- pending -> processed / failed
 );
 CREATE INDEX idx_delivery_reports_raw_message_id ON delivery_reports_raw(message_id);
 CREATE INDEX idx_delivery_reports_raw_processing_status ON delivery_reports_raw(processing_status, received_at);


-- Wallets (One per Service Provider per Currency)
CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL, -- ISO 4217 code (e.g., NGN, USD, EUR)
    balance NUMERIC(19, 4) NOT NULL DEFAULT 0.00, -- High precision decimal/numeric
    low_balance_threshold NUMERIC(19, 4) NOT NULL DEFAULT 0.00,
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
    message_id BIGINT REFERENCES messages(id) ON DELETE SET NULL, -- Link debit to the specific message
    transaction_type VARCHAR(10) NOT NULL, -- 'debit' or 'credit'
    amount NUMERIC(19, 4) NOT NULL,
    balance_before NUMERIC(19, 4) NOT NULL,
    balance_after NUMERIC(19, 4) NOT NULL,
    description TEXT,
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_message_id ON wallet_transactions(message_id);
CREATE INDEX idx_wallet_transactions_transaction_date ON wallet_transactions(transaction_date);
CREATE INDEX idx_wallet_transactions_transaction_type ON wallet_transactions(transaction_type);

-- Pricing Rules per Service Provider
CREATE TABLE pricing_rules (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    mno_id INT REFERENCES mnos(id) ON DELETE SET NULL, -- If NULL, it's the default price for this SP
    currency_code VARCHAR(3) NOT NULL, -- Should match a wallet currency for the SP
    price_per_sms NUMERIC(10, 6) NOT NULL, -- Price per SMS segment
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Ensure an SP has only one default price per currency
    UNIQUE (service_provider_id, currency_code) WHERE mno_id IS NULL,
    -- Ensure an SP has only one price for a specific MNO per currency
    UNIQUE (service_provider_id, mno_id, currency_code) WHERE mno_id IS NOT NULL
);
CREATE INDEX idx_pricing_rules_service_provider_id ON pricing_rules(service_provider_id);
CREATE INDEX idx_pricing_rules_mno_id ON pricing_rules(mno_id);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pricing_rules;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS delivery_reports_raw; -- Uncomment if using raw DLR table
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS sender_ids;
DROP TABLE IF EXISTS routing_rules;
DROP TABLE IF EXISTS mno_connections;
DROP TABLE IF EXISTS mnos;
DROP TABLE IF EXISTS smpp_credentials;
DROP TABLE IF EXISTS service_providers;

-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd
