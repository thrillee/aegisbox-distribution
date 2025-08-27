-- name: GetActiveOtpTemplateAssignment :one
-- Selects the highest priority active template for a given alternative sender and optionally MNO.
SELECT
    osta.*,
    omt.content_template,
    omt.default_brand_name
FROM otp_sender_template_assignments osta
JOIN otp_message_templates omt ON osta.otp_message_template_id = omt.id
WHERE osta.otp_alternative_sender_id = $1
  AND osta.status = 'active'
  AND omt.status = 'active' -- Ensure template itself is active
  AND (osta.mno_id = sqlc.narg(mno_id) OR (sqlc.narg(mno_id) IS NULL AND osta.mno_id IS NULL)) -- Matches specific MNO or global assignment
ORDER BY osta.priority ASC, osta.id ASC -- Lower priority number is better
LIMIT 1;

-- name: IncrementOtpAssignmentUsage :one
UPDATE otp_sender_template_assignments
SET assignment_usage_count = assignment_usage_count + 1
WHERE id = $1
RETURNING id, assignment_usage_count;

-- name: CreateOtpSenderTemplateAssignment :one
INSERT INTO otp_sender_template_assignments (
    otp_alternative_sender_id,
    otp_message_template_id,
    mno_id,
    priority,
    status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetOtpSenderTemplateAssignmentByID :one
SELECT
    osta.*,
    oas.sender_id_string AS alternative_sender_string,
    omt.name AS template_name,
    m.name AS mno_name
FROM otp_sender_template_assignments osta
JOIN otp_alternative_senders oas ON osta.otp_alternative_sender_id = oas.id
JOIN otp_message_templates omt ON osta.otp_message_template_id = omt.id
LEFT JOIN mnos m ON osta.mno_id = m.id
WHERE osta.id = $1 LIMIT 1;

-- name: ListOtpSenderTemplateAssignments :many
SELECT
    osta.*,
    oas.sender_id_string AS alternative_sender_string,
    omt.name AS template_name,
    m.name AS mno_name
FROM otp_sender_template_assignments osta
JOIN otp_alternative_senders oas ON osta.otp_alternative_sender_id = oas.id
JOIN otp_message_templates omt ON osta.otp_message_template_id = omt.id
LEFT JOIN mnos m ON osta.mno_id = m.id
WHERE
    (sqlc.narg(otp_alternative_sender_id)::INT IS NULL OR osta.otp_alternative_sender_id = sqlc.narg(otp_alternative_sender_id)::INT)
AND (sqlc.narg(otp_message_template_id)::INT IS NULL OR osta.otp_message_template_id = sqlc.narg(otp_message_template_id)::INT)
AND (sqlc.narg(mno_id)::INT IS NULL OR osta.mno_id = sqlc.narg(mno_id)::INT)
AND (sqlc.narg(status)::VARCHAR IS NULL OR osta.status = sqlc.narg(status)::VARCHAR)
ORDER BY osta.priority ASC, osta.id DESC
LIMIT sqlc.arg(limit_val)::INT
OFFSET sqlc.arg(offset_val)::INT;

-- name: CountOtpSenderTemplateAssignments :one
SELECT count(*)
FROM otp_sender_template_assignments
WHERE
    (sqlc.narg(otp_alternative_sender_id)::INT IS NULL OR otp_alternative_sender_id = sqlc.narg(otp_alternative_sender_id)::INT)
AND (sqlc.narg(otp_message_template_id)::INT IS NULL OR otp_message_template_id = sqlc.narg(otp_message_template_id)::INT)
AND (sqlc.narg(mno_id)::INT IS NULL OR mno_id = sqlc.narg(mno_id)::INT)
AND (sqlc.narg(status)::VARCHAR IS NULL OR status = sqlc.narg(status)::VARCHAR);

-- name: UpdateOtpSenderTemplateAssignment :one
UPDATE otp_sender_template_assignments
SET
    otp_alternative_sender_id = COALESCE(sqlc.narg(otp_alternative_sender_id), otp_alternative_sender_id),
    otp_message_template_id = COALESCE(sqlc.narg(otp_message_template_id), otp_message_template_id),
    mno_id = CASE
                WHEN sqlc.narg(clear_mno_id)::BOOLEAN = TRUE THEN NULL
                ELSE COALESCE(sqlc.narg(mno_id), mno_id)
             END,
    priority = COALESCE(sqlc.narg(priority), priority),
    status = COALESCE(sqlc.narg(status), status),
    assignment_usage_count = COALESCE(sqlc.narg(assignment_usage_count), assignment_usage_count), -- If manual adjustment is needed
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteOtpSenderTemplateAssignment :exec
DELETE FROM otp_sender_template_assignments WHERE id = $1;
