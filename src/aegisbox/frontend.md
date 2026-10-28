# Aegisbox Frontend Development Prompt

## Project Overview

You are building a **dual-portal SMS Gateway Administration Platform** for Aegisbox. The system consists of two connected portals:

1. **Admin Portal** - For platform administrators to manage the entire SMS gateway
2. **Service Provider Portal** - For customers (Service Providers) to manage their SMS operations

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         AEGISBOX SMS GATEWAY                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────┐     ┌─────────────────────────────────────┐   │
│  │    ADMIN PORTAL     │     │      SERVICE PROVIDER PORTAL        │   │
│  │  (Clerk Auth)       │     │        (Clerk Auth)                │   │
│  │                     │     │                                       │   │
│  │ • Service Providers │     │ • Dashboard & Analytics              │   │
│  │ • MNO Management   │     │ • Send SMS                          │   │
│  │ • Routing Rules    │     │ • Transaction History                │   │
│  │ • Pricing/Fees     │     │ • Wallet Management                 │   │
│  │ • Transactions     │     │ • Sender IDs                         │   │
│  │ • Wallets         │     │ • API Keys                           │   │
│  │ • Analytics       │     │ • Reports                           │   │
│  └─────────┬───────────┘     └─────────────────┬───────────────────┘   │
│            │                                   │                         │
│            └───────────────┬───────────────────┘                         │
│                            │                                             │
│              ┌─────────────▼─────────────────────┐                      │
│              │        NEX.JS BACKEND API         │                      │
│              │     (Next.js App Router)           │                      │
│              │                                     │                      │
│              │  • Authentication (Clerk)            │                      │
│              │  • Authorization (RBAC)              │                      │
│              │  • Drizzle ORM                      │                      │
│              │  • API Routes / Server Actions      │                      │
│              └─────────────────┬───────────────────┘                      │
│                                │                                          │
│              ┌─────────────────▼──────────────────┐                    │
│              │         POSTGRES DATABASE            │                    │
│              │  (Drizzle ORM + pgvector)           │                    │
│              └──────────────────────────────────────┘                    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    EXTERNAL CONNECTIONS                          │   │
│  │                                                                 │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │   │
│  │  │   SMPP    │  │  HTTP    │  │  Clerk   │  │  Email   │    │   │
│  │  │  Servers  │  │ Providers│  │  Auth    │  │  Service │    │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

## Database Schema (Complete)

All tables are prefixed with `aegisbox_` for production safety.

### Core Tables

