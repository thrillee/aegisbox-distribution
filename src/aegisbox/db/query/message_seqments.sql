-- name: FindSegmentByMnoMessageID :one
-- Finds segment details based on MNO Message ID for DLR processing
SELECT 
    ms.id as segment_id, 
    ms.message_id, 
    ms.segment_seqn,
    ms.routing_assignment_id,
    ms.sent_to_mno_at,
    m.approved_sender_id,
    m.routed_mno_id
FROM message_segments ms
JOIN messages m ON ms.message_id = m.id
WHERE ms.mno_message_id = $1 -- Expects sql.NullString
LIMIT 1;

-- name: CreateMessageSegment :one
-- Inserts initial segment record OR returns existing ID if conflict occurs on (message_id, segment_seqn).
INSERT INTO message_segments (
    message_id,
    segment_seqn,
    created_at
) VALUES (
    $1, $2, NOW()
)
ON CONFLICT (message_id, segment_seqn) DO UPDATE SET
    -- No-op update: Set a column to its existing value just to enable RETURNING
    segment_seqn = EXCLUDED.segment_seqn
    -- updated_at = NOW()
RETURNING id;

-- name: UpdateSegmentSent :exec
-- Updates segment after successful MNO submission
UPDATE message_segments
SET
    -- SMPP connectors report the MNO message ID asynchronously, possibly before this update
    -- runs; never replace a recorded ID with NULL.
    mno_message_id = COALESCE(sqlc.narg(mno_message_id), mno_message_id),
    mno_connection_id = sqlc.narg(mno_connection_id),
    sent_to_mno_at = NOW(),
    -- Clear an earlier error once the MNO accepts (an ID is known). Without an ID (an SMPP submit
    -- whose response has not been seen) keep it: a fast rejection may already have been recorded.
    error_code = CASE WHEN COALESCE(sqlc.narg(mno_message_id), mno_message_id) IS NULL THEN error_code ELSE NULL END,
    error_description = CASE WHEN COALESCE(sqlc.narg(mno_message_id), mno_message_id) IS NULL THEN error_description ELSE NULL END
WHERE id = sqlc.arg(id);

-- name: UpdateSegmentSendFailed :exec
-- Updates segment if MNO submission attempt failed
UPDATE message_segments
SET
    error_code = $1,
    error_description = $2,
    sent_to_mno_at = NOW()  -- Mark attempt time even on failure
WHERE id = $3;


-- name: UpdateSegmentDLR :exec
-- Updates segment status based on received DLR
UPDATE message_segments
SET
    dlr_status = $1,
    dlr_received_at = NOW(),
    error_code = $2,      -- (Error code from DLR)
    dlr_source = NULL     -- Real DLRs replace any earlier manual/auto DLR
    -- Keep segment-level error_description from MNO? Or clear it? Let's clear.
    -- error_description = NULL
WHERE id = $3;

-- name: GetSegmentStatusesForMessage :many
-- Gets all segment statuses for final status aggregation
SELECT segment_seqn, dlr_status -- dlr_status is sql.NullString
FROM message_segments
WHERE message_id = $1
ORDER BY segment_seqn;

-- name: ApplyManualDLRAndEnqueue :one
-- Completes segments of a message that the MNO accepted but that have no final DLR yet
-- (no DLR, 'no_dlr' or an intermediate status like ENROUTE) and, in the same statement,
-- queues the DLR for forwarding to the SP. Either both happen or neither does, so a DLR
-- can never be applied without being forwarded. No job is queued when nothing was pending.
-- When sent_before is set, only segments sent before it are updated.
-- The pending condition must match idx_message_segments_awaiting_final_dlr.
WITH updated AS (
    UPDATE message_segments ms
    SET
        dlr_status = sqlc.arg(dlr_status),
        dlr_received_at = NOW(),
        error_code = sqlc.narg(error_code),
        dlr_source = sqlc.arg(dlr_source)
    WHERE ms.message_id = sqlc.arg(message_id)::BIGINT
      AND ms.mno_message_id IS NOT NULL
      AND (ms.dlr_status IS NULL OR UPPER(ms.dlr_status) NOT IN ('DELIVRD', 'UNDELIV', 'EXPIRED', 'REJECTD', 'DELETED'))
      AND NOT EXISTS (
        -- Never complete a message that was already refunded or closed as failed
        -- (e.g. part of it failed to send): the SP must not get a DLR that contradicts the refund.
        SELECT 1 FROM wallet_transactions wt
        WHERE wt.message_id = ms.message_id AND wt.transaction_type = 'reversal'
      )
      AND EXISTS (
        SELECT 1 FROM messages m
        WHERE m.id = ms.message_id AND m.final_status NOT IN ('failed', 'rejected', 'expired')
      )
      AND (
        sqlc.narg(sent_before)::TIMESTAMPTZ IS NULL
        OR (ms.sent_to_mno_at IS NOT NULL AND ms.sent_to_mno_at < sqlc.narg(sent_before)::TIMESTAMPTZ)
      )
    RETURNING ms.id
),
queued AS (
    INSERT INTO dlr_forwarding_queue (message_id, payload, max_attempts, created_at, status)
    SELECT sqlc.arg(message_id)::BIGINT, sqlc.arg(payload)::JSONB, sqlc.arg(max_attempts)::INT, NOW(), 'pending'
    WHERE EXISTS (SELECT 1 FROM updated)
    RETURNING id
)
SELECT
    (SELECT COUNT(*) FROM updated)::INT AS segments_updated,
    COALESCE((SELECT id FROM queued), 0)::BIGINT AS queue_id;

-- name: GetAutoDLRCandidates :many
-- Finds messages of SPs with auto DLR enabled that have segments still waiting for a
-- final DLR past the SP's configured delay. not_before bounds the scan of
-- idx_message_segments_awaiting_final_dlr so old undelivered segments are not rescanned.
SELECT DISTINCT ON (m.id)
    m.id AS message_id,
    sp.auto_dlr_status,
    (NOW() - make_interval(secs => sp.auto_dlr_delay_seconds))::TIMESTAMPTZ AS sent_before
FROM message_segments ms
JOIN messages m ON ms.message_id = m.id
JOIN service_providers sp ON m.service_provider_id = sp.id
WHERE sp.auto_dlr_enabled
  AND ms.mno_message_id IS NOT NULL
  AND (ms.dlr_status IS NULL OR UPPER(ms.dlr_status) NOT IN ('DELIVRD', 'UNDELIV', 'EXPIRED', 'REJECTD', 'DELETED'))
  AND ms.sent_to_mno_at >= sqlc.arg(not_before)::TIMESTAMPTZ
  AND ms.sent_to_mno_at >= sp.auto_dlr_enabled_at
  AND ms.sent_to_mno_at < NOW() - make_interval(secs => sp.auto_dlr_delay_seconds)
  AND m.processing_status NOT IN ('failed_validation', 'failed_routing', 'failed_pricing')
  AND m.final_status NOT IN ('failed', 'rejected', 'expired')
  AND NOT EXISTS (
    SELECT 1 FROM wallet_transactions wt
    WHERE wt.message_id = m.id AND wt.transaction_type = 'reversal'
  )
ORDER BY m.id
LIMIT sqlc.arg(max_results);
