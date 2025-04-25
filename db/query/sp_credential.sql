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
