-- name: CreateServiceProvider :one
INSERT INTO service_providers (name, email, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetServiceProviderByID :one
SELECT * FROM service_providers
WHERE id = $1 LIMIT 1;

-- name: CreateSMPPCredential :one
INSERT INTO smpp_credentials (service_provider_id, system_id, password_hash, bind_type, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSMPPCredentialBySystemID :one
SELECT sp.id as service_provider_id, sp.name as service_provider_name, sc.*
FROM smpp_credentials sc
JOIN service_providers sp ON sc.service_provider_id = sp.id
WHERE sc.system_id = $1 AND sc.status = 'active'
LIMIT 1;

-- name: CreateWallet :one
INSERT INTO wallets (service_provider_id, currency_code, balance, low_balance_threshold)
VALUES ($1, $2, $3, $4)
RETURNING *;