```sql
-- Service Providers (Your Customers)
CREATE TABLE aegisbox_service_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- 'pending' | 'active' | 'suspended' | 'cancelled'
    default_currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    phone VARCHAR(20),
    address TEXT,
    logo_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SP Credentials (How SPs connect to YOUR system - SMPP or HTTP)
CREATE TABLE aegisbox_sp_credentials (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp',
    -- 'smpp' | 'http'
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    -- 'active' | 'inactive' | 'suspended'
    
    -- SMPP Specific Fields (for SPs connecting via SMPP to us)
    system_id VARCHAR(16) UNIQUE,
    password_hash VARCHAR(255),
    bind_type VARCHAR(10) DEFAULT 'trx',
    
    -- HTTP Specific Fields (for SPs connecting via HTTP API)
    api_key_hash VARCHAR(255),
    api_key_identifier VARCHAR(64) UNIQUE,
    api_key_prefix VARCHAR(10) DEFAULT 'aegis_',
    
    -- Rate Limiting
    rate_limit_per_second INT DEFAULT 100,
    max_concurrent_connections INT DEFAULT 5,
    
    -- Scope & Routing
    scope VARCHAR(50) NOT NULL DEFAULT 'default',
    -- 'default' | 'dedicated'
    routing_group_id INT,
    
    -- Webhook Configuration (for HTTP SPs receiving DLRs)
    webhook_url VARCHAR(500),
    webhook_secret VARCHAR(255),
    
    -- Audit Fields
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SP Wallets (Prepaid or Postpaid)
CREATE TABLE aegisbox_wallets (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    balance DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    wallet_type VARCHAR(20) NOT NULL DEFAULT 'prepaid',
    -- 'prepaid' | 'postpaid' | 'mixed'
    credit_limit DECIMAL(15, 4) DEFAULT 0.0000,
    auto_recharge_enabled BOOLEAN DEFAULT false,
    auto_recharge_threshold DECIMAL(15, 4) DEFAULT 100.0000,
    auto_recharge_amount DECIMAL(15, 4) DEFAULT 1000.0000,
    low_balance_threshold DECIMAL(15, 4) DEFAULT 50.0000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Wallet Transactions
CREATE TABLE aegisbox_wallet_transactions (
    id SERIAL PRIMARY KEY,
    wallet_id INT NOT NULL REFERENCES aegisbox_wallets(id) ON DELETE CASCADE,
    transaction_type VARCHAR(30) NOT NULL,
    -- 'credit' | 'debit' | 'reversal' | 'adjustment' | 'auto_recharge'
    amount DECIMAL(15, 4) NOT NULL,
    balance_before DECIMAL(15, 4) NOT NULL,
    balance_after DECIMAL(15, 4) NOT NULL,
    description TEXT,
    reference_id VARCHAR(100),
    reference_type VARCHAR(50),
    -- 'manual_credit' | 'sms_transaction' | 'refund' | 'adjustment'
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pricing/Fee Groups (Define pricing tiers)
CREATE TABLE aegisbox_pricing_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pricing Rules (Specific rates within a pricing group)
CREATE TABLE aegisbox_pricing_rules (
    id SERIAL PRIMARY KEY,
    pricing_group_id INT NOT NULL REFERENCES aegisbox_pricing_groups(id) ON DELETE CASCADE,
    mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    mno_country_code VARCHAR(5),
    -- Country code like '234' for Nigeria
    message_type VARCHAR(20) NOT NULL DEFAULT 'sms',
    -- 'sms' | 'otp' | 'flash' | 'unicode'
    price_per_segment DECIMAL(10, 4) NOT NULL,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'NGN',
    segment_length INT DEFAULT 160,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Mobile Network Operators (MNOs)
CREATE TABLE aegisbox_mnos (
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

-- MNO Connections (Our connections TO MNOs - SMPP or HTTP)
CREATE TABLE aegisbox_mno_connections (
    id SERIAL PRIMARY KEY,
    mno_id INT NOT NULL REFERENCES aegisbox_mnos(id) ON DELETE RESTRICT,
    protocol VARCHAR(10) NOT NULL DEFAULT 'smpp',
    -- 'smpp' | 'http'
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
    
    -- HTTP Specific Fields
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
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- MSISDN Prefix Groups (For routing by phone number prefix)
CREATE TABLE aegisbox_msisdn_prefix_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    reference VARCHAR(100) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- MSISDN Prefix Entries
CREATE TABLE aegisbox_msisdn_prefix_entries (
    id SERIAL PRIMARY KEY,
    msisdn_prefix_group_id INT NOT NULL REFERENCES aegisbox_msisdn_prefix_groups(id) ON DELETE CASCADE,
    msisdn_prefix VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (msisdn_prefix_group_id, msisdn_prefix)
);

-- Routing Groups
CREATE TABLE aegisbox_routing_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    reference VARCHAR(100) UNIQUE,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Routing Assignments (Which MNO to use for which prefix)
CREATE TABLE aegisbox_routing_assignments (
    id SERIAL PRIMARY KEY,
    routing_group_id INT NOT NULL REFERENCES aegisbox_routing_groups(id) ON DELETE CASCADE,
    msisdn_prefix_group_id INT NOT NULL REFERENCES aegisbox_msisdn_prefix_groups(id) ON DELETE RESTRICT,
    mno_id INT NOT NULL REFERENCES aegisbox_mnos(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    priority INT NOT NULL DEFAULT 0,
    -- Lower number = higher priority
    failover_mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (routing_group_id, msisdn_prefix_group_id)
);

-- Messages (The main SMS record)
CREATE TABLE aegisbox_messages (
    id BIGSERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES aegisbox_service_providers(id) ON DELETE RESTRICT,
    sp_credential_id INT REFERENCES aegisbox_sp_credentials(id) ON DELETE SET NULL,
    wallet_transaction_id BIGINT REFERENCES aegisbox_wallet_transactions(id) ON DELETE SET NULL,
    
    -- Message Identification
    external_message_id VARCHAR(100) UNIQUE,
    internal_message_id UUID DEFAULT gen_random_uuid(),
    
    -- Sender & Recipient
    sender_id VARCHAR(16) NOT NULL,
    destination_msisdn VARCHAR(20) NOT NULL,
    country_code VARCHAR(5),
    
    -- Content
    message_content TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'sms',
    -- 'sms' | 'otp' | 'flash' | 'unicode' | 'binary'
    is_otp BOOLEAN DEFAULT false,
    
    -- Routing
    mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    mno_connection_id INT REFERENCES aegisbox_mno_connections(id) ON DELETE SET NULL,
    routing_group_id INT REFERENCES aegisbox_routing_groups(id) ON DELETE SET NULL,
    
    -- Pricing
    pricing_rule_id BIGINT REFERENCES aegisbox_pricing_rules(id) ON DELETE SET NULL,
    cost_per_segment DECIMAL(10, 4) DEFAULT 0,
    total_cost DECIMAL(15, 4) DEFAULT 0,
    currency_code VARCHAR(3) DEFAULT 'NGN',
    
    -- Status
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    -- 'pending' | 'queued' | 'validated' | 'routed' | 'sent' | 
    -- 'delivered' | 'failed' | 'rejected' | 'expired'
    status_reason VARCHAR(200),
    
    -- Timing
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    
    -- Error Handling
    error_code VARCHAR(50),
    error_description TEXT,
    
    -- Metadata
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Message Segments (For long messages)
CREATE TABLE aegisbox_message_segments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES aegisbox_messages(id) ON DELETE CASCADE,
    segment_seqn INT NOT NULL,
    udh_header VARCHAR(10),
    message_content TEXT NOT NULL,
    encoding VARCHAR(20) DEFAULT 'GSM7',
    segment_length INT DEFAULT 160,
    
    -- MNO Response
    mno_message_id VARCHAR(100),
    
    -- Status
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    -- 'pending' | 'sent' | 'delivered' | 'failed' | 'expired'
    
    -- Timing
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    
    -- Error
    error_code VARCHAR(50),
    error_description TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (message_id, segment_seqn)
);

-- OTP Templates
CREATE TABLE aegisbox_otp_message_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    content_template TEXT NOT NULL,
    -- Variables: [OTP], [APP_NAME], [BRAND_NAME], [VALIDITY_MINUTES]
    default_brand_name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sender IDs (Approved sender IDs for SPs)
CREATE TABLE aegisbox_sender_ids (
    id SERIAL PRIMARY KEY,
    service_provider_id INT REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    sender_id_string VARCHAR(16) NOT NULL,
    mno_connection_id INT REFERENCES aegisbox_mno_connections(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- 'pending' | 'active' | 'rejected' | 'suspended'
    
    -- OTP Configuration
    is_otp_enabled BOOLEAN DEFAULT false,
    otp_template_id INT REFERENCES aegisbox_otp_message_templates(id) ON DELETE SET NULL,
    otp_max_usage_count INT DEFAULT 10000,
    otp_current_usage_count INT DEFAULT 0,
    otp_reset_interval_hours INT DEFAULT 24,
    otp_last_reset_at TIMESTAMPTZ,
    otp_last_used_at TIMESTAMPTZ,
    
    -- Rate Limits
    rate_limit INTEGER DEFAULT 100,
    daily_limit INTEGER DEFAULT 10000,
    monthly_limit INTEGER DEFAULT 300000,
    
    -- Tracking
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (service_provider_id, sender_id_string)
);

-- OTP Alternative Senders (For OTP failover)
CREATE TABLE aegisbox_otp_alternative_senders (
    id SERIAL PRIMARY KEY,
    sender_id_string VARCHAR(16) NOT NULL,
    service_provider_id INT REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_usage_count INT DEFAULT 0,
    max_usage_count INT DEFAULT 10000,
    reset_interval_hours INT DEFAULT 24,
    last_reset_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OTP Sender-Template Assignments
CREATE TABLE aegisbox_otp_sender_template_assignments (
    id SERIAL PRIMARY KEY,
    sender_id INT NOT NULL REFERENCES aegisbox_sender_ids(id) ON DELETE CASCADE,
    template_id INT NOT NULL REFERENCES aegisbox_otp_message_templates(id) ON DELETE CASCADE,
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    usage_count INT DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- DLR (Delivery Receipt) Log
CREATE TABLE aegisbox_dlr_logs (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT REFERENCES aegisbox_messages(id) ON DELETE CASCADE,
    segment_id BIGINT REFERENCES aegisbox_message_segments(id) ON DELETE SET NULL,
    
    -- MNO DLR Info
    mno_message_id VARCHAR(100),
    mno_connection_id INT REFERENCES aegisbox_mno_connections(id) ON DELETE SET NULL,
    
    -- DLR Status
    status VARCHAR(30) NOT NULL,
    -- 'delivered' | 'failed' | 'undelivered' | 'expired' | 'rejected' | 'unknown'
    dlr_raw_data TEXT,
    
    -- Error Info
    error_code VARCHAR(50),
    error_description TEXT,
    
    -- Timing
    submit_date TIMESTAMPTZ,
    done_date TIMESTAMPTZ,
    received_at TIMESTAMPTZ DEFAULT NOW(),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Analytics Daily Summary (Pre-computed daily stats)
CREATE TABLE aegisbox_analytics_daily (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    service_provider_id INT REFERENCES aegisbox_service_providers(id) ON DELETE CASCADE,
    mno_id INT REFERENCES aegisbox_mnos(id) ON DELETE SET NULL,
    
    -- Volume Stats
    total_messages INT DEFAULT 0,
    delivered_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    pending_count INT DEFAULT 0,
    sent_count INT DEFAULT 0,
    
    -- OTP Stats
    otp_messages INT DEFAULT 0,
    otp_delivered_count INT DEFAULT 0,
    
    -- Revenue Stats
    total_revenue DECIMAL(15, 4) DEFAULT 0,
    currency_code VARCHAR(3) DEFAULT 'NGN',
    
    -- Segment Stats
    total_segments INT DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, service_provider_id, mno_id)
);

-- System Configuration
CREATE TABLE aegisbox_system_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) DEFAULT 'string',
    -- 'string' | 'number' | 'boolean' | 'json'
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Admin Users (For additional admin accounts beyond Clerk)
CREATE TABLE aegisbox_admin_users (
    id SERIAL PRIMARY KEY,
    clerk_user_id VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(200),
    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    -- 'super_admin' | 'admin' | 'support' | 'viewer'
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit Log (For admin actions)
CREATE TABLE aegisbox_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT REFERENCES aegisbox_admin_users(id) ON DELETE SET NULL,
    user_type VARCHAR(20) NOT NULL DEFAULT 'admin',
    -- 'admin' | 'service_provider' | 'system'
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for Performance
CREATE INDEX idx_messages_service_provider ON aegisbox_messages(service_provider_id);
CREATE INDEX idx_messages_status ON aegisbox_messages(status);
CREATE INDEX idx_messages_created_at ON aegisbox_messages(created_at DESC);
CREATE INDEX idx_messages_mno_id ON aegisbox_messages(mno_id);
CREATE INDEX idx_segments_message_id ON aegisbox_message_segments(message_id);
CREATE INDEX idx_wallet_transactions_wallet_id ON aegisbox_wallet_transactions(wallet_id);
CREATE INDEX idx_analytics_daily_date ON aegisbox_analytics_daily(date DESC);
CREATE INDEX idx_audit_logs_created_at ON aegisbox_audit_logs(created_at DESC);
CREATE INDEX idx_dlr_logs_message_id ON aegisbox_dlr_logs(message_id);
```

