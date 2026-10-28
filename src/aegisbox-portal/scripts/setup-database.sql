-- Aegisbox SMS Gateway Platform - Database Schema
-- All tables prefixed with aegisbox_ for production safety

-- Service Providers (Your Customers)
CREATE TABLE IF NOT EXISTS aegisbox_service_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    default_currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    phone VARCHAR(20),
    address TEXT,
    logo_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SP Credentials (How SPs connect to YOUR system - SMPP or HTTP)
CREATE TABLE IF NOT EXISTS aegisbox_sp_credentials (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    system_id VARCHAR(16) UNIQUE,
    password_hash VARCHAR(255),
    bind_type VARCHAR(10) DEFAULT 'trx',
    api_key_hash VARCHAR(255),
    api_key_identifier VARCHAR(64) UNIQUE,
    api_key_prefix VARCHAR(10) DEFAULT 'aegis_',
    rate_limit_per_second INT DEFAULT 100,
    max_concurrent_connections INT DEFAULT 5,
    scope VARCHAR(50) NOT NULL DEFAULT 'default',
    routing_group_id INT,
    webhook_url VARCHAR(500),
    webhook_secret VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SP Wallets (Prepaid or Postpaid)
CREATE TABLE IF NOT EXISTS aegisbox_wallets (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    balance DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    wallet_type VARCHAR(20) NOT NULL DEFAULT 'prepaid',
    credit_limit DECIMAL(15, 4) DEFAULT 0.0000,
    auto_recharge_enabled BOOLEAN DEFAULT false,
    auto_recharge_threshold DECIMAL(15, 4) DEFAULT 100.0000,
    auto_recharge_amount DECIMAL(15, 4) DEFAULT 1000.0000,
    low_balance_threshold DECIMAL(15, 4) DEFAULT 50.0000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Wallet Transactions
CREATE TABLE IF NOT EXISTS aegisbox_wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id INT NOT NULL REFERENCES aegisbox_wallets(id) ON DELETE CASCADE,
    transaction_type VARCHAR(30) NOT NULL,
    amount DECIMAL(15, 4) NOT NULL,
    balance_before DECIMAL(15, 4) NOT NULL,
    balance_after DECIMAL(15, 4) NOT NULL,
    description TEXT,
    reference_id VARCHAR(100),
    reference_type VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pricing/Fee Groups
CREATE TABLE IF NOT EXISTS aegisbox_pricing_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Mobile Network Operators (MNOs)
CREATE TABLE IF NOT EXISTS aegisbox_mnos (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    country_code VARCHAR(5) NOT NULL,
    network_code VARCHAR(10),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    logo_url VARCHAR(500),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pricing Rules
CREATE TABLE IF NOT EXISTS aegisbox_pricing_rules (
    id SERIAL PRIMARY KEY,
    pricing_group_id INT NOT NULL REFERENCES aegisbox_pricing_groups(id) ON DELETE CASCADE,
    mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    mno_country_code VARCHAR(5),
    message_type VARCHAR(20) NOT NULL DEFAULT 'sms',
    price_per_segment DECIMAL(10, 4) NOT NULL,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    segment_length INT DEFAULT 160,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- MNO Connections
CREATE TABLE IF NOT EXISTS aegisbox_mno_connections (
    id SERIAL PRIMARY KEY,
    mno_id INT NOT NULL REFERENCES aegisbox_mnos(id) ON DELETE RESTRICT,
    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    priority INT DEFAULT 1,
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
    provider_config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- MSISDN Prefix Groups
CREATE TABLE IF NOT EXISTS aegisbox_msisdn_prefix_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    reference VARCHAR(100) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- MSISDN Prefix Entries
CREATE TABLE IF NOT EXISTS aegisbox_msisdn_prefix_entries (
    id SERIAL PRIMARY KEY,
    msisdn_prefix_group_id INT NOT NULL REFERENCES aegisbox_msisdn_prefix_groups(id) ON DELETE CASCADE,
    msisdn_prefix VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (msisdn_prefix_group_id, msisdn_prefix)
);

-- Routing Groups
CREATE TABLE IF NOT EXISTS aegisbox_routing_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    reference VARCHAR(100) UNIQUE,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Routing Assignments
CREATE TABLE IF NOT EXISTS aegisbox_routing_assignments (
    id SERIAL PRIMARY KEY,
    routing_group_id INT NOT NULL REFERENCES aegisbox_routing_groups(id) ON DELETE CASCADE,
    msisdn_prefix_group_id INT NOT NULL REFERENCES aegisbox_msisdn_prefix_groups(id) ON DELETE RESTRICT,
    mno_id INT NOT NULL REFERENCES aegisbox_mnos(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    priority INT NOT NULL DEFAULT 0,
    failover_mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (routing_group_id, msisdn_prefix_group_id)
);

-- Messages (The main SMS record)
CREATE TABLE IF NOT EXISTS aegisbox_messages (
    id BIGSERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES aegisbox_service_providers(id) ON DELETE RESTRICT,
    sp_credential_id INT REFERENCES aegisbox_sp_credentials(id) ON DELETE SET NULL,
    wallet_transaction_id BIGINT REFERENCES aegisbox_wallet_transactions(id) ON DELETE SET NULL,
    external_message_id VARCHAR(100) UNIQUE,
    internal_message_id UUID DEFAULT gen_random_uuid(),
    sender_name VARCHAR(100),
    sender_id VARCHAR(100),
    recipient_msisdn VARCHAR(20) NOT NULL,
    message_text TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'sms',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    delivery_status VARCHAR(50) DEFAULT 'pending',
    segments INT DEFAULT 1,
    cost_per_segment DECIMAL(10, 4),
    total_cost DECIMAL(15, 4),
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    failure_reason TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    routing_group_id INT,
    mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    mno_connection_id INT REFERENCES aegisbox_mno_connections(id) ON DELETE SET NULL,
    dlr_received BOOLEAN DEFAULT false,
    dlr_received_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Admin Users (Clerk auth will sync here)
CREATE TABLE IF NOT EXISTS aegisbox_admin_users (
    id SERIAL PRIMARY KEY,
    clerk_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    permissions JSONB,
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SP Users (Clerk auth will sync here)
CREATE TABLE IF NOT EXISTS aegisbox_sp_users (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    clerk_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    permissions JSONB,
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_sp_providers_email ON aegisbox_service_providers(email);
CREATE INDEX IF NOT EXISTS idx_sp_providers_status ON aegisbox_service_providers(status);
CREATE INDEX IF NOT EXISTS idx_sp_credentials_sp_id ON aegisbox_sp_credentials(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_wallets_sp_id ON aegisbox_wallets(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON aegisbox_wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_messages_sp_id ON aegisbox_messages(service_provider_id);
CREATE INDEX IF NOT EXISTS idx_messages_status ON aegisbox_messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON aegisbox_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_admin_users_clerk_id ON aegisbox_admin_users(clerk_id);
CREATE INDEX IF NOT EXISTS idx_sp_users_clerk_id ON aegisbox_sp_users(clerk_id);
CREATE INDEX IF NOT EXISTS idx_sp_users_sp_id ON aegisbox_sp_users(service_provider_id);
