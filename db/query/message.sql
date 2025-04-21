-- name: InsertMessageIn :one
INSERT INTO messages (
    service_provider_id, smpp_credential_id, client_message_id, sender_id_used,
    destination_msisdn, short_message, processing_status, final_status,
    received_at
) VALUES (
    $1, $2, $3, $4, $5, $6, 'received', 'pending', NOW()
) RETURNING id; -- Return only the ID initially

-- name: GetMessagesToRoute :many
SELECT id, destination_msisdn, service_provider_id, sender_id_used -- Select only needed fields
FROM messages
WHERE processing_status = 'received' -- Or 'validated' etc. depending on your flow
ORDER BY received_at
LIMIT $1 -- Batch size (e.g., 5000)
FOR UPDATE SKIP LOCKED; -- Crucial for concurrent workers

-- name: UpdateMessageRoutingInfo :exec
UPDATE messages
SET processing_status = $1, -- e.g., 'routed' or 'failed_routing'
    routed_mno_id = $2,
    error_code = $3,
    error_description = $4
WHERE id = $5;

-- name: GetMessagesToSend :many
SELECT id -- Add other fields needed by the sender (destination, message, sender_id, mno_connection_id etc.)
FROM messages
WHERE processing_status = 'queued_for_send' -- Ready to be sent
ORDER BY processed_for_queue_at -- Or received_at
LIMIT $1 -- Batch size
FOR UPDATE SKIP LOCKED;

-- name: MarkMessageSent :exec
UPDATE messages
SET processing_status = 'sent', -- Or maybe move straight to final status if no DLR expected
    send_status = 'sent',
    sent_to_mno_at = NOW(),
    mno_message_id = $1,
    mno_connection_id = $2
WHERE id = $3;

-- name: MarkMessageSendFailed :exec
UPDATE messages
SET processing_status = 'failed_queueing', -- Or a more specific failure state
    send_status = 'send_failed',
    final_status = 'failed',
    error_code = $1,
    error_description = $2,
    completed_at = NOW()
WHERE id = $3;

-- name: FindMessageByMnoMessageID :one
SELECT id, service_provider_id -- Select fields needed for DLR processing & billing
FROM messages
WHERE mno_message_id = $1
AND sent_to_mno_at > NOW() - INTERVAL '7 days' -- Optimization: Limit search window
LIMIT 1;

-- name: UpdateMessageDLRStatus :exec
UPDATE messages
SET dlr_status = $1,
    final_status = $2, -- e.g., 'delivered', 'failed'
    dlr_received_at = NOW(),
    completed_at = NOW(),
    error_code = $3 -- DLR error code if any
WHERE id = $4;
