-- +goose Up
-- +goose StatementBegin

-- Auto DLR can also complete messages the upstream never answered: sent, but no MNO message ID and
-- no rejection (the submit timed out, or its response was never recorded). Off by default.
ALTER TABLE service_providers
ADD COLUMN IF NOT EXISTS auto_dlr_include_unconfirmed BOOLEAN NOT NULL DEFAULT FALSE;

-- Unconfirmed segments still waiting, for the auto DLR scan (the accepted ones use
-- idx_message_segments_awaiting_final_dlr).
CREATE INDEX IF NOT EXISTS idx_message_segments_awaiting_unconfirmed ON message_segments(sent_to_mno_at)
WHERE mno_message_id IS NULL AND sent_to_mno_at IS NOT NULL
  AND (dlr_status IS NULL OR UPPER(dlr_status) NOT IN ('DELIVRD', 'UNDELIV', 'EXPIRED', 'REJECTD', 'DELETED'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_message_segments_awaiting_unconfirmed;
ALTER TABLE service_providers DROP COLUMN IF EXISTS auto_dlr_include_unconfirmed;
-- +goose StatementEnd
