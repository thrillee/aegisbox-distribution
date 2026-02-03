-- name: GetActiveMNOConnections :many
-- Selects all connections marked as 'active' in status
SELECT
    mc.id,
    mc.mno_id,
    mc.protocol,
    mc.status,
    mc.priority,
    -- SMPP Fields
    mc.system_id,
    mc.password,
    mc.host,
    mc.port,
    mc.use_tls,
    mc.bind_type,
    mc.system_type,
    mc.enquire_link_interval_secs,
    mc.request_timeout_secs,
    mc.connect_retry_delay_secs,
    mc.max_window_size,
    mc.default_data_coding,
    mc.source_addr_ton,
    mc.source_addr_npi,
    mc.dest_addr_ton,
    mc.dest_addr_npi,
    -- HTTP Fields
    mc.api_key,
    mc.base_url,
    mc.username,
    mc.http_password,
    mc.secret_key,
    mc.default_sender,
    mc.rate_limit,
    mc.email,
    mc.supports_webhook,
    mc.webhook_path,
    mc.timeout_secs,
    mc.provider_config,
    mc.created_at,
    mc.updated_at
FROM mno_connections mc
WHERE mc.status = 'active';

-- name: CreateMNOConnection :one
INSERT INTO mno_connections (
    mno_id, protocol, status, priority,
    -- SMPP Fields
    system_id, password, host, port, use_tls, bind_type, system_type,
    enquire_link_interval_secs, request_timeout_secs, connect_retry_delay_secs,
    max_window_size, default_data_coding, source_addr_ton, source_addr_npi,
    dest_addr_ton, dest_addr_npi,
    -- HTTP Fields
    api_key, base_url, username, http_password, secret_key, default_sender,
    rate_limit, email, supports_webhook, webhook_path, timeout_secs,
    provider_config
) VALUES (
    $1, $2, COALESCE($3, 'active'), COALESCE($4, 1),
    $5, $6, $7, $8, $9, $10, $11,
    $12, $13, $14,
    $15, $16, $17, $18,
    $19, $20,
    $21, $22, $23, $24, $25, $26,
    $27, $28, $29, $30, $31,
    $32
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
    priority = COALESCE(sqlc.narg(priority), priority),
    system_id = COALESCE(sqlc.narg(system_id), system_id),
    password = COALESCE(sqlc.narg(password), password),
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
    api_key = COALESCE(sqlc.narg(api_key), api_key),
    base_url = COALESCE(sqlc.narg(base_url), base_url),
    username = COALESCE(sqlc.narg(username), username),
    http_password = COALESCE(sqlc.narg(http_password), http_password),
    secret_key = COALESCE(sqlc.narg(secret_key), secret_key),
    default_sender = COALESCE(sqlc.narg(default_sender), default_sender),
    rate_limit = COALESCE(sqlc.narg(rate_limit), rate_limit),
    email = COALESCE(sqlc.narg(email), email),
    supports_webhook = COALESCE(sqlc.narg(supports_webhook), supports_webhook),
    webhook_path = COALESCE(sqlc.narg(webhook_path), webhook_path),
    timeout_secs = COALESCE(sqlc.narg(timeout_secs), timeout_secs),
    provider_config = COALESCE(sqlc.narg(provider_config), provider_config),
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
