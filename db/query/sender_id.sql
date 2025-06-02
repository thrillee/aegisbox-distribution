-- name: ValidateSenderID :one
SELECT id
FROM sender_ids
WHERE service_provider_id = $1
  AND sender_id_string = $2
  AND status = 'approved'
LIMIT 1;

-- name: GetTemplateContent :one
SELECT content
FROM templates
WHERE id = $1 AND status = 'active'
  AND (service_provider_id IS NULL OR service_provider_id = $2); -- Allow global or SP-specific template

-- ### OTP Alternative Senders ###

-- name: CreateOtpAlternativeSender :one
INSERT INTO otp_alternative_senders (
    sender_id_string, mno_id, status, max_usage_count, reset_interval_hours, notes
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

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

-- name: GetOtpAlternativeSenderByID :one
SELECT * FROM otp_alternative_senders WHERE id = $1;

-- ### OTP Message Templates ###

-- name: CreateOtpMessageTemplate :one
INSERT INTO otp_message_templates (
    name, content_template, default_brand_name, status
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetOtpMessageTemplateByID :one
SELECT * FROM otp_message_templates WHERE id = $1;

-- CRUD for otp_message_templates (List, Update, Delete) as needed...


-- ### OTP Sender Template Assignments ###

-- name: CreateOtpSenderTemplateAssignment :one
INSERT INTO otp_sender_template_assignments (
    otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

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

-- CRUD for otp_sender_template_assignments (List, Update, Delete) as needed...

