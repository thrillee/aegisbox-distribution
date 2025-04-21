-- name: FindSegmentByMnoMessageID :one
SELECT 
    ms.message_id, 
    ms.segment_seqn, 
    ms.id AS segment_id,
    m.service_provider_id,
    m.client_message_id
FROM message_segments ms
JOIN messages m ON m.id = ms.message_id
WHERE ms.mno_message_id = $1
LIMIT 1;

-- name: CreateMessageSegment :one
INSERT INTO message_segments (message_id, segment_seqn, created_at)
VALUES ($1, $2, NOW())
RETURNING id;

-- name: UpdateSegmentSent :exec
UPDATE message_segments
SET mno_message_id = $1,
    mno_connection_id = $2,
    sent_to_mno_at = NOW()
WHERE id = $3;

-- name: UpdateSegmentSendFailed :exec
UPDATE message_segments
SET error_code = $1,
    error_description = $2,
    sent_to_mno_at = NOW() -- Mark attempt time even on failure
WHERE id = $3;


-- name: UpdateSegmentDLR :exec
UPDATE message_segments
SET dlr_status = $1,
    dlr_received_at = NOW(),
    error_code = $2 -- DLR error code
WHERE id = $3;

-- name: GetSegmentStatusesForMessage :many
SELECT segment_seqn, dlr_status FROM message_segments
WHERE message_id = $1
ORDER BY segment_seqn;

-- name: UpdateMessageFinalStatus :exec
UPDATE messages
SET final_status = $1,
    completed_at = NOW()
WHERE id = $2;

-- name: GetMessageTotalSegments :one
SELECT total_segments
FROM messages
WHERE id = $1;

-- name: GetMessageFinalStatus :one
SELECT final_status
FROM messages
WHERE id = $1;

