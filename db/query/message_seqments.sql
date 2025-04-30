-- name: FindSegmentByMnoMessageID :one
-- Finds segment details based on MNO Message ID for DLR processing
SELECT id as segment_id, message_id, segment_seqn
FROM message_segments
WHERE mno_message_id = $1 -- Expects sql.NullString
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
    mno_message_id = $1,    -- sql.NullString
    mno_connection_id = $2, -- sql.NullInt32
    sent_to_mno_at = NOW(),
    error_code = NULL,      -- Clear previous errors if any
    error_description = NULL
WHERE id = $3;

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
    error_code = $2       -- (Error code from DLR)
    -- Keep segment-level error_description from MNO? Or clear it? Let's clear.
    -- error_description = NULL
WHERE id = $3;

-- name: GetSegmentStatusesForMessage :many
-- Gets all segment statuses for final status aggregation
SELECT segment_seqn, dlr_status -- dlr_status is sql.NullString
FROM message_segments
WHERE message_id = $1
ORDER BY segment_seqn;