## Technology Stack

```
Frontend:
- Next.js 14+ (App Router)
- TypeScript
- Tailwind CSS
- shadcn/ui (UI components)
- Lucide React (icons)
- Recharts (charts & analytics)
- date-fns (date formatting)
- React Hook Form + Zod (forms & validation)
- TanStack Query (server state)
- Zustand (client state)

Backend (Next.js API Routes):
- Next.js Server Actions
- Drizzle ORM
- PostgreSQL
- Clerk (authentication)
- Zod (validation)

DevOps:
- Docker
- Vercel (deployment)
```

## Authentication Flow (Clerk)

### Admin Portal Authentication
```
1. Admin visits /login
2. Clerk handles authentication (email/password, Google, etc.)
3. On successful login, Clerk webhook syncs user to aegisbox_admin_users table
4. User redirected to /dashboard
5. API routes verify Clerk session and check aegisbox_admin_users.role
6. RBAC middleware checks permissions for each resource
```

### Service Provider Portal Authentication
```
1. SP user visits /sp/login
2. Clerk handles authentication
3. On successful login, system looks up service_provider_id from SP credentials
4. User redirected to /sp/dashboard
5. All data filtered by service_provider_id
6. Users can only see their own data (row-level security)
```

### Clerk Webhook Handler
Create a webhook endpoint to sync Clerk users:
```typescript
// app/api/webhooks/clerk/route.ts
export async function POST(req: Request) {
  const evt = await clerkClient.webhooks.receive(
    (await req.json()) as WebhookEvent
  );
  
  switch (evt.type) {
    case 'user.created':
      // Sync to aegisbox_admin_users or create SP user
      break;
    case 'user.updated':
      // Update user record
      break;
    case 'user.deleted':
      // Deactivate user
      break;
  }
  
  return Response.json({ success: true });
}
```

## Admin Portal Features & UI Flow

### 1. Dashboard (`/admin/dashboard`)
**Purpose**: Overview of entire platform

**Components**:
- **Stats Cards** (top row):
  - Total Service Providers (with trend)
  - Active Messages Today
  - Delivery Rate (percentage)
  - Total Revenue Today
  
- **Revenue Chart**: Line chart showing revenue over last 7/30/90 days
- **Messages Volume Chart**: Bar chart of messages by day
- **Top SPs Table**: Top 5 service providers by volume
- **Recent Transactions**: Latest 10 transactions with status
- **System Health**: Connection status to MNOs

**API Endpoints**:
- `GET /api/admin/dashboard/stats`
- `GET /api/admin/dashboard/charts/revenue`
- `GET /api/admin/dashboard/charts/volume`
- `GET /api/admin/dashboard/top-sp`
- `GET /api/admin/dashboard/recent-transactions`
- `GET /api/admin/dashboard/mno-status`

