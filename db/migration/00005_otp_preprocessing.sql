-- +goose Up
-- +goose StatementBegin

-- Add 'scope' to sp_credentials
ALTER TABLE sp_credentials
ADD COLUMN scope VARCHAR(50) NOT NULL DEFAULT 'default'; -- e.g., default, otp, marketing

CREATE INDEX idx_sp_credentials_scope ON sp_credentials(scope);

-- Table for OTP Alternative Sender IDs
CREATE TABLE otp_alternative_senders (
    id SERIAL PRIMARY KEY,
    sender_id_string VARCHAR(16) NOT NULL UNIQUE, -- The alternative sender ID to use (e.g., "InfoSMS", "VerifyOTP")
    service_provider_id INT REFERENCES service_providers(id) ON DELETE CASCADE,
    mno_id INT REFERENCES mnos(id) ON DELETE SET NULL, -- Optional: If this sender is specific to an MNO
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'inactive', 'depleted_awaiting_reset'
    current_usage_count INT NOT NULL DEFAULT 0,
    max_usage_count INT NOT NULL DEFAULT 10000, -- Max messages before status might change or reset needed
    -- Optional: For automatic reset logic based on interval
    reset_interval_hours INT DEFAULT NULL, -- e.g., 24 for daily reset. NULL means manual or event-driven reset.
    last_reset_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_otp_alternative_senders_mno_id ON otp_alternative_senders(mno_id);
CREATE INDEX idx_otp_alternative_senders_status_usage ON otp_alternative_senders(status, current_usage_count);
CREATE INDEX idx_otp_alternative_senders_last_used ON otp_alternative_senders(last_used_at);
CREATE INDEX IF NOT EXISTS idx_otp_alt_senders_sp_mno_status_usage 
    ON otp_alternative_senders(service_provider_id, mno_id, status, current_usage_count, last_used_at);
CREATE INDEX IF NOT EXISTS idx_otp_alt_senders_sp_globalmno_status_usage 
    ON otp_alternative_senders(service_provider_id, status, current_usage_count, last_used_at) WHERE mno_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_otp_alt_senders_global_mno_status_usage 
    ON otp_alternative_senders(mno_id, status, current_usage_count, last_used_at) WHERE service_provider_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_otp_alt_senders_global_globalmno_status_usage 
    ON otp_alternative_senders(status, current_usage_count, last_used_at) WHERE service_provider_id IS NULL AND mno_id IS NULL;

-- Table for OTP Message Templates
CREATE TABLE otp_message_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE, -- For easy identification, e.g., "Standard OTP English v1"
    content_template TEXT NOT NULL,
        -- Example: "Your verification code for [APP_NAME] is [OTP]. Do not share it. This code is valid for [VALIDITY_MINUTES] minutes. [BRAND_NAME]"
        -- Placeholders: [OTP], [BRAND_NAME], [APP_NAME], [VALIDITY_MINUTES] (you define these)
    default_brand_name VARCHAR(100) DEFAULT NULL, -- Fallback brand name if not derived from SP
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'inactive'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_otp_message_templates_status ON otp_message_templates(status);

-- Table to Assign Templates to Alternative Senders (allows flexibility)
CREATE TABLE otp_sender_template_assignments (
    id SERIAL PRIMARY KEY,
    otp_alternative_sender_id INT NOT NULL REFERENCES otp_alternative_senders(id) ON DELETE CASCADE,
    otp_message_template_id INT NOT NULL REFERENCES otp_message_templates(id) ON DELETE CASCADE,
    mno_id INT REFERENCES mnos(id) ON DELETE SET NULL, -- Optional: Make template choice MNO-specific for this sender
    priority INT NOT NULL DEFAULT 0, -- Lower number = higher priority for selection
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'inactive'
    assignment_usage_count BIGINT NOT NULL DEFAULT 0, -- Tracks usage of this specific sender+template combo
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (otp_alternative_sender_id, otp_message_template_id, mno_id) -- Ensure unique assignment
);
CREATE INDEX idx_otp_assignments_sender_template ON otp_sender_template_assignments(otp_alternative_sender_id, otp_message_template_id);
CREATE INDEX idx_otp_assignments_mno_status_priority ON otp_sender_template_assignments(mno_id, status, priority);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS otp_sender_template_assignments;
DROP TABLE IF EXISTS otp_message_templates;
DROP TABLE IF EXISTS otp_alternative_senders;

ALTER TABLE sp_credentials
DROP COLUMN IF EXISTS scope;
-- DROP INDEX IF EXISTS idx_sp_credentials_scope; -- Usually dropped automatically with column

-- +goose StatementEnd
