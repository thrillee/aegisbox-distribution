-- Migration 002: Add intelligent routing health tracking tables

-- Route Health Tracking
CREATE TABLE IF NOT EXISTS route_health (
    assignment_id INT NOT NULL PRIMARY KEY REFERENCES routing_assignments(id) ON DELETE CASCADE,
    health_score DECIMAL(5, 4) DEFAULT '1.0000',
    success_count INT DEFAULT 0 NOT NULL,
    fail_count INT DEFAULT 0 NOT NULL,
    total_attempts INT DEFAULT 0 NOT NULL,
    consecutive_failures INT DEFAULT 0 NOT NULL,
    consecutive_successes INT DEFAULT 0 NOT NULL,
    is_available BOOLEAN DEFAULT true NOT NULL,
    last_attempt_at TIMESTAMPTZ,
    marked_unavailable_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Sender ID Health Tracking
CREATE TABLE IF NOT EXISTS sender_id_health (
    id SERIAL PRIMARY KEY,
    sender_id VARCHAR(16) NOT NULL,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE CASCADE,
    delivery_rate DECIMAL(5, 4) DEFAULT '1.0000',
    dlr_response_rate DECIMAL(5, 4) DEFAULT '1.0000',
    health_score DECIMAL(5, 4) DEFAULT '1.0000',
    total_messages INT DEFAULT 0 NOT NULL,
    is_blocked BOOLEAN DEFAULT false NOT NULL,
    blocked_at TIMESTAMPTZ,
    blocked_reason TEXT,
    last_message_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Route Delivery Attempts (Ring Buffer for Health Calculations)
CREATE TABLE IF NOT EXISTS route_delivery_attempts (
    id SERIAL PRIMARY KEY,
    assignment_id INT NOT NULL REFERENCES routing_assignments(id) ON DELETE CASCADE,
    success BOOLEAN NOT NULL,
    attempted_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Sender ID Delivery Attempts
CREATE TABLE IF NOT EXISTS sender_id_delivery_attempts (
    id SERIAL PRIMARY KEY,
    sender_id VARCHAR(16) NOT NULL,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE CASCADE,
    dlr_received BOOLEAN NOT NULL,
    attempted_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Routing Config Table
CREATE TABLE IF NOT EXISTS routing_config (
    id SERIAL PRIMARY KEY,
    routing_group_id INT NOT NULL UNIQUE REFERENCES routing_groups(id) ON DELETE CASCADE,
    selection_strategy VARCHAR(50) DEFAULT 'priority_weighted' NOT NULL,
    max_route_retries INT DEFAULT 2 NOT NULL,
    dlr_timeout_seconds INT DEFAULT 300 NOT NULL,
    sender_block_threshold DECIMAL(3, 2) DEFAULT '0.50' NOT NULL,
    sender_unblock_threshold DECIMAL(3, 2) DEFAULT '0.70' NOT NULL,
    route_availability_threshold DECIMAL(3, 2) DEFAULT '0.70' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Health Ring Buffer Config Table
CREATE TABLE IF NOT EXISTS health_ring_buffer_config (
    id SERIAL PRIMARY KEY,
    health_calculation_window INT DEFAULT 1000 NOT NULL,
    route_failure_threshold INT DEFAULT 5 NOT NULL,
    route_recovery_threshold INT DEFAULT 3 NOT NULL,
    sender_auto_unblock_check_interval_minutes INT DEFAULT 30 NOT NULL,
    dlr_timeout_check_interval_minutes INT DEFAULT 5 NOT NULL,
    ring_buffer_cleanup_interval_hours INT DEFAULT 24 NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Add missing columns to routing_assignments
DO $$ BEGIN
    ALTER TABLE routing_assignments ADD COLUMN IF NOT EXISTS weight INT DEFAULT 100 NOT NULL;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

DO $$ BEGIN
    ALTER TABLE routing_assignments ADD COLUMN IF NOT EXISTS max_retry_attempts INT DEFAULT 2 NOT NULL;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

DO $$ BEGIN
    ALTER TABLE routing_assignments ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT true NOT NULL;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

DO $$ BEGIN
    ALTER TABLE routing_assignments ADD COLUMN IF NOT EXISTS consecutive_failures INT DEFAULT 0 NOT NULL;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

-- Create indexes for health tracking
CREATE INDEX IF NOT EXISTS idx_route_health_assignment ON route_health(assignment_id);
CREATE INDEX IF NOT EXISTS idx_route_health_consecutive_failures ON route_health(consecutive_failures);
CREATE INDEX IF NOT EXISTS idx_route_health_is_available ON route_health(is_available);
CREATE INDEX IF NOT EXISTS idx_sender_id_health_sender_id ON sender_id_health(sender_id);
CREATE INDEX IF NOT EXISTS idx_sender_id_health_mno_id ON sender_id_health(mno_id);
CREATE INDEX IF NOT EXISTS idx_sender_id_health_is_blocked ON sender_id_health(is_blocked);
CREATE INDEX IF NOT EXISTS idx_route_delivery_attempts_assignment ON route_delivery_attempts(assignment_id);
CREATE INDEX IF NOT EXISTS idx_route_delivery_attempts_attempted_at ON route_delivery_attempts(attempted_at DESC);
CREATE INDEX IF NOT EXISTS idx_sender_id_delivery_attempts_mno_id ON sender_id_delivery_attempts(mno_id);
CREATE INDEX IF NOT EXISTS idx_sender_id_delivery_attempts_attempted_at ON sender_id_delivery_attempts(attempted_at);