### 2. Service Providers (`/admin/service-providers`)
**Purpose**: Manage all service providers

**Tabs**:
- **List View** (`/admin/service-providers`):
  - Search by name/email
  - Filter by status (pending, active, suspended)
  - Filter by created date range
  - Table with: Name, Email, Status, Wallet Balance, Messages Today, Created At
  - Actions: View, Edit, Suspend, Delete
  
- **Detail View** (`/admin/service-providers/[id]`):
  - **Overview Tab**: SP details, stats, wallet summary
  - **Credentials Tab**: API keys, SMPP credentials
  - **Wallet Tab**: Balance, transactions, adjustments
  - **Messages Tab**: All messages sent by this SP
  - **Senders Tab**: Approved sender IDs
  - **Activity Tab**: Audit log of changes

**API Endpoints**:
- `GET /api/admin/service-providers` (paginated list with filters)
- `GET /api/admin/service-providers/[id]`
- `POST /api/admin/service-providers`
- `PUT /api/admin/service-providers/[id]`
- `DELETE /api/admin/service-providers/[id]`
- `POST /api/admin/service-providers/[id]/suspend`
- `POST /api/admin/service-providers/[id]/activate`

### 3. Create Service Provider (`/admin/service-providers/new`)
**Purpose**: Add new service provider

**Form Fields**:
```typescript
// Zod Schema
const createSPSchema = z.object({
  name: z.string().min(2).max(255),
  email: z.string().email(),
  phone: z.string().optional(),
  address: z.string().optional(),
  currency_code: z.string().length(3).default('NGN'),
  wallet_type: z.enum(['prepaid', 'postpaid', 'mixed']).default('prepaid'),
  credit_limit: z.number().min(0).default(0),
  auto_recharge_enabled: z.boolean().default(false),
  auto_recharge_threshold: z.number().min(0).optional(),
  auto_recharge_amount: z.number().min(0).optional(),
  // Pricing Group Assignment
  pricing_group_id: z.number().optional(),
  // Routing Group Assignment
  routing_group_id: z.number().optional(),
  // Initial Wallet Balance (for prepaid)
  initial_balance: z.number().min(0).default(0),
});
```

**UI Flow**:
1. Fill basic info (name, email, phone)
2. Select wallet type (prepaid/postpaid)
3. Set credit limit (if postpaid)
4. Choose pricing group
5. Choose routing group
6. Optionally add initial wallet balance
7. Review and submit
8. Auto-generate API key
9. Show API credentials to admin (display once)

### 4. Pricing Groups (`/admin/pricing`)
**Purpose**: Manage pricing tiers

**UI Components**:
- **Pricing Group List**: Cards showing group name, currency, active rules count
- **Pricing Rule Table**: Within a group, list rules by MNO/country
- **Quick Add Rule Modal**: MNO select, country select, price per segment

**Form**:
```typescript
const pricingGroupSchema = z.object({
  name: z.string().min(1).max(100),
  description: z.string().optional(),
  currency_code: z.string().length(3),
  is_default: z.boolean().default(false),
});

const pricingRuleSchema = z.object({
  pricing_group_id: z.number(),
  mno_id: z.number().optional(),
  mno_country_code: z.string().optional(),
  message_type: z.enum(['sms', 'otp', 'flash', 'unicode']).default('sms'),
  price_per_segment: z.number().min(0),
  segment_length: z.number().min(1).max(1600).default(160),
  effective_from: z.date().optional(),
  effective_until: z.date().optional(),
});
```

### 5. MNO Management (`/admin/mnos`)
**Purpose**: Configure MNO connections (SMPP and HTTP)

**MNO List View**:
- Cards showing MNO name, country, active connections, status
- Quick stats: messages today, delivery rate

**MNO Detail View** (`/admin/mnos/[id]`):

**Tabs**:
1. **Overview**: MNO info, aggregate stats
2. **Connections** (`/admin/mnos/[id]/connections`):
   - List all connections (SMPP + HTTP)
   - Add new connection
   - Connection status indicator (active/inactive)
   
3. **Connection Form**:
```typescript
// SMPP Connection
const smppConnectionSchema = z.object({
  name: z.string(),
  system_id: z.string(),
  password: z.string(),
  host: z.string().ip(),
  port: z.number().port(),
  use_tls: z.boolean(),
  bind_type: z.enum(['transceiver', 'transmitter', 'receiver']),
  system_type: z.string().optional(),
  enquire_link_interval_secs: z.number().default(30),
  request_timeout_secs: z.number().default(10),
  max_window_size: z.number().default(10),
  default_data_coding: z.number().default(0),
  priority: z.number().default(1),
});

// HTTP Connection
const httpConnectionSchema = z.object({
  name: z.string(),
  base_url: z.string().url(),
  api_key: z.string().optional(),
  username: z.string().optional(),
  http_password: z.string().optional(),
  secret_key: z.string().optional(),
  timeout_secs: z.number().default(30),
  supports_webhook: z.boolean(),
  webhook_path: z.string().optional(),
  rate_limit: z.number().default(100),
  priority: z.number().default(1),
});
```

4. **Routing** (`/admin/mnos/[id]/routing`):
   - Show which prefixes route to this MNO
   - Add/remove prefix assignments

5. **Analytics** (`/admin/mnos/[id]/analytics`):
   - Daily message volume
   - Delivery rate
   - Response times

### 6. Routing Configuration (`/admin/routing`)
**Purpose**: Configure message routing rules

**UI Components**:
- **Routing Groups List**: Cards showing group name, reference, rule count
- **Routing Rules Table**: Within a group, show prefix group → MNO mapping
- **Visual Routing Flow**: Diagram showing routing path

**Create Routing Group**:
```typescript
const routingGroupSchema = z.object({
  name: z.string().min(1).max(100),
  description: z.string().optional(),
  reference: z.string().regex(/^[a-zA-Z0-9_]+$/),
  is_default: z.boolean().default(false),
});
```

**Add Routing Rule**:
```typescript
const routingRuleSchema = z.object({
  routing_group_id: z.number(),
  msisdn_prefix_group_id: z.number(),
  mno_id: z.number(),
  priority: z.number().default(0),
  failover_mno_id: z.number().optional(),
  comment: z.string().optional(),
});
```

