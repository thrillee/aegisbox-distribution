-- name: InsertMessageIn :one
INSERT INTO messages (
    service_provider_id,
    sp_credential_id, -- Use new table name FK
    client_message_id,
    client_ref,
    original_source_addr,
    original_destination_addr,
    short_message,
    total_segments,
    currency_code,
    submitted_at,
    received_at,
    processing_status, -- Initial status
    final_status       -- Initial status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), 'received', 'pending'
) RETURNING id;

-- name: UpdateMessageRoutingInfo :exec
UPDATE messages
SET processing_status = $1,
    routed_mno_id = $2,
    error_code = $3,
    error_description = $4
WHERE id = $5;

-- name: MarkMessageSent :exec
UPDATE messages
SET processing_status = 'sent',
    send_status = 'sent',
    completed_at = NOW()
WHERE id = $1;

-- name: MarkMessageSendFailed :exec
-- Use this if the Sender.Send fails catastrophically *before* MNO gets involved,
-- or potentially as the final aggregated status if needed.
UPDATE messages
SET
    processing_status = 'send_failed', -- Or keep 'send_attempted' and set final_status?
    final_status = 'failed',
    error_code = $1,
    error_description = $2,
    completed_at = NOW()
WHERE id = $3;

-- name: FindMessageByMnoMessageID :one
SELECT id, service_provider_id, routed_mno_id
FROM messages
WHERE id IN (
    SELECT message_id FROM message_segments
    WHERE mno_message_id = $1
)
AND sent_to_mno_at > NOW() - INTERVAL '7 days'
LIMIT 1;

-- name: UpdateMessageDLRStatus :exec
UPDATE message_segments
SET dlr_status = $1,
    dlr_received_at = NOW(),
    error_code = $2,
    error_description = $3
WHERE mno_message_id = $4;

-- name: UpdateMessageValidatedRouted :exec
UPDATE messages
SET
    processing_status = $1, -- 'routed', 'failed_validation', 'failed_routing'
    routed_mno_id = $2,   
    approved_sender_id = $3,
    template_id = $4,
    error_code = $5,        -- (Mapped internal code)
    error_description = $6 -- (Reason for failure)
WHERE id = $7;

-- name: UpdateMessagePriced :exec
UPDATE messages
SET
    processing_status = $1, -- 'queued_for_send' or 'failed_pricing'
    cost=$2,
    error_code = $3,    
    error_description = $4, 
    processed_for_queue_at = CASE WHEN $1 = 'queued_for_send' THEN NOW() ELSE processed_for_queue_at END
WHERE id = $5;

-- name: GetMessagesToPrice :many
SELECT id, service_provider_id, routed_mno_id
FROM messages
WHERE processing_status = 'routed'
ORDER BY received_at
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: GetMessageDetailsForSending :one
SELECT
    id,
    service_provider_id,
    original_source_addr, -- Used as SenderID for MNO
    original_destination_addr,
    short_message,        -- Full content
    total_segments,
    currency_code,
    client_ref,
    submitted_at,
    -- Also select fields needed for DataCoding, ESMClass, RequestDLR etc. if stored
    -- data_coding, esm_class, request_dlr, is_flash
    routed_mno_id         -- Needed by Processor to select MNO Connector
FROM messages
WHERE id = $1
LIMIT 1;

-- name: GetMessagesByStatus :many
SELECT id, service_provider_id, original_destination_addr, original_source_addr 
FROM messages
WHERE processing_status = $1 -- e.g., 'received', 'routed', 'queued_for_send'
ORDER BY received_at -- Process in FIFO order
LIMIT $2
FOR UPDATE SKIP LOCKED;


-- name: GetMessageDetailsForPricing :one
-- Select fields needed by the Pricer
SELECT
    id,
    service_provider_id,
    routed_mno_id,
    currency_code,
    total_segments
FROM messages
WHERE id = $1
LIMIT 1;

-- name: UpdateMessageSendAttempted :exec
-- Updates status after Sender.Send finishes (regardless of segment success/failure)
UPDATE messages
SET
    processing_status = 'send_attempted'
    -- Optionally clear error code/description if previous failure was transient?
WHERE id = $1;

-- name: GetMessageTotalSegments :one
SELECT total_segments FROM messages WHERE id = $1;

-- name: GetMessageFinalStatus :one
SELECT final_status FROM messages WHERE id = $1;

-- name: UpdateMessageFinalStatus :exec
-- Updates the final aggregated status and error info
UPDATE messages
SET
    final_status = $1, -- 'delivered', 'failed', 'expired', 'rejected'
    error_code = $2,   -- (Aggregated/final error code)
    error_description = $3, -- (Aggregated/final error description)
    completed_at = NOW()
WHERE id = $4
  AND final_status != $1; -- Optional: Prevent redundant updates

-- name: GetSPMessageInfoForDLR :one
-- Select fields needed for DLR forwarding
SELECT
    m.id,
    m.service_provider_id,
    m.client_ref,
    m.original_source_addr,
    m.original_destination_addr,
    m.submitted_at,
    m.completed_at, -- May not be set when DLR arrives
    m.total_segments,
    spc.protocol,
    spc.system_id,    -- This column is nullable
    spc.http_config   -- This is JSONB (likely []byte or custom type in Go)
FROM messages m
JOIN sp_credentials spc ON m.sp_credential_id = spc.id -- Use correct table name
WHERE m.id = $1
LIMIT 1;
