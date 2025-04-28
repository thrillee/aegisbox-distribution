-- name: GetActiveMNOConnections :many
-- Selects all connections marked as 'active' in status
SELECT
    id,
    mno_id,
    protocol,
    status,
    -- SMPP Fields
    system_id,
    password,
    host,
    port,
    use_tls,
    bind_type,
    system_type,
    enquire_link_interval_secs,
    request_timeout_secs,
    connect_retry_delay_secs,
    max_window_size,
    default_data_coding,
    source_addr_ton,
    source_addr_npi,
    dest_addr_ton,
    dest_addr_npi,
    -- HTTP Fields
    http_config,
    created_at,
    updated_at
FROM mno_connections
WHERE status = 'active';

-- name: CreateMNOConnection :one
INSERT INTO mno_connections (
    mno_id, protocol, status,
    system_id, password, host, port, use_tls, bind_type, system_type,
    enquire_link_interval_secs, request_timeout_secs, connect_retry_delay_secs,
    max_window_size, default_data_coding, source_addr_ton, source_addr_npi,
    dest_addr_ton, dest_addr_npi,
    http_config
) VALUES (
    $1, $2, COALESCE($3, 'active'),
    $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13,
    $14, $15, $16, $17,
    $18, $19,
    $20
) RETURNING *;

-- name: GetMNOConnectionByID :one
SELECT * FROM mno_connections WHERE id = $1 LIMIT 1;

-- name: ListMNOConnectionsByMNO :many
SELECT * FROM mno_connections
WHERE mno_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMNOConnectionsByMNO :one
SELECT count(*) FROM mno_connections
WHERE mno_id = $1;

-- name: UpdateMNOConnection :one
UPDATE mno_connections
SET
    status = COALESCE(sqlc.narg(status), status),
    system_id = COALESCE(sqlc.narg(system_id), system_id),
    password = COALESCE(sqlc.narg(password), password), -- Update password directly if provided
    host = COALESCE(sqlc.narg(host), host),
    port = COALESCE(sqlc.narg(port), port),
    use_tls = COALESCE(sqlc.narg(use_tls), use_tls),
    bind_type = COALESCE(sqlc.narg(bind_type), bind_type),
    system_type = COALESCE(sqlc.narg(system_type), system_type),
    enquire_link_interval_secs = COALESCE(sqlc.narg(enquire_link_interval_secs), enquire_link_interval_secs),
    request_timeout_secs = COALESCE(sqlc.narg(request_timeout_secs), request_timeout_secs),
    connect_retry_delay_secs = COALESCE(sqlc.narg(connect_retry_delay_secs), connect_retry_delay_secs),
    max_window_size = COALESCE(sqlc.narg(max_window_size), max_window_size),
    default_data_coding = COALESCE(sqlc.narg(default_data_coding), default_data_coding),
    source_addr_ton = COALESCE(sqlc.narg(source_addr_ton), source_addr_ton),
    source_addr_npi = COALESCE(sqlc.narg(source_addr_npi), source_addr_npi),
    dest_addr_ton = COALESCE(sqlc.narg(dest_addr_ton), dest_addr_ton),
    dest_addr_npi = COALESCE(sqlc.narg(dest_addr_npi), dest_addr_npi),
    http_config = COALESCE(sqlc.narg(http_config), http_config),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMNOConnection :exec
DELETE FROM mno_connections WHERE id = $1;

-- name: UpdateMNOConnectionStatus :exec
-- Updates the runtime status (e.g., bound, disconnected) - Use with caution, might conflict with DB status field
-- Consider if a separate runtime status cache is better than updating DB frequently.
-- UPDATE mno_connections SET runtime_status = $1, updated_at = NOW() WHERE id = $2;

-- Already Exists (Potentially):
-- name: GetActiveMNOConnections :many