**Prefix Group Management**:
```typescript
const prefixGroupSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  reference: z.string().regex(/^[a-zA-Z0-9_]+$/),
  // Bulk add prefixes with textarea
  prefixes_textarea: z.string(),
});
```

**Bulk Add Prefixes UI**:
- Textarea where admin can paste prefixes (one per line)
- Validates format (numbers only)
- Shows preview of valid/invalid entries
- "Add Valid Prefixes" button

### 7. Transactions (`/admin/transactions`)
**Purpose**: View all SMS transactions

**Filters**:
- Date range (created_at)
- Status (pending, sent, delivered, failed)
- Service Provider
- MNO
- Message type (sms, otp)
- Search by message ID or phone number
- Cost range

**Table Columns**:
- Message ID (internal & external)
- SP Name
- Sender ID
- Recipient (masked)
- MNO
- Status
- Cost
- Created At
- Actions (View Details, Resend, Refund)

**Detail View Modal**:
- Full message content
- All segments
- DLR timeline
- Pricing breakdown

**API Endpoints**:
- `GET /api/admin/transactions` (paginated with filters)
- `GET /api/admin/transactions/[id]`
- `POST /api/admin/transactions/[id]/refund`
- `POST /api/admin/transactions/[id]/resend`
- `GET /api/admin/transactions/export` (CSV/Excel)

### 8. Wallet Management (`/admin/wallets`)
**Purpose**: View and manage SP wallets

**Wallet List**:
- SP Name
- Balance
- Wallet Type
- Credit Limit (for postpaid)
- Low Balance Alert
- Last Activity
- Actions: View, Adjust

**Wallet Detail** (`/admin/wallets/[id]`):
- Balance summary card
- Transaction History (credit/debit with details)
- Adjust Balance Modal:
```typescript
const adjustmentSchema = z.object({
  amount: z.number(),
  transaction_type: z.enum(['credit', 'debit', 'adjustment']),
  description: z.string(),
  reference_id: z.string().optional(),
});
```

### 9. Analytics & Reports (`/admin/analytics`)
**Purpose**: Detailed analytics and reporting

**Sections**:
1. **Overview Dashboard**:
   - Revenue by day/week/month
   - Messages by status (pie chart)
   - Messages by MNO (bar chart)
   - Delivery rate trends
   - OTP success rate

2. **Service Provider Analytics**:
   - SP ranking by volume
   - SP ranking by revenue
   - Top performers
   - Underperformers

3. **MNO Performance**:
   - Delivery rates by MNO
   - Response times by MNO
   - Failure reasons breakdown

4. **Revenue Reports**:
   - Daily/Monthly revenue
   - Revenue by SP
   - Revenue by MNO
   - Outstanding balances

5. **Export Reports**:
   - Date range selector
   - Group by (day/week/month)
   - Format (CSV, Excel, PDF)
   - Scheduled reports (email)

### 10. Settings (`/admin/settings`)
**Purpose**: Platform configuration

**Sections**:
- **General**: Platform name, timezone, date format
- **Notifications**: Email settings, alert thresholds
- **SMS Defaults**: Default sender, encoding, validity period
- **Security**: API rate limits, session timeouts
- **Billing**: Invoice settings, tax configuration
- **Webhooks**: Platform webhook endpoints

## Service Provider Portal Features & UI Flow

### 1. SP Dashboard (`/sp/dashboard`)
**Purpose**: SP's personal overview

**Components**:
- **Stats Cards**:
  - Wallet Balance
  - Messages Today
  - Delivery Rate
  - This Month's Spend
  
- **Quick Actions**:
  - Send SMS button (opens compose modal)
  - Buy Credit button (opens payment modal)
  
- **Recent Messages**: Last 5 messages with status
- **Monthly Spend Chart**: Bar chart of spend by day

### 2. Send SMS (`/sp/compose`)
**Purpose**: Compose and send SMS

**Form**:
```typescript
const composeSchema = z.object({
  recipients: z.string().min(1),
  // Multiple formats: comma-separated, newline-separated, file upload
  sender_id: z.string().min(1),
  message: z.string().min(1).max(1600),
  message_type: z.enum(['sms', 'otp']).default('sms'),
  schedule_at: z.date().optional(),
  // OTP specific
  otp_template_id: z.number().optional(),
});
```

**Features**:
- Recipient input with validation (phone format)
- Sender ID dropdown (only approved senders)
- Character counter with segmentation preview
- OTP template selector (if OTP message type)
- Send now or schedule for later
- Delivery confirmation checkbox
- Save as draft option

### 3. Message History (`/sp/messages`)
**Purpose**: View sent messages

**Filters**:
- Date range
- Status
- Sender ID
- Search by phone number

**Features**:
- Same as admin but filtered to current SP
- Export functionality
- Message detail modal with DLR info

### 4. Sender IDs (`/sp/senders`)
**Purpose**: Manage approved sender IDs

**List**:
- Sender ID string
- Type (regular/OTP)
- Status
- Rate limits
- Usage today
- Actions: View, Edit OTP settings

**Request New Sender**:
- Sender ID string
- Description
- Document upload (if required)
- OTP template selection (if OTP)

### 5. Wallet (`/sp/wallet`)
**Purpose**: Manage wallet and payments

**Sections**:
- **Balance Card**: Current balance, credit limit, currency
- **Quick Credit**: Input amount, payment method
- **Transaction History**: List of credits/debits
- **Auto-Recharge Settings**: Configure threshold and amount
- **Payment Methods**: Add/remove payment methods

**Buy Credit Flow**:
1. Enter amount
2. Select payment method (if stored) or add new
3. Review and pay
4. Redirect to payment provider (Paystack/Stripe)
5. Webhook confirms payment
6. Balance updated

### 6. API Keys (`/sp/api-keys`)
**Purpose**: Manage API access

**List**:
- Key name/description
- Prefix (for identification)
- Created at
- Last used
- Status (active/revoked)
- Actions: View, Revoke

**Create Key**:
- Name/description
- Rate limit
- IP whitelist (optional)

### 7. Analytics (`/sp/analytics`)
**Purpose**: SP's personal analytics

**Charts**:
- Messages by day
- Spend by day
- Delivery rate trend
- Messages by MNO/carrier
- OTP success rate

**Reports**:
- Daily summary PDF
- Monthly invoice PDF
- Custom date range export

