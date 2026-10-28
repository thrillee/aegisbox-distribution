-- name: PickRouteSender :one
-- Picks a sender from a connection's pool by the connection's sender_pool_strategy and records
-- the use in the same statement. No row when the pool has no active sender.
WITH conn AS (
    SELECT mc.sender_pool_strategy AS strategy FROM mno_connections mc WHERE mc.id = sqlc.arg(mno_connection_id)
), picked AS (
    SELECT rs.id
    FROM route_senders rs, conn
    WHERE rs.mno_connection_id = sqlc.arg(mno_connection_id)
      AND rs.status = 'active'
    ORDER BY
        CASE WHEN conn.strategy = 'ordered' THEN rs.position END ASC,
        CASE WHEN conn.strategy = 'round_robin' THEN rs.last_used_at END ASC NULLS FIRST,
        CASE WHEN conn.strategy = 'random' THEN random() END,
        rs.position ASC,
        rs.id ASC
    LIMIT 1
)
UPDATE route_senders
SET usage_count = route_senders.usage_count + 1,
    last_used_at = clock_timestamp()
FROM picked
WHERE route_senders.id = picked.id
RETURNING route_senders.id, route_senders.sender_id_string;

-- name: ListRouteSenders :many
SELECT * FROM route_senders
WHERE mno_connection_id = $1
ORDER BY position, id;

-- name: GetRouteSender :one
SELECT * FROM route_senders WHERE id = $1 LIMIT 1;

-- name: CreateRouteSender :one
INSERT INTO route_senders (mno_connection_id, sender_id_string, status, position)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateRouteSender :one
UPDATE route_senders
SET status = $2,
    position = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRouteSender :execrows
DELETE FROM route_senders WHERE id = $1;

-- name: GetMNOConnectionSenderPoolStrategy :one
SELECT id, mno_id, sender_pool_strategy FROM mno_connections WHERE id = $1 LIMIT 1;

-- name: UpdateMNOConnectionSenderPoolStrategy :one
UPDATE mno_connections
SET sender_pool_strategy = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, mno_id, sender_pool_strategy;

-- name: UpdateServiceProviderRouteSenderPool :one
UPDATE service_providers
SET use_route_sender_pool = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, use_route_sender_pool;

-- name: UsesRouteSenderPool :one
SELECT use_route_sender_pool FROM service_providers WHERE id = $1 LIMIT 1;

-- name: SetMessageSentSourceAddr :exec
UPDATE messages SET sent_source_addr = $2 WHERE id = $1;
