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
    cost=$2::numeric, -- $2 is decimal.Decimal
    error_code = $3,    
    error_description = $4, 
    processed_for_queue_at = CASE WHEN $5 = 'queued_for_send' THEN NOW() ELSE processed_for_queue_at END
WHERE id = $6;

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
SELECT id, service_provider_id, sp_credential_id, original_destination_addr, original_source_addr 
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
    m.client_message_id,
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

-- name: CountMessageDetails :one
-- Counts messages based on optional filters, joining necessary tables.
SELECT count(DISTINCT m.id)
FROM messages m
JOIN service_providers sp ON m.service_provider_id = sp.id
JOIN sp_credentials spc ON m.sp_credential_id = spc.id
LEFT JOIN mnos mn ON m.routed_mno_id = mn.id
-- Only JOIN segments if filtering by mno_message_id
LEFT JOIN message_segments ms_filter ON m.id = ms_filter.message_id AND sqlc.narg(mno_message_id)::TEXT IS NOT NULL
WHERE
  (sqlc.narg(sp_id)::INT IS NULL OR m.service_provider_id = sqlc.narg(sp_id))
  AND (sqlc.narg(sp_credential_id)::INT IS NULL OR m.sp_credential_id = sqlc.narg(sp_credential_id))
  AND (sqlc.narg(mno_id)::INT IS NULL OR m.routed_mno_id = sqlc.narg(mno_id))
  AND (sqlc.narg(sender_id)::TEXT IS NULL OR m.original_source_addr = sqlc.narg(sender_id))
  AND (sqlc.narg(destination)::TEXT IS NULL OR m.original_destination_addr = sqlc.narg(destination))
  AND (sqlc.narg(final_status)::TEXT IS NULL OR m.final_status = sqlc.narg(final_status))
  AND (sqlc.narg(client_message_id)::TEXT IS NULL OR m.client_message_id = sqlc.narg(client_message_id))
  AND (sqlc.narg(client_ref)::TEXT IS NULL OR m.client_ref = sqlc.narg(client_ref))
  -- Subquery EXISTS check for mno_message_id filter to avoid inflating count due to join
  AND (sqlc.narg(mno_message_id)::TEXT IS NULL OR EXISTS (
        SELECT 1 FROM message_segments ms_exists
        WHERE ms_exists.message_id = m.id AND ms_exists.mno_message_id = sqlc.narg(mno_message_id)
  ))
  AND (sqlc.narg(start_date)::TIMESTAMPTZ IS NULL OR m.submitted_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::TIMESTAMPTZ IS NULL OR m.submitted_at < sqlc.narg(end_date)::TIMESTAMPTZ + INTERVAL '1 day');

-- name: ListMessageDetails :many
-- Lists detailed message information based on optional filters with pagination. Excludes message content.
SELECT DISTINCT ON (m.id) -- Ensures one row per message ID even with segment join
    m.id,
    m.service_provider_id,
    sp.name as service_provider_name, -- Joined SP Name
    m.sp_credential_id,
    COALESCE(spc.system_id, spc.api_key_identifier) AS credential_identifier, -- Joined Credential Identifier
    spc.protocol, -- Joined Protocol
    m.client_message_id,
    m.client_ref,
    m.original_source_addr,
    m.original_destination_addr,
    m.total_segments,
    m.submitted_at,
    m.received_at,
    m.routed_mno_id,
    mn.name as mno_name, -- Joined MNO Name
    m.currency_code,
    m.cost,
    m.processing_status,
    m.final_status,
    m.error_code,
    m.error_description,
    m.completed_at,
    -- Aggregate MNO Message IDs from segments for display (can be NULL if no segments sent/recorded)
    (SELECT string_agg(ms.mno_message_id, ', ') FROM message_segments ms WHERE ms.message_id = m.id) AS mno_message_ids
FROM messages m
JOIN service_providers sp ON m.service_provider_id = sp.id
JOIN sp_credentials spc ON m.sp_credential_id = spc.id
LEFT JOIN mnos mn ON m.routed_mno_id = mn.id
-- LEFT JOIN needed for EXISTS filter below, but DISTINCT ON handles duplicates if join remains
LEFT JOIN message_segments ms_filter ON m.id = ms_filter.message_id AND sqlc.narg(mno_message_id)::TEXT IS NOT NULL
WHERE
  (sqlc.narg(sp_id)::INT IS NULL OR m.service_provider_id = sqlc.narg(sp_id))                     -- $1
  AND (sqlc.narg(sp_credential_id)::INT IS NULL OR m.sp_credential_id = sqlc.narg(sp_credential_id)) -- $2
  AND (sqlc.narg(mno_id)::INT IS NULL OR m.routed_mno_id = sqlc.narg(mno_id))                     -- $3
  AND (sqlc.narg(sender_id)::TEXT IS NULL OR m.original_source_addr = sqlc.narg(sender_id))       -- $4
  AND (sqlc.narg(destination)::TEXT IS NULL OR m.original_destination_addr = sqlc.narg(destination)) -- $5
  AND (sqlc.narg(final_status)::TEXT IS NULL OR m.final_status = sqlc.narg(final_status))         -- $6
  AND (sqlc.narg(client_message_id)::TEXT IS NULL OR m.client_message_id = sqlc.narg(client_message_id)) -- $7
  AND (sqlc.narg(client_ref)::TEXT IS NULL OR m.client_ref = sqlc.narg(client_ref))             -- $8
  AND (sqlc.narg(mno_message_id)::TEXT IS NULL OR EXISTS (                                       -- $9 (mno_message_id filter)
        SELECT 1 FROM message_segments ms_exists
        WHERE ms_exists.message_id = m.id AND ms_exists.mno_message_id = sqlc.narg(mno_message_id)
  ))
  AND (sqlc.narg(start_date)::TIMESTAMPTZ IS NULL OR m.submitted_at >= sqlc.narg(start_date))        -- $10
  AND (sqlc.narg(end_date)::TIMESTAMPTZ IS NULL OR m.submitted_at < sqlc.narg(end_date)::TIMESTAMPTZ + INTERVAL '1 day') -- $11 (end_date filter)
ORDER BY
    m.id DESC -- Order must match DISTINCT ON column first for predictable results
    -- m.submitted_at DESC -- Cannot easily use submitted_at with DISTINCT ON (m.id) unless included
LIMIT sqlc.arg(query_limit)::INT OFFSET sqlc.arg(query_offset)::INT;