### 8. Settings (`/sp/settings`)
**Purpose**: SP account settings

**Sections**:
- **Profile**: Company info, contact details
- **Notifications**: Email/SMS alerts, low balance alerts
- **Security**: Password change, 2FA
- **Team**: Invite team members (if multi-user)
- **Billing**: Invoice details, payment methods

## Drizzle ORM Schema

```typescript
// drizzle/schema.ts
import { pgTable, serial, varchar, text, timestamp, boolean, integer, decimal, jsonb, uuid, pgEnum } from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';

// Enums
export const statusEnum = pgEnum('status', ['pending', 'active', 'suspended', 'cancelled', 'inactive']);
export const protocolEnum = pgEnum('protocol', ['smpp', 'http']);
export const messageTypeEnum = pgEnum('message_type', ['sms', 'otp', 'flash', 'unicode', 'binary']);
export const walletTypeEnum = pgEnum('wallet_type', ['prepaid', 'postpaid', 'mixed']);

// Service Providers
export const serviceProviders = pgTable('aegisbox_service_providers', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 255 }).notNull(),
  email: varchar('email', { length: 255 }).notNull().unique(),
  status: statusEnum('status').default('pending'),
  defaultCurrencyCode: varchar('default_currency_code', { length: 3 }).default('NGN'),
  phone: varchar('phone', { length: 20 }),
  address: text('address'),
  logoUrl: varchar('logo_url', { length: 500 }),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// SP Credentials
export const spCredentials = pgTable('aegisbox_sp_credentials', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id').references(() => serviceProviders.id).notNull(),
  protocol: protocolEnum('protocol').default('smpp'),
  status: statusEnum('status').default('active'),
  systemId: varchar('system_id', { length: 16 }).unique(),
  passwordHash: varchar('password_hash', { length: 255 }),
  bindType: varchar('bind_type', { length: 10 }).default('trx'),
  apiKeyHash: varchar('api_key_hash', { length: 255 }),
  apiKeyIdentifier: varchar('api_key_identifier', { length: 64 }).unique(),
  rateLimitPerSecond: integer('rate_limit_per_second').default(100),
  routingGroupId: integer('routing_group_id'),
  webhookUrl: varchar('webhook_url', { length: 500 }),
  lastLoginAt: timestamp('last_login_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Wallets
export const wallets = pgTable('aegisbox_wallets', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id').references(() => serviceProviders.id).notNull(),
  balance: decimal('balance', { precision: 15, scale: 4 }).default(0).notNull(),
  currencyCode: varchar('currency_code', { length: 3 }).default('NGN'),
  walletType: walletTypeEnum('wallet_type').default('prepaid'),
  creditLimit: decimal('credit_limit', { precision: 15, scale: 4 }).default(0),
  autoRechargeEnabled: boolean('auto_recharge_enabled').default(false),
  lowBalanceThreshold: decimal('low_balance_threshold', { precision: 15, scale: 4 }).default(50),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Wallet Transactions
export const walletTransactions = pgTable('aegisbox_wallet_transactions', {
  id: serial('id').primaryKey(),
  walletId: integer('wallet_id').references(() => wallets.id).notNull(),
  transactionType: varchar('transaction_type', { length: 30 }).notNull(),
  amount: decimal('amount', { precision: 15, scale: 4 }).notNull(),
  balanceBefore: decimal('balance_before', { precision: 15, scale: 4 }).notNull(),
  balanceAfter: decimal('balance_after', { precision: 15, scale: 4 }).notNull(),
  description: text('description'),
  referenceId: varchar('reference_id', { length: 100 }),
  referenceType: varchar('reference_type', { length: 50 }),
  metadata: jsonb('metadata'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
});

// Pricing Groups
export const pricingGroups = pgTable('aegisbox_pricing_groups', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 100 }).notNull(),
  description: text('description'),
  currencyCode: varchar('currency_code', { length: 3 }).default('NGN'),
  status: statusEnum('status').default('active'),
  isDefault: boolean('is_default').default(false),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Pricing Rules
export const pricingRules = pgTable('aegisbox_pricing_rules', {
  id: serial('id').primaryKey(),
  pricingGroupId: integer('pricing_group_id').references(() => pricingGroups.id).notNull(),
  mnoId: integer('mno_id').references(() => mnos.id),
  mnoCountryCode: varchar('mno_country_code', { length: 5 }),
  messageType: messageTypeEnum('message_type').default('sms'),
  pricePerSegment: decimal('price_per_segment', { precision: 10, scale: 4 }).notNull(),
  currencyCode: varchar('currency_code', { length: 3 }).default('NGN'),
  segmentLength: integer('segment_length').default(160),
  status: statusEnum('status').default('active'),
  effectiveFrom: timestamp('effective_from'),
  effectiveUntil: timestamp('effective_until'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// MNOs
export const mnos = pgTable('aegisbox_mnos', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 100 }).notNull().unique(),
  countryCode: varchar('country_code', { length: 5 }).notNull(),
  networkCode: varchar('network_code', { length: 10 }),
  status: statusEnum('status').default('active'),
  logoUrl: varchar('logo_url', { length: 500 }),
  metadata: jsonb('metadata'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// MNO Connections
export const mnoConnections = pgTable('aegisbox_mno_connections', {
  id: serial('id').primaryKey(),
  mnoId: integer('mno_id').references(() => mnos.id).notNull(),
  protocol: protocolEnum('protocol').default('smpp'),
  status: statusEnum('status').default('active'),
  priority: integer('priority').default(1),
  
  // SMPP
  systemId: varchar('system_id', { length: 16 }),
  password: varchar('password', { length: 255 }),
  host: varchar('host', { length: 255 }),
  port: integer('port'),
  useTls: boolean('use_tls').default(false),
  bindType: varchar('bind_type', { length: 10 }).default('trx'),
  systemType: varchar('system_type', { length: 13 }),
  
  // HTTP
  apiKey: text('api_key'),
  baseUrl: varchar('base_url', { length: 255 }).default(''),
  username: varchar('username', { length: 255 }),
  httpPassword: text('http_password'),
  timeoutSecs: integer('timeout_secs').default(30),
  supportsWebhook: boolean('supports_webhook').default(true),
  webhookPath: varchar('webhook_path', { length: 255 }),
  
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Messages
export const messages = pgTable('aegisbox_messages', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id').references(() => serviceProviders.id).notNull(),
  spCredentialId: integer('sp_credential_id').references(() => spCredentials.id),
  walletTransactionId: bigint('wallet_transaction_id'),
  
  externalMessageId: varchar('external_message_id', { length: 100 }).unique(),
  internalMessageId: uuid('internal_message_id').defaultRandom(),
  
  senderId: varchar('sender_id', { length: 16 }).notNull(),
  destinationMsisdn: varchar('destination_msisdn', { length: 20 }).notNull(),
  countryCode: varchar('country_code', { length: 5 }),
  
  messageContent: text('message_content').notNull(),
  messageType: messageTypeEnum('message_type').default('sms'),
  isOtp: boolean('is_otp').default(false),
  
  mnoId: integer('mno_id').references(() => mnos.id),
  mnoConnectionId: integer('mno_connection_id').references(() => mnoConnections.id),
  routingGroupId: integer('routing_group_id').references(() => routingGroups.id),
  
  status: varchar('status', { length: 30 }).default('pending'),
  statusReason: varchar('status_reason', { length: 200 }),
  
  submittedAt: timestamp('submitted_at').defaultNow(),
  sentAt: timestamp('sent_at'),
  deliveredAt: timestamp('delivered_at'),
  
  costPerSegment: decimal('cost_per_segment', { precision: 10, scale: 4 }).default(0),
  totalCost: decimal('total_cost', { precision: 15, scale: 4 }).default(0),
  currencyCode: varchar('currency_code', { length: 3 }).default('NGN'),
  
  errorCode: varchar('error_code', { length: 50 }),
  errorDescription: text('error_description'),
  
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Relations (define all)
export const serviceProviderRelations = relations(serviceProviders, ({ many }) => ({
  credentials: many(spCredentials),
  wallets: many(wallets),
  messages: many(messages),
}));

export const walletRelations = relations(wallets, ({ one, many }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [wallets.serviceProviderId],
    references: [serviceProviders.id],
  }),
  transactions: many(walletTransactions),
}));

// ... define all other relations
```

