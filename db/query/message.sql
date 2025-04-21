-- name: InsertMessageIn :one
INSERT INTO messages (
    service_provider_id, smpp_credential_id, client_message_id, sender_id_used,
    destination_msisdn, short_message, processing_status, final_status,
    received_at
) VALUES (
    $1, $2, $3, $4, $5, $6, 'received', 'pending', NOW()
) RETURNING id;

-- name: GetMessagesToRoute :many
SELECT id, destination_msisdn, service_provider_id, sender_id_used
FROM messages
WHERE processing_status = 'received'
ORDER BY received_at
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: UpdateMessageRoutingInfo :exec
UPDATE messages
SET processing_status = $1,
    routed_mno_id = $2,
    error_code = $3,
    error_description = $4
WHERE id = $5;

-- name: GetMessagesToSend :many
SELECT id, destination_msisdn, short_message, sender_id_used, routed_mno_id, approved_sender_id, template_id
FROM messages
WHERE processing_status = 'queued_for_send'
ORDER BY processed_for_queue_at
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkMessageSent :exec
UPDATE messages
SET processing_status = 'sent',
    send_status = 'sent',
    completed_at = NOW()
WHERE id = $1;

-- name: MarkMessageSendFailed :exec
UPDATE messages
SET processing_status = 'failed_queueing',
    send_status = 'send_failed',
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
SET processing_status = $1,
    routed_mno_id = $2,
    approved_sender_id = $3,
    template_id = $4,
    error_code = $5,
    error_description = $6
WHERE id = $7;

-- name: UpdateMessagePriced :exec
UPDATE messages
SET processing_status = $1,
    error_code = $2,
    error_description = $3,
    processed_for_queue_at = CASE WHEN $1 = 'queued_for_send' THEN NOW() ELSE processed_for_queue_at END
WHERE id = $4;

-- name: GetMessagesToPrice :many
SELECT id, service_provider_id, routed_mno_id
FROM messages
WHERE processing_status = 'routed'
ORDER BY received_at
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: GetMessageDetailsForSending :one
SELECT id, destination_msisdn, short_message, approved_sender_id, routed_mno_id, sender_id_used
FROM messages
WHERE id = $1;

-- name: GetMessagesByStatus :many
SELECT id, service_provider_id, destination_msisdn, sender_id_used, routed_mno_id
FROM messages
WHERE processing_status = $1
ORDER BY received_at
LIMIT $2
FOR UPDATE SKIP LOCKED;


-- name: GetMessageDetailsForPricing :one
SELECT service_provider_id, routed_mno_id, currency_code, total_segments
FROM messages
WHERE id = $1
LIMIT 1;

