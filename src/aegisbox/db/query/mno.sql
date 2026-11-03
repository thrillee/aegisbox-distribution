-- name: CreateMNO :one
INSERT INTO mnos (
    name, country_code, network_code, status
) VALUES (
    $1, $2, $3, COALESCE($4, 'active') -- Use input status or default to 'active'
) RETURNING *;

-- name: GetMNOByID :one
SELECT * FROM mnos WHERE id = $1 LIMIT 1;

-- name: ListMNOs :many
SELECT * FROM mnos
ORDER BY name -- Or country_code, name
LIMIT $1 OFFSET $2;

-- name: CountMNOs :one
SELECT count(*) FROM mnos;

-- name: UpdateMNO :one
UPDATE mnos
SET
    name = COALESCE(sqlc.narg(name), name),
    country_code = COALESCE(sqlc.narg(country_code), country_code),
    network_code = sqlc.narg(network_code), -- Allow setting network_code to NULL
    status = COALESCE(sqlc.narg(status), status),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMNO :exec
-- Consider consequences - deletes might fail if MNO is referenced by connections/rules (ON DELETE RESTRICT)
-- It might be safer to just set status to 'inactive'.
DELETE FROM mnos WHERE id = $1;

-- Optional: Query to set MNO status to inactive instead of deleting
-- name: DeactivateMNO :one
UPDATE mnos SET status = 'inactive', updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: IsSimulatorBlockedForSP :one
-- True when the MNO is a simulator and the service provider is not a test customer: the message must
-- not be sent through it.
SELECT COALESCE((
    SELECT m.is_simulator AND NOT sp.is_test
    FROM mnos m, service_providers sp
    WHERE m.id = sqlc.arg(mno_id) AND sp.id = sqlc.arg(service_provider_id)
), FALSE)::BOOLEAN AS blocked;

-- name: GetSimulatorRoutingPolicy :one
-- For routing: whether the credential's customer is a test customer, and whether the destination
-- is whitelisted for it (a whitelisted number gets real delivery instead of the simulator).
SELECT
    sp.is_test,
    EXISTS (
        SELECT 1 FROM simulator_whitelist_numbers w
        WHERE w.msisdn = sqlc.arg(msisdn)::TEXT
          AND (w.service_provider_id IS NULL OR w.service_provider_id = sp.id)
    )::BOOLEAN AS is_whitelisted
FROM sp_credentials c
JOIN service_providers sp ON sp.id = c.service_provider_id
WHERE c.id = sqlc.arg(sp_credential_id);