## API Route Structure

```
/api/admin/
├── dashboard/
│   ├── stats
│   ├── charts/
│   │   ├── revenue
│   │   └── volume
│   ├── top-sp
│   └── recent-transactions
├── service-providers/
│   ├── GET (list)
│   ├── POST (create)
│   └── [id]/
│       ├── GET (detail)
│       ├── PUT (update)
│       ├── DELETE
│       ├── suspend
│       ├── activate
│       ├── wallet
│       └── credentials
├── mnos/
│   ├── GET (list)
│   ├── POST (create)
│   └── [id]/
│       ├── GET
│       ├── PUT
│       ├── DELETE
│       └── connections/
│           ├── GET
│           ├── POST
│           └── [connId]/
├── pricing/
│   ├── GET (groups)
│   ├── POST (group)
│   └── [groupId]/
│       ├── GET
│       ├── PUT
│       ├── DELETE
│       └── rules/
│           ├── GET
│           ├── POST
│           └── [ruleId]/
├── routing/
│   ├── GET (groups)
│   ├── POST (group)
│   └── [groupId]/
│       ├── GET
│       ├── PUT
│       ├── DELETE
│       └── rules/
│           ├── GET
│           ├── POST
│           └── [ruleId]/
├── transactions/
│   ├── GET (list with filters)
│   ├── GET [id]
│   ├── export
│   ├── [id]/refund
│   └── [id]/resend
├── wallets/
│   ├── GET (list)
│   └── [id]/
│       ├── GET
│       ├── adjust
│       └── transactions
├── analytics/
│   ├── overview
│   ├── sp-performance
│   ├── mno-performance
│   ├── revenue
│   └── export
└── settings/
    └── GET/PUT

/api/sp/
├── dashboard/stats
├── compose/
│   ├── validate-senders
│   └── send
├── messages/
│   ├── GET (list)
│   ├── GET [id]
│   └── export
├── wallet/
│   ├── GET (balance)
│   ├── transactions
│   ├── credit
│   └── auto-recharge
├── senders/
│   ├── GET
│   └── request
├── api-keys/
│   ├── GET
│   ├── POST
│   └── [id]/revoke
└── analytics/
    ├── overview
    └── export
```

## UI Component Guidelines

### Use shadcn/ui Components
- **Cards**: Stats overview, summary panels
- **Tables**: Transaction lists, paginated data
- **Forms**: All input forms with React Hook Form + Zod
- **Modals/Dialogs**: Confirmations, quick edits, details views
- **Tabs**: Navigate within detail pages
- **Select**: Dropdowns for MNO, status, etc.
- **Input/Textarea**: Form fields
- **Button**: All actions
- **Badge**: Status indicators
- **Avatar**: User profiles
- **DropdownMenu**: Row actions
- **Toast**: Notifications
- **Progress**: OTP sending progress, wallet loading

### Color Scheme
```typescript
// tailwind.config.ts
colors: {
  primary: {
    50: '#f0f9ff',
    100: '#e0f2fe',
    500: '#0ea5e9',  // Main brand color
    600: '#0284c7',
    700: '#0369a1',
  },
  success: '#22c55e',
  warning: '#f59e0b',
  error: '#ef4444',
  info: '#3b82f6',
}
```

### Status Badge Colors
```typescript
const statusColors: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800',
  active: 'bg-green-100 text-green-800',
  delivered: 'bg-green-100 text-green-800',
  sent: 'bg-blue-100 text-blue-800',
  failed: 'bg-red-100 text-red-800',
  suspended: 'bg-red-100 text-red-800',
  inactive: 'bg-gray-100 text-gray-800',
};
```

## Key UI Flows

### Sending an SMS (SP Portal)
```
1. User clicks "Send SMS" or visits /sp/compose
2. Form loads with:
   - Sender ID dropdown (pre-populated with approved senders)
   - Recipients input (comma/newline/file)
   - Message textarea with character counter
   - Message type selector (SMS/OTP)
   - OTP template selector (if OTP selected)
   - Schedule toggle
3. User fills form
4. Clicks "Send"
5. System validates:
   - Balance sufficient (prepaid)
   - Sender ID approved
   - Recipients valid format
   - Within rate limits
6. If valid:
   - Creates message record
   - Deducts wallet (prepaid)
   - Routes to MNO
   - Shows success toast
7. If invalid:
   - Shows specific error message
```

