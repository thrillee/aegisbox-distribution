-- name: ListAlternativeSendersWithAssignedTemplates :many
SELECT
    oas.id AS alternative_sender_id,
    oas.sender_id_string,
    oas.status AS alternative_sender_status,
    oas.mno_id AS sender_mno_id, -- MNO ID of the alternative sender itself
    sender_mno.name AS sender_mno_name,
    oas.service_provider_id AS sender_service_provider_id,
    sp.name AS sender_service_provider_name,
    osta.id AS assignment_id, -- Nullable if no active assignment/template
    osta.status AS assignment_status, -- Nullable
    osta.priority AS assignment_priority, -- Nullable
    osta.mno_id AS assignment_mno_id, -- MNO ID specific to the assignment (can be different from sender's MNO or NULL)
    assignment_mno.name AS assignment_mno_name, -- Nullable
    omt.id AS template_id, -- Nullable
    omt.name AS template_name -- Nullable
FROM otp_alternative_senders oas
LEFT JOIN mnos sender_mno ON oas.mno_id = sender_mno.id
LEFT JOIN service_providers sp ON oas.service_provider_id = sp.id
LEFT JOIN otp_sender_template_assignments osta
    ON oas.id = osta.otp_alternative_sender_id AND osta.status = 'active'
LEFT JOIN otp_message_templates omt
    ON osta.otp_message_template_id = omt.id AND omt.status = 'active'
LEFT JOIN mnos assignment_mno ON osta.mno_id = assignment_mno.id
WHERE
    oas.status = 'active' -- Consider if you want to filter alternative senders by status here, or show all and their assignment statuses.
AND (sqlc.narg(filter_sender_id_string)::VARCHAR IS NULL OR oas.sender_id_string = sqlc.narg(filter_sender_id_string)::VARCHAR)
AND (sqlc.narg(filter_service_provider_id)::INT IS NULL OR oas.service_provider_id = sqlc.narg(filter_service_provider_id)::INT)
ORDER BY
    oas.sender_id_string ASC, -- Group by sender
    osta.mno_id ASC NULLS FIRST, -- Order assignments for a sender: MNO-agnostic first, then by MNO ID
    osta.priority ASC; -- Then by priority of assignment

-- name: CreateOtpMessageTemplate :one
INSERT INTO otp_message_templates (
    name,
    content_template,
    default_brand_name,
    status
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetOtpMessageTemplateByID :one
SELECT * FROM otp_message_templates
WHERE id = $1 LIMIT 1;

-- name: ListOtpMessageTemplates :many
SELECT *
FROM otp_message_templates
WHERE (sqlc.narg(status)::VARCHAR IS NULL OR status = sqlc.narg(status)::VARCHAR)
ORDER BY name ASC
LIMIT sqlc.arg(limit_val)::INT
OFFSET sqlc.arg(offset_val)::INT;

-- name: CountOtpMessageTemplates :one
SELECT count(*)
FROM otp_message_templates
WHERE (sqlc.narg(status)::VARCHAR IS NULL OR status = sqlc.narg(status)::VARCHAR);

-- name: UpdateOtpMessageTemplate :one
UPDATE otp_message_templates
SET
    name = COALESCE(sqlc.narg(name), name),
    content_template = COALESCE(sqlc.narg(content_template), content_template),
    default_brand_name = COALESCE(sqlc.narg(default_brand_name), default_brand_name),
    status = COALESCE(sqlc.narg(status), status),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteOtpMessageTemplate :exec
DELETE FROM otp_message_templates WHERE id = $1;
