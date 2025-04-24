-- name: CreateServiceProvider :one
INSERT INTO service_providers (name, email, status, default_currency_code)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetServiceProviderByID :one
SELECT * FROM service_providers WHERE id = $1 LIMIT 1;

-- name: CreateSMPPCredential :one
INSERT INTO sp_credentials (
    service_provider_id,
    protocol,
    status,
    system_id,
    password_hash,
    bind_type,
    http_config
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSPCredentialBySystemID :one
SELECT
    spc.id,
    spc.service_provider_id,
    spc.protocol,
    spc.status,
    spc.system_id,
    spc.password_hash,
    spc.bind_type,
    spc.http_config,
    spc.created_at,
    spc.updated_at,
    sp.name as service_provider_name,
    sp.default_currency_code -- Include currency
FROM sp_credentials spc -- Use new table name
JOIN service_providers sp ON spc.service_provider_id = sp.id
WHERE spc.protocol = 'smpp' -- Ensure it's an SMPP credential
  AND spc.system_id = $1    -- Expects sql.NullString or string based on sqlc generation
  AND spc.status = 'active'
LIMIT 1;

-- name: CreateWallet :one
INSERT INTO wallets (service_provider_id, currency_code, balance, low_balance_threshold)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSPCredentialForHTTPAuth :one
-- SELECT ... FROM sp_credentials WHERE protocol = 'http' AND http_config->>'api_key' = $1 ...

