-- +goose Up
-- +goose StatementBegin

-- =========================================
-- INTELLIGENT ROUTING SYSTEM MIGRATION
-- =========================================

-- Phase 1: Modify routing_assignments to support multiple MNOs per prefix group
-- Remove unique constraint that prevents multiple MNOs
ALTER TABLE routing_assignments 
DROP CONSTRAINT IF EXISTS routing_assignments_routing_group_id_msisdn_prefix_group__key;

-- Add new fields for intelligent routing
ALTER TABLE routing_assignments
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS weight INT DEFAULT 100 CHECK (weight >= 0 AND weight <= 100),
ADD COLUMN IF NOT EXISTS max_retry_attempts INT DEFAULT 2;

-- Create new index for efficient active route lookups
CREATE INDEX IF NOT EXISTS idx_routing_assignments_active_lookup 
ON routing_assignments(routing_group_id, msisdn_prefix_group_id, is_active, status, priority);

COMMENT ON COLUMN routing_assignments.is_active IS 'Quick toggle for route without changing status. Allows temporary disable for maintenance.';
COMMENT ON COLUMN routing_assignments.weight IS 'Load balancing weight (0-100). Higher weight gets more traffic within same priority.';
COMMENT ON COLUMN routing_assignments.max_retry_attempts IS 'Max retry attempts for this specific route before trying next route.';

-- Phase 2: Track which routing assignment was used for each message
ALTER TABLE messages
ADD COLUMN IF NOT EXISTS routing_assignment_id INT REFERENCES routing_assignments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_messages_routing_assignment 
ON messages(routing_assignment_id);

ALTER TABLE message_segments
ADD COLUMN IF NOT EXISTS routing_assignment_id INT REFERENCES routing_assignments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_message_segments_routing_assignment
ON message_segments(routing_assignment_id);

COMMENT ON COLUMN messages.routing_assignment_id IS 'Tracks which route was selected for this message (for health feedback loop)';
COMMENT ON COLUMN message_segments.routing_assignment_id IS 'Tracks which route was used for this segment';

