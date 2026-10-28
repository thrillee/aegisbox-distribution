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