### Adding an MNO Connection (Admin)
```
1. Admin visits /admin/mnos/[id]/connections
2. Clicks "Add Connection"
3. Selects protocol (SMPP or HTTP)
4. Fills connection details:
   - SMPP: host, port, system_id, password, bind_type, timeouts
   - HTTP: base_url, api_key, timeout, webhook settings
5. Sets priority
6. Clicks "Test Connection"
7. System attempts connection test
8. If successful:
   - Shows success
   - Saves connection as "active"
9. If failed:
   - Shows error details
   - User can retry or save as "inactive"
```

### Configuring Routing (Admin)
```
1. Admin visits /admin/routing
2. Creates or selects routing group
3. Clicks "Add Rule"
4. Selects:
   - MSISDN Prefix Group (e.g., "MTN Nigeria")
   - MNO (e.g., "MTN Nigeria")
   - Priority (for multiple rules)
   - Failover MNO (optional)
5. Rule saved
6. Messages to matching prefixes now route to selected MNO
```

## Implementation Instructions

### 1. Set Up Next.js Project
```bash
npx create-next-app@latest aegisbox-frontend \
  --typescript \
  --tailwind \
  --eslint \
  --app \
  --src-dir \
  --import-alias "@/*" \
  --use-npm

cd aegisbox-frontend
npm install @clerk/nextjs drizzle-orm @libsql/client
npm install -D drizzle-kit
npm install lucide-react recharts date-fns react-hook-form zod @tanstack/react-query zustand
npm install -D @types/node
```

### 2. Configure shadcn/ui
```bash
npx shadcn@latest init
npx shadcn@latest add button card table input textarea select dialog tabs badge toast dropdown-menu avatar progress skeleton tooltip alert
```

### 3. Set Up Drizzle
```bash
npm install drizzle-orm @types/pg
npm install -D drizzle-kit
```

Create `drizzle.config.ts`:
```typescript
import type { Config } from 'drizzle-kit';

export default {
  schema: './src/db/schema.ts',
  out: './drizzle',
  dialect: 'postgresql',
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
} satisfies Config;
```

### 4. Set Up Clerk
```bash
npm install @clerk/nextjs
```

Add to `middleware.ts`:
```typescript
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server';

const isAdminRoute = createRouteMatcher(['/admin(.*)']);
const isSPRoute = createRouteMatcher(['/sp(.*)']);

export default clerkMiddleware(async (auth, req) => {
  const { userId, redirectToSignIn } = await auth();
  
  if (isAdminRoute(req) || isSPRoute(req)) {
    if (!userId) return redirectToSignIn();
  }
});

export const config = {
  matcher: ['/((?!.*\\..*|_next).*)', '/', '/(api|trpc)(.*)'],
};
```

### 5. Environment Variables
Create `.env.local`:
```env
# Clerk
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL=/dashboard
NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL=/dashboard

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/aegisbox

# App
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## Testing Checklist

### Admin Portal Tests
- [ ] Admin can log in with Clerk
- [ ] Admin can create service provider
- [ ] Admin can edit service provider details
- [ ] Admin can suspend/activate service provider
- [ ] Admin can view all transactions
- [ ] Admin can refund transaction
- [ ] Admin can adjust wallet balance
- [ ] Admin can create MNO connection (SMPP)
- [ ] Admin can create MNO connection (HTTP)
- [ ] Admin can test MNO connection
- [ ] Admin can configure routing rules
- [ ] Admin can create pricing groups
- [ ] Admin can add pricing rules
- [ ] Analytics dashboard loads correctly
- [ ] Export works for all reports

### Service Provider Portal Tests
- [ ] SP user can log in with Clerk
- [ ] SP user sees own dashboard data
- [ ] SP user can send SMS
- [ ] SP user can send OTP
- [ ] SP user can request new sender ID
- [ ] SP user can view wallet balance
- [ ] SP user can buy credit
- [ ] SP user can view transaction history
- [ ] SP user can export reports
- [ ] SP user cannot access admin routes
- [ ] SP user sees only own data (row-level security)

### Integration Tests
- [ ] Clerk webhook syncs users correctly
- [ ] Database queries work with filters
- [ ] Pagination works correctly
- [ ] API rate limiting works
- [ ] Webhook endpoints receive and process callbacks

## Expected Deliverables

1. **Complete Next.js Application** with:
   - Admin Portal (/admin/*)
   - Service Provider Portal (/sp/*)
   - Authentication (/sign-in, /sign-up)
   - Landing page

2. **Database Schema** with:
   - Drizzle ORM models
   - Migration files
   - Seed data (example MNOs, pricing groups)

3. **API Routes** with:
   - Admin endpoints
   - SP endpoints
   - Clerk webhook handler

4. **UI Components** with:
   - shadcn/ui components
   - Custom components for charts, tables
   - Responsive design

5. **Tests**:
   - Unit tests for utils
   - Integration tests for key flows

## Notes for AI Developer

1. **Start Small**: Begin with the database schema and authentication, then build features incrementally.

2. **Use Server Actions**: Prefer Next.js Server Actions over API routes for type safety.

3. **Row-Level Security**: Always filter queries by `service_provider_id` for SP portal routes.

4. **Error Handling**: Use try/catch with proper error messages and toast notifications.

5. **Loading States**: Use Suspense and loading skeletons for good UX.

6. **Pagination**: Implement cursor-based pagination for large datasets.

7. **Search/Filter**: Use debounced search inputs.

8. **Optimistic Updates**: Use for actions like wallet credits.

9. **Webhooks**: Test webhook endpoints with Clerk dashboard.

10. **Security**: Never expose sensitive data (passwords, API keys) in API responses.

11. **TypeScript**: Use strict typing everywhere, especially for API responses.

12. **Accessibility**: Ensure all interactive elements have proper ARIA attributes.

13. **Responsive**: Mobile-first approach, test on various screen sizes.

14. **Performance**: Use React Query for server state, implement virtualization for large tables.

15. **Internationalization**: Use i18n if supporting multiple languages (phone number formatting, currency).

This prompt should give you everything needed to build a complete, production-ready SMS Gateway admin and customer portal. Good luck!
