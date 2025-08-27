-- name: ListSpSpecificOtpAlternativeSenders :many
-- Selects Service Provider-specific, active senders that are not depleted.
-- Can filter by a specific MNO (if $2 is not NULL) OR select SP-specific senders that are MNO-agnostic (sender.mno_id IS NULL, if $2 is NULL).
SELECT *
FROM otp_alternative_senders
WHERE service_provider_id = $1 -- Required: service_provider_id
  AND status = 'active'
  AND current_usage_count < max_usage_count
  AND (mno_id = sqlc.narg(mno_id) OR (sqlc.narg(mno_id) IS NULL AND mno_id IS NULL))
ORDER BY last_used_at ASC NULLS FIRST, id ASC -- Rotate by least recently used
LIMIT $2;

-- name: IncrementOtpAlternativeSenderUsage :one
UPDATE otp_alternative_senders
SET
    current_usage_count = current_usage_count + 1,
    last_used_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListGlobalOtpAlternativeSenders :many
-- Selects active senders that are not depleted, optionally for a specific MNO or global ones (MNO ID IS NULL).
-- Orders by least recently used to attempt rotation.
SELECT *
FROM otp_alternative_senders
WHERE status = 'active'
  AND current_usage_count < max_usage_count
  AND (mno_id = sqlc.narg(mno_id) OR (sqlc.narg(mno_id) IS NULL AND mno_id IS NULL)) -- Matches specific MNO or global
ORDER BY last_used_at ASC NULLS FIRST, id ASC -- Prioritize those never used or least recently used
LIMIT $1; -- Limit how many to fetch for selection in code


-- name: CreateOtpAlternativeSender :one
INSERT INTO otp_alternative_senders (
    sender_id_string,
    mno_id,
    service_provider_id,
    status,
    max_usage_count,
    current_usage_count, -- Allow setting initial count if needed
    reset_interval_hours,
    last_reset_at,
    notes
) VALUES (
    $1, $2, $3, $4, $5, COALESCE(sqlc.narg(current_usage_count), 0), $6, COALESCE(sqlc.narg(last_reset_at), NOW()), $7
) RETURNING *;

-- name: GetOtpAlternativeSenderByID :one
SELECT
    oas.*,
    m.name AS mno_name,
    sp.name AS service_provider_name
FROM otp_alternative_senders oas
LEFT JOIN mnos m ON oas.mno_id = m.id
LEFT JOIN service_providers sp ON oas.service_provider_id = sp.id
WHERE oas.id = $1 LIMIT 1;

-- name: ListOtpAlternativeSenders :many
SELECT
    oas.*,
    m.name AS mno_name,
    sp.name AS service_provider_name
FROM otp_alternative_senders oas
LEFT JOIN mnos m ON oas.mno_id = m.id
LEFT JOIN service_providers sp ON oas.service_provider_id = sp.id
WHERE
    (sqlc.narg(status)::VARCHAR IS NULL OR oas.status = sqlc.narg(status)::VARCHAR)
AND (sqlc.narg(mno_id)::INT IS NULL OR oas.mno_id = sqlc.narg(mno_id)::INT)
AND (
        (sqlc.narg(service_provider_id)::INT IS NOT NULL AND oas.service_provider_id = sqlc.narg(service_provider_id)::INT) OR
        (sqlc.narg(service_provider_id)::INT IS NULL AND sqlc.narg(include_global_sp_null)::BOOLEAN = TRUE AND oas.service_provider_id IS NULL) OR
        (sqlc.narg(service_provider_id)::INT IS NULL AND sqlc.narg(include_global_sp_null)::BOOLEAN = FALSE) -- if sp_id not given and not including global, effectively no filter on sp_id
    )
ORDER BY oas.id DESC -- Or other preferred order
LIMIT sqlc.arg(limit_val)::INT
OFFSET sqlc.arg(offset_val)::INT;

-- name: CountOtpAlternativeSenders :one
SELECT count(*)
FROM otp_alternative_senders
WHERE
    (sqlc.narg(status)::VARCHAR IS NULL OR status = sqlc.narg(status)::VARCHAR)
AND (sqlc.narg(mno_id)::INT IS NULL OR mno_id = sqlc.narg(mno_id)::INT)
AND (
        (sqlc.narg(service_provider_id)::INT IS NOT NULL AND service_provider_id = sqlc.narg(service_provider_id)::INT) OR
        (sqlc.narg(service_provider_id)::INT IS NULL AND sqlc.narg(include_global_sp_null)::BOOLEAN = TRUE AND service_provider_id IS NULL) OR
        (sqlc.narg(service_provider_id)::INT IS NULL AND sqlc.narg(include_global_sp_null)::BOOLEAN = FALSE)
    );

-- name: UpdateOtpAlternativeSender :one
UPDATE otp_alternative_senders
SET
    sender_id_string = COALESCE(sqlc.narg(sender_id_string), sender_id_string),
    mno_id = CASE
                WHEN sqlc.narg(clear_mno_id)::BOOLEAN = TRUE THEN NULL
                ELSE COALESCE(sqlc.narg(mno_id), mno_id)
             END,
    service_provider_id = CASE
                            WHEN sqlc.narg(clear_service_provider_id)::BOOLEAN = TRUE THEN NULL
                            ELSE COALESCE(sqlc.narg(service_provider_id), service_provider_id)
                          END,
    status = COALESCE(sqlc.narg(status), status),
    max_usage_count = COALESCE(sqlc.narg(max_usage_count), max_usage_count),
    current_usage_count = COALESCE(sqlc.narg(current_usage_count), current_usage_count),
    reset_interval_hours = COALESCE(sqlc.narg(reset_interval_hours), reset_interval_hours),
    last_reset_at = COALESCE(sqlc.narg(last_reset_at), last_reset_at),
    notes = COALESCE(sqlc.narg(notes), notes),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteOtpAlternativeSender :exec
DELETE FROM otp_alternative_senders WHERE id = $1;

