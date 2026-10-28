-- +goose Up
-- +goose StatementBegin

-- =========================================
-- MANUAL AND AUTO DLR
-- =========================================

-- Per service provider auto DLR: when enabled, segments that get no DLR from the
-- MNO within auto_dlr_delay_seconds are completed with auto_dlr_status and the
-- DLR is forwarded to the SP as if it came from the MNO.
ALTER TABLE service_providers
ADD COLUMN auto_dlr_enabled BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN auto_dlr_status VARCHAR(20) NOT NULL DEFAULT 'DELIVRD',
ADD COLUMN auto_dlr_delay_seconds INT NOT NULL DEFAULT 300,
-- Only segments sent after this time are auto completed, so enabling auto DLR
-- does not fire DLRs for an old backlog
ADD COLUMN auto_dlr_enabled_at TIMESTAMPTZ;

ALTER TABLE service_providers
ADD CONSTRAINT chk_service_providers_auto_dlr_status
CHECK (auto_dlr_status IN ('DELIVRD', 'UNDELIV', 'EXPIRED', 'REJECTD', 'DELETED'));

ALTER TABLE service_providers
ADD CONSTRAINT chk_service_providers_auto_dlr_delay
CHECK (auto_dlr_delay_seconds BETWEEN 60 AND 172800);

-- 'manual' (admin) or 'auto' (auto DLR worker) when the DLR did not come from the MNO; NULL otherwise
ALTER TABLE message_segments
ADD COLUMN dlr_source VARCHAR(20);

CREATE INDEX idx_service_providers_auto_dlr_enabled ON service_providers(id) WHERE auto_dlr_enabled;

-- Segments accepted by the MNO that still have no final DLR. Used by the auto DLR worker;
-- the predicate must match the pending condition in ApplyManualSegmentDLR/GetAutoDLRCandidates.
CREATE INDEX idx_message_segments_awaiting_final_dlr ON message_segments(sent_to_mno_at)
WHERE mno_message_id IS NOT NULL
  AND (dlr_status IS NULL OR UPPER(dlr_status) NOT IN ('DELIVRD', 'UNDELIV', 'EXPIRED', 'REJECTD', 'DELETED'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_message_segments_awaiting_final_dlr;
DROP INDEX IF EXISTS idx_service_providers_auto_dlr_enabled;

ALTER TABLE message_segments
DROP COLUMN IF EXISTS dlr_source;

ALTER TABLE service_providers
DROP CONSTRAINT IF EXISTS chk_service_providers_auto_dlr_delay,
DROP CONSTRAINT IF EXISTS chk_service_providers_auto_dlr_status,
DROP COLUMN IF EXISTS auto_dlr_enabled_at,
DROP COLUMN IF EXISTS auto_dlr_delay_seconds,
DROP COLUMN IF EXISTS auto_dlr_status,
DROP COLUMN IF EXISTS auto_dlr_enabled;

-- +goose StatementEnd
