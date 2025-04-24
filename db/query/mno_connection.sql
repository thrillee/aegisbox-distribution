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

-- name: UpdateMNOConnectionStatus :exec
-- Updates the runtime status (e.g., bound, disconnected) - Use with caution, might conflict with DB status field
-- Consider if a separate runtime status cache is better than updating DB frequently.
-- UPDATE mno_connections SET runtime_status = $1, updated_at = NOW() WHERE id = $2;
