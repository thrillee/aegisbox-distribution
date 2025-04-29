-- name: CreateRoutingRule :one
INSERT INTO routing_rules (
    prefix, mno_id, priority
) VALUES (
    $1, $2, COALESCE($3, 0) -- Default priority to 0 if not provided
) RETURNING *;

-- name: GetRoutingRuleByID :one
-- Join with MNOs to get the name for display
SELECT rr.*, m.name as mno_name
FROM routing_rules rr
LEFT JOIN mnos m ON rr.mno_id = m.id
WHERE rr.id = $1 LIMIT 1;

-- name: ListRoutingRules :many
-- Lists routing rules, optionally filtered by MNO ID, with pagination.
SELECT rr.*, m.name as mno_name
FROM routing_rules rr
LEFT JOIN mnos m ON rr.mno_id = m.id
WHERE ($1::INT IS NULL OR rr.mno_id = $1) -- Optional MNO ID filter
ORDER BY length(rr.prefix) DESC, rr.priority ASC, rr.id -- Consistent ordering
LIMIT $2 OFFSET $3;

-- name: CountRoutingRules :one
-- Counts routing rules, optionally filtered by MNO ID.
SELECT count(*)
FROM routing_rules
WHERE ($1::INT IS NULL OR mno_id = $1); -- Optional MNO ID filter


-- name: UpdateRoutingRule :one
UPDATE routing_rules
SET
    prefix = COALESCE(sqlc.narg(prefix), prefix),
    mno_id = COALESCE(sqlc.narg(mno_id), mno_id),
    priority = COALESCE(sqlc.narg(priority), priority),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;
-- Note: If updating mno_id, ensure the new MNO exists (handled by FK or check in handler)

-- name: DeleteRoutingRule :exec
DELETE FROM routing_rules WHERE id = $1;

-- Already exists: GetApplicableRoutingRule (used by core logic)