-- Phase 3: Ring Buffer for Sender ID Delivery Attempts
CREATE TABLE IF NOT EXISTS sender_id_delivery_attempts (
    id BIGSERIAL PRIMARY KEY,
    sender_id INT NOT NULL REFERENCES sender_ids(id) ON DELETE CASCADE,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE CASCADE,
    message_segment_id BIGINT REFERENCES message_segments(id) ON DELETE CASCADE,
    
    -- Outcome
    status VARCHAR(20) NOT NULL CHECK (status IN ('delivered', 'failed', 'no_dlr')),
    error_code VARCHAR(10),
    
    -- Timestamps
    sent_at TIMESTAMPTZ NOT NULL,
    dlr_received_at TIMESTAMPTZ,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Partition key for efficient cleanup
    recorded_date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE INDEX idx_sender_delivery_attempts_sender_mno_date 
ON sender_id_delivery_attempts(sender_id, mno_id, recorded_at DESC);

CREATE INDEX idx_sender_delivery_attempts_recorded 
ON sender_id_delivery_attempts(recorded_at DESC);

-- Index for cleanup (without WHERE clause to avoid IMMUTABLE function requirement)
CREATE INDEX idx_sender_delivery_attempts_date 
ON sender_id_delivery_attempts(recorded_date);

CREATE INDEX idx_sender_delivery_attempts_segment
ON sender_id_delivery_attempts(message_segment_id);

COMMENT ON TABLE sender_id_delivery_attempts IS 'Ring buffer storing individual delivery attempts for accurate health metrics (last 1000 per sender-MNO)';
COMMENT ON COLUMN sender_id_delivery_attempts.status IS 'delivered: got success DLR, failed: got failure DLR, no_dlr: timeout without DLR';

-- Phase 4: Sender ID Health Summary Table
CREATE TABLE IF NOT EXISTS sender_id_health (
    id BIGSERIAL PRIMARY KEY,
    sender_id INT NOT NULL REFERENCES sender_ids(id) ON DELETE CASCADE,
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE CASCADE,
    
    -- Calculated metrics from ring buffer (last N attempts)
    total_sent INT NOT NULL DEFAULT 0,
    total_delivered INT NOT NULL DEFAULT 0,
    total_failed INT NOT NULL DEFAULT 0,
    total_no_dlr INT NOT NULL DEFAULT 0,
    
    -- Calculated percentages (0-100)
    delivery_rate NUMERIC(5,2) DEFAULT 100.00,
    dlr_response_rate NUMERIC(5,2) DEFAULT 100.00,
    health_score NUMERIC(5,2) DEFAULT 100.00,
    
    -- Blocking management
    is_blocked BOOLEAN DEFAULT false,
    blocked_at TIMESTAMPTZ,
    blocked_reason TEXT,
    auto_blocked BOOLEAN DEFAULT false,
    
    -- Auto-unblock tracking
    last_unblock_check_at TIMESTAMPTZ,
    
    -- Activity timestamps
    last_success_at TIMESTAMPTZ,
    last_failure_at TIMESTAMPTZ,
    last_no_dlr_at TIMESTAMPTZ,
    last_attempt_at TIMESTAMPTZ,
    
    -- Window configuration
    window_size INT DEFAULT 1000,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (sender_id, mno_id)
);

CREATE INDEX idx_sender_id_health_sender_mno ON sender_id_health(sender_id, mno_id);
CREATE INDEX idx_sender_id_health_blocked ON sender_id_health(is_blocked, mno_id);
CREATE INDEX idx_sender_id_health_score ON sender_id_health(health_score DESC);
CREATE INDEX idx_sender_id_health_auto_unblock ON sender_id_health(is_blocked, last_attempt_at) WHERE is_blocked = true;

COMMENT ON TABLE sender_id_health IS 'Aggregated sender ID delivery performance per MNO using rolling window from ring buffer';
COMMENT ON COLUMN sender_id_health.health_score IS 'Weighted score: 70% delivery_rate + 30% dlr_response_rate';
COMMENT ON COLUMN sender_id_health.auto_blocked IS 'True if blocked by health thresholds, false if manually blocked by admin';
COMMENT ON COLUMN sender_id_health.last_attempt_at IS 'Last delivery attempt time (for 7-day auto-unblock check)';

-- Phase 5: Ring Buffer for Route Delivery Attempts
CREATE TABLE IF NOT EXISTS route_delivery_attempts (
    id BIGSERIAL PRIMARY KEY,
    routing_assignment_id INT NOT NULL REFERENCES routing_assignments(id) ON DELETE CASCADE,
    sender_id INT REFERENCES sender_ids(id) ON DELETE SET NULL,
    message_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
    
    -- Outcome
    success BOOLEAN NOT NULL,
    error_code VARCHAR(10),
    error_description TEXT,
    
    -- Timestamps
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    recorded_date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE INDEX idx_route_delivery_attempts_assignment_date 
ON route_delivery_attempts(routing_assignment_id, recorded_date DESC);

CREATE INDEX idx_route_delivery_attempts_sender
ON route_delivery_attempts(sender_id, recorded_date DESC);

-- Index for cleanup (without WHERE clause to avoid IMMUTABLE function requirement)
CREATE INDEX idx_route_delivery_attempts_date 
ON route_delivery_attempts(recorded_date);

COMMENT ON TABLE route_delivery_attempts IS 'Ring buffer for route-level delivery attempts (connection availability tracking)';

-- Phase 6: Route Health Summary Table
CREATE TABLE IF NOT EXISTS route_health (
    id BIGSERIAL PRIMARY KEY,
    routing_assignment_id INT NOT NULL REFERENCES routing_assignments(id) ON DELETE CASCADE,
    sender_id INT REFERENCES sender_ids(id) ON DELETE CASCADE,
    
    -- Calculated metrics from ring buffer
    total_attempts INT NOT NULL DEFAULT 0,
    successful_attempts INT NOT NULL DEFAULT 0,
    failed_attempts INT NOT NULL DEFAULT 0,
    
    -- Calculated percentages
    success_rate NUMERIC(5,2) DEFAULT 100.00,
    health_score NUMERIC(5,2) DEFAULT 100.00,
    
    -- Availability tracking
    is_available BOOLEAN DEFAULT true,
    unavailable_since TIMESTAMPTZ,
    unavailable_reason TEXT,
    auto_unavailable BOOLEAN DEFAULT false,
    
    -- Consecutive tracking for threshold detection
    consecutive_failures INT DEFAULT 0,
    consecutive_successes INT DEFAULT 0,
    
    -- Activity timestamps
    last_attempt_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    last_failure_at TIMESTAMPTZ,
    
    -- Window configuration
    window_size INT DEFAULT 1000,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (routing_assignment_id, sender_id)
);

CREATE INDEX idx_route_health_assignment ON route_health(routing_assignment_id);
CREATE INDEX idx_route_health_sender ON route_health(sender_id);
CREATE INDEX idx_route_health_available ON route_health(is_available, health_score DESC);
CREATE INDEX idx_route_health_last_attempt ON route_health(last_attempt_at);

COMMENT ON TABLE route_health IS 'Route availability and health tracking combining MNO connection status with sender-specific performance';
COMMENT ON COLUMN route_health.sender_id IS 'NULL = route-level health (all senders), non-NULL = sender-specific route health';

-- Phase 7: Routing Configuration Table
CREATE TABLE IF NOT EXISTS routing_config (
    id SERIAL PRIMARY KEY,
    routing_group_id INT UNIQUE REFERENCES routing_groups(id) ON DELETE CASCADE,
    
    -- Sender ID health thresholds
    sender_id_block_threshold NUMERIC(5,2) DEFAULT 50.00 CHECK (sender_id_block_threshold >= 0 AND sender_id_block_threshold <= 100),
    sender_id_unblock_threshold NUMERIC(5,2) DEFAULT 70.00 CHECK (sender_id_unblock_threshold >= 0 AND sender_id_unblock_threshold <= 100),
    sender_id_dlr_timeout_secs INT DEFAULT 300,
    sender_id_health_window_size INT DEFAULT 1000,
    
    -- Route health thresholds
    route_availability_threshold NUMERIC(5,2) DEFAULT 60.00 CHECK (route_availability_threshold >= 0 AND route_availability_threshold <= 100),
    route_failure_threshold INT DEFAULT 5,
    route_recovery_threshold INT DEFAULT 3,
    route_health_window_size INT DEFAULT 1000,
    
    -- Failover & retry configuration
    enable_automatic_failover BOOLEAN DEFAULT true,
    max_route_retries INT DEFAULT 2,
    
    -- Route selection strategy
    selection_strategy VARCHAR(20) DEFAULT 'priority_weighted' CHECK (selection_strategy IN ('priority_weighted', 'health_based', 'round_robin')),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_routing_config_group ON routing_config(routing_group_id);

COMMENT ON TABLE routing_config IS 'Configuration for intelligent routing behavior. NULL routing_group_id = global defaults';
COMMENT ON COLUMN routing_config.sender_id_block_threshold IS 'Auto-block sender when delivery_rate falls below this percentage';
COMMENT ON COLUMN routing_config.sender_id_unblock_threshold IS 'Consider unblocking when delivery_rate rises above this (manual unblock only)';
COMMENT ON COLUMN routing_config.sender_id_dlr_timeout_secs IS 'Seconds to wait for DLR before marking as no_dlr';
COMMENT ON COLUMN routing_config.route_failure_threshold IS 'Consecutive failures before marking route unavailable';
COMMENT ON COLUMN routing_config.route_recovery_threshold IS 'Consecutive successes to mark route available again';
COMMENT ON COLUMN routing_config.max_route_retries IS 'Number of alternate routes to try on failure (0 = no retry, 2 = try 3 routes total)';
COMMENT ON COLUMN routing_config.selection_strategy IS 'priority_weighted: priority first then weight, health_based: best health first, round_robin: distribute evenly';

-- Insert global default configuration
INSERT INTO routing_config (
    routing_group_id, 
    sender_id_block_threshold, 
    sender_id_unblock_threshold,
    sender_id_dlr_timeout_secs, 
    sender_id_health_window_size,
    route_availability_threshold,
    route_failure_threshold,
    route_recovery_threshold,
    max_route_retries, 
    selection_strategy
)
VALUES (
    NULL,  -- Global default
    50.00, -- Block below 50% delivery rate
    70.00, -- Unblock threshold
    300,   -- 5 minutes DLR timeout
    1000,  -- Window size
    60.00, -- Route availability threshold
    5,     -- 5 consecutive failures
    3,     -- 3 consecutive successes
    2,     -- Try up to 3 routes total (original + 2 retries)
    'priority_weighted'
)
ON CONFLICT (routing_group_id) DO NOTHING;

-- Phase 8: Ring Buffer Cleanup Configuration
CREATE TABLE IF NOT EXISTS health_ring_buffer_config (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(100) NOT NULL UNIQUE,
    retention_days INT NOT NULL DEFAULT 90,
    cleanup_enabled BOOLEAN DEFAULT true,
    last_cleanup_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO health_ring_buffer_config (table_name, retention_days) VALUES
('sender_id_delivery_attempts', 90),
('route_delivery_attempts', 90)
ON CONFLICT (table_name) DO NOTHING;

COMMENT ON TABLE health_ring_buffer_config IS 'Configuration for automatic cleanup of ring buffer tables';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop in reverse order to respect foreign key constraints
DROP TABLE IF EXISTS health_ring_buffer_config;
DROP TABLE IF EXISTS routing_config;
DROP TABLE IF EXISTS route_health;
DROP TABLE IF EXISTS route_delivery_attempts;
DROP TABLE IF EXISTS sender_id_health;
DROP TABLE IF EXISTS sender_id_delivery_attempts;

-- Remove added columns
ALTER TABLE message_segments DROP COLUMN IF EXISTS routing_assignment_id;
ALTER TABLE messages DROP COLUMN IF EXISTS routing_assignment_id;

-- Remove added columns from routing_assignments
ALTER TABLE routing_assignments DROP COLUMN IF EXISTS max_retry_attempts;
ALTER TABLE routing_assignments DROP COLUMN IF EXISTS weight;
ALTER TABLE routing_assignments DROP COLUMN IF EXISTS is_active;

-- Restore unique constraint
ALTER TABLE routing_assignments 
ADD CONSTRAINT routing_assignments_routing_group_id_msisdn_prefix_group__key 
UNIQUE (routing_group_id, msisdn_prefix_group_id);

-- +goose StatementEnd
