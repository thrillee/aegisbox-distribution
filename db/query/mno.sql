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
