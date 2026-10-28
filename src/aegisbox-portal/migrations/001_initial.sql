-- Aegisbox SMS Gateway Database Schema
-- Migration 001: Initial schema

-- Service Providers (Your Customers)
CREATE TABLE IF NOT EXISTS aegisbox_service_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'pending',
    default_currency_code VARCHAR(3) DEFAULT 'NGN' NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    logo_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- SP Credentials (API keys + SMPP)
CREATE TABLE IF NOT EXISTS aegisbox_sp_credentials (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL,
    protocol VARCHAR(10) DEFAULT 'smpp' NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    system_id VARCHAR(16),
    password_hash VARCHAR(255),
    bind_type VARCHAR(10) DEFAULT 'trx',
    api_key_hash VARCHAR(255),
    api_key_identifier VARCHAR(64) UNIQUE,
    rate_limit_per_second INT DEFAULT 100,
    routing_group_id INT,
    scope VARCHAR(50) DEFAULT 'default' NOT NULL,
    webhook_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE
);

-- Mobile Network Operators
CREATE TABLE IF NOT EXISTS aegisbox_mnos (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    country_code VARCHAR(5) NOT NULL,
    network_code VARCHAR(10),
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    logo_url VARCHAR(500),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- MNO Connections (SMPP or HTTP)
CREATE TABLE IF NOT EXISTS aegisbox_mno_connections (
    id SERIAL PRIMARY KEY,
    mno_id INT NOT NULL,
    protocol VARCHAR(10) DEFAULT 'smpp' NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    priority INT DEFAULT 1 NOT NULL,
    system_id VARCHAR(16),
    password VARCHAR(255),
    host VARCHAR(255),
    port INT,
    use_tls BOOLEAN DEFAULT FALSE,
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
    api_key TEXT,
    base_url VARCHAR(255) DEFAULT '',
    username VARCHAR(255),
    http_password TEXT,
    secret_key TEXT,
    default_sender VARCHAR(100),
    rate_limit INT DEFAULT 100,
    email VARCHAR(255),
    supports_webhook BOOLEAN DEFAULT TRUE,
    webhook_path VARCHAR(255),
    timeout_secs INT DEFAULT 30,
    provider_config JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (mno_id) REFERENCES aegisbox_mnos(id) ON DELETE RESTRICT
);

-- Sender IDs
CREATE TABLE IF NOT EXISTS aegisbox_sender_ids (
    id SERIAL PRIMARY KEY,
    service_provider_id INT,
    sender_id_string VARCHAR(16) NOT NULL,
    mno_connection_id INT,
    status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    type VARCHAR(20) DEFAULT 'regular',
    provider_id INT,
    rate_limit INT DEFAULT 100,
    otp_template TEXT,
    networks TEXT[],
    max_usage_count INT DEFAULT 10000,
    current_usage_count INT DEFAULT 0,
    reset_interval_hours INT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    FOREIGN KEY (mno_connection_id) REFERENCES aegisbox_mno_connections(id) ON DELETE SET NULL
);

-- Messages
CREATE TABLE IF NOT EXISTS aegisbox_messages (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL,
    sp_credential_id INT NOT NULL,
    client_message_id VARCHAR(64),
    client_ref VARCHAR(64),
    original_source_addr VARCHAR(21) NOT NULL,
    original_destination_addr VARCHAR(21) NOT NULL,
    short_message TEXT NOT NULL,
    total_segments INT DEFAULT 1 NOT NULL,
    submitted_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    approved_sender_id INT,
    template_id INT,
    routed_mno_id INT,
    currency_code VARCHAR(3) NOT NULL,
    cost DECIMAL(19, 6) DEFAULT 0.000000,
    processing_status VARCHAR(50) DEFAULT 'received' NOT NULL,
    final_status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    error_code VARCHAR(10),
    error_description TEXT,
    processed_for_queue_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    retry_count INT DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE RESTRICT,
    FOREIGN KEY (sp_credential_id) REFERENCES aegisbox_sp_credentials(id) ON DELETE RESTRICT
);

-- Message Segments
CREATE TABLE IF NOT EXISTS aegisbox_message_segments (
    id SERIAL PRIMARY KEY,
    message_id INT NOT NULL,
    segment_seqn INT NOT NULL,
    mno_message_id VARCHAR(64) UNIQUE,
    mno_connection_id INT,
    sent_to_mno_at TIMESTAMPTZ,
    dlr_status VARCHAR(50),
    dlr_received_at TIMESTAMPTZ,
    error_code VARCHAR(10),
    error_description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (message_id) REFERENCES aegisbox_messages(id) ON DELETE CASCADE
);

-- Wallets
CREATE TABLE IF NOT EXISTS aegisbox_wallets (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL,
    currency_code VARCHAR(3) NOT NULL,
    balance DECIMAL(19, 6) DEFAULT 0.000000 NOT NULL,
    wallet_type VARCHAR(20) DEFAULT 'prepaid' NOT NULL,
    credit_limit DECIMAL(19, 6) DEFAULT 0.000000,
    auto_recharge_enabled BOOLEAN DEFAULT FALSE,
    low_balance_threshold DECIMAL(19, 6) DEFAULT 0.000000,
    low_balance_notified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    UNIQUE (service_provider_id, currency_code)
);

-- Wallet Transactions
CREATE TABLE IF NOT EXISTS aegisbox_wallet_transactions (
    id SERIAL PRIMARY KEY,
    wallet_id INT NOT NULL,
    message_id INT,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(19, 6) NOT NULL,
    balance_before DECIMAL(19, 6) NOT NULL,
    balance_after DECIMAL(19, 6) NOT NULL,
    description TEXT,
    transaction_date TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    reference_transaction_id INT,
    FOREIGN KEY (wallet_id) REFERENCES aegisbox_wallets(id) ON DELETE RESTRICT
);

-- Pricing Rules
CREATE TABLE IF NOT EXISTS aegisbox_pricing_rules (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL,
    mno_id INT,
    mno_country_code VARCHAR(5),
    currency_code VARCHAR(3) NOT NULL,
    price_per_sms DECIMAL(10, 6) NOT NULL,
    message_type VARCHAR(20) DEFAULT 'sms' NOT NULL,
    segment_length INT DEFAULT 160,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    effective_from TIMESTAMPTZ,
    effective_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE
);

-- Routing Groups
CREATE TABLE IF NOT EXISTS aegisbox_routing_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    reference VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Routing Assignments
CREATE TABLE IF NOT EXISTS aegisbox_routing_assignments (
    id SERIAL PRIMARY KEY,
    routing_group_id INT NOT NULL,
    msisdn_prefix_group_id INT NOT NULL,
    mno_id INT NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    priority INT DEFAULT 0 NOT NULL,
    failover_mno_id INT,
    comment TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (routing_group_id) REFERENCES aegisbox_routing_groups(id) ON DELETE CASCADE
);

-- MSISDN Prefix Groups
CREATE TABLE IF NOT EXISTS aegisbox_msisdn_prefix_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    reference VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- MSISDN Prefix Entries
CREATE TABLE IF NOT EXISTS aegisbox_msisdn_prefix_entries (
    id SERIAL PRIMARY KEY,
    msisdn_prefix_group_id INT NOT NULL,
    msisdn_prefix VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (msisdn_prefix_group_id) REFERENCES aegisbox_msisdn_prefix_groups(id) ON DELETE CASCADE
);

-- OTP Message Templates
CREATE TABLE IF NOT EXISTS aegisbox_otp_message_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    content_template TEXT NOT NULL,
    default_brand_name VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Templates
CREATE TABLE IF NOT EXISTS aegisbox_templates (
    id SERIAL PRIMARY KEY,
    service_provider_id INT,
    name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    UNIQUE (service_provider_id, name)
);

-- Delivery Reports Raw
CREATE TABLE IF NOT EXISTS aegisbox_delivery_reports_raw (
    id SERIAL PRIMARY KEY,
    message_segment_id INT,
    mno_connection_id INT NOT NULL,
    raw_payload TEXT,
    received_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    processed_at TIMESTAMPTZ,
    processing_status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    FOREIGN KEY (message_segment_id) REFERENCES aegisbox_message_segments(id) ON DELETE CASCADE
);

-- DLR Forwarding Queue
CREATE TABLE IF NOT EXISTS aegisbox_dlr_forwarding_queue (
    id SERIAL PRIMARY KEY,
    message_id INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' NOT NULL,
    payload JSONB NOT NULL,
    attempts INT DEFAULT 0 NOT NULL,
    max_attempts INT DEFAULT 5 NOT NULL,
    last_attempt_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(100)
);

-- Dead Letter Queue
CREATE TABLE IF NOT EXISTS aegisbox_dead_letter_queue (
    id SERIAL PRIMARY KEY,
    message_id INT NOT NULL,
    original_status VARCHAR(50),
    error_code VARCHAR(50),
    error_description TEXT,
    retry_count INT DEFAULT 0,
    failed_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB
);

-- Org Mappings (Clerk → Database)
CREATE TABLE IF NOT EXISTS aegisbox_org_mappings (
    id SERIAL PRIMARY KEY,
    clerk_org_id VARCHAR(255) NOT NULL UNIQUE,
    service_provider_id INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE
);

-- Admin Users
CREATE TABLE IF NOT EXISTS aegisbox_admin_users (
    id SERIAL PRIMARY KEY,
    clerk_user_id VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(50) DEFAULT 'admin' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Exchange Rates
CREATE TABLE IF NOT EXISTS aegisbox_exchange_rates (
    id SERIAL PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(15, 6) NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE (from_currency, to_currency)
);

-- OTP Alternative Senders
CREATE TABLE IF NOT EXISTS aegisbox_otp_alternative_senders (
    id SERIAL PRIMARY KEY,
    sender_id_string VARCHAR(16) NOT NULL UNIQUE,
    service_provider_id INT,
    mno_id INT,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    current_usage_count INT DEFAULT 0 NOT NULL,
    max_usage_count INT DEFAULT 10000 NOT NULL,
    reset_interval_hours INT,
    last_reset_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (service_provider_id) REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE
);

-- OTP Sender Template Assignments
CREATE TABLE IF NOT EXISTS aegisbox_otp_sender_template_assignments (
    id SERIAL PRIMARY KEY,
    otp_alternative_sender_id INT NOT NULL,
    otp_message_template_id INT NOT NULL,
    mno_id INT,
    priority INT DEFAULT 0 NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    assignment_usage_count DECIMAL(19) DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    FOREIGN KEY (otp_alternative_sender_id) REFERENCES aegisbox_otp_alternative_senders(id) ON DELETE CASCADE
);

-- Create Indexes
CREATE INDEX IF NOT EXISTS idx_service_providers_status ON aegisbox_service_providers(status);
CREATE INDEX IF NOT EXISTS idx_sp_credentials_service_provider ON aegisbox_sp_credentials(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_sp_credentials_api_key ON aegisbox_sp_credentials(api_key_identifier);
CREATE INDEX IF NOT EXISTS idx_mno_connections_mno ON aegisbox_mno_connections(mno_id);
CREATE INDEX IF NOT EXISTS idx_mno_connections_status ON aegisbox_mno_connections(status);
CREATE INDEX IF NOT EXISTS idx_sender_ids_service_provider ON aegisbox_sender_ids(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_sender_ids_status ON aegisbox_sender_ids(status);
CREATE INDEX IF NOT EXISTS idx_messages_service_provider ON aegisbox_messages(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_messages_received_at ON aegisbox_messages(received_at);
CREATE INDEX IF NOT EXISTS idx_messages_processing_status ON aegisbox_messages(processing_status);
CREATE INDEX IF NOT EXISTS idx_messages_final_status ON aegisbox_messages(final_status);
CREATE INDEX IF NOT EXISTS idx_message_segments_message ON aegisbox_message_segments(message_id);
CREATE INDEX IF NOT EXISTS idx_wallets_service_provider ON aegisbox_wallets(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet ON aegisbox_wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_pricing_rules_service_provider ON aegisbox_pricing_rules(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_routing_groups_status ON aegisbox_routing_groups(status);
CREATE INDEX IF NOT EXISTS idx_routing_assignments_routing_group ON aegisbox_routing_assignments(routing_group_id);
CREATE INDEX IF NOT EXISTS idx_msisdn_prefix_entries_group ON aegisbox_msisdn_prefix_entries(msisdn_prefix_group_id);
