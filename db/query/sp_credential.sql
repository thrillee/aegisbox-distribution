-- name: GetSPCredentialByAPIKey :one
-- Gets SP credential based on a hashed API key for HTTP auth.
SELECT
    spc.id,
    spc.service_provider_id,
    spc.protocol,
    spc.status,
    spc.api_key_hash,
    spc.http_config, -- Needed for callback URL?
    sp.default_currency_code
FROM sp_credentials spc
JOIN service_providers sp ON spc.service_provider_id = sp.id
WHERE spc.protocol = 'http'
  AND spc.api_key_hash = $1 -- Compare against the hashed key
  AND spc.status = 'active'
LIMIT 1;

-- name: GetSPCredentialByKeyIdentifier :one
-- Gets SP credential based on a unique API key identifier for HTTP auth.
SELECT
    spc.id,
    spc.service_provider_id,
    spc.protocol,
    spc.status,
    spc.api_key_hash, -- Need the HASH to compare against in Go
    spc.http_config,
    sp.default_currency_code
FROM sp_credentials spc
JOIN service_providers sp ON spc.service_provider_id = sp.id
WHERE spc.protocol = 'http'
  AND spc.api_key_identifier = $1 -- Lookup by IDENTIFIER
  AND spc.status = 'active'
LIMIT 1;

-- name: CreateSPCredential :one
-- Creates an SP credential (SMPP or HTTP)
INSERT INTO sp_credentials (
    service_provider_id, protocol, status,
    system_id, password_hash, bind_type, -- SMPP
    api_key_identifier, api_key_hash, http_config -- HTTP
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetSPCredentialByID :one
SELECT * FROM sp_credentials WHERE id = $1 LIMIT 1;

-- name: ListSPCredentials :many
-- Lists credentials, optionally filtered by Service Provider ID
SELECT * FROM sp_credentials
WHERE ($1::INT IS NULL OR service_provider_id = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSPCredentials :one
-- Counts credentials, optionally filtered by Service Provider ID
SELECT count(*) FROM sp_credentials
WHERE ($1::INT IS NULL OR service_provider_id = $1);

-- name: UpdateSPCredential :one
-- Updates status, password hash, http_config for a credential.
UPDATE sp_credentials
SET
    status = COALESCE(sqlc.narg(status), status),
    password_hash = COALESCE(sqlc.narg(password_hash), password_hash), -- Only for SMPP if password reset
    http_config = COALESCE(sqlc.narg(http_config), http_config),     -- Only for HTTP
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteSPCredential :exec
DELETE FROM sp_credentials WHERE id = $1;
