-- name: GetApplicableRoutingRule :one
SELECT mno_id
FROM routing_rules
WHERE $1 LIKE prefix || '%' -- Use LIKE for prefix matching (e.g., '234803...' LIKE '234803%')
ORDER BY length(prefix) DESC, priority ASC -- Match longest prefix first, then by priority
LIMIT 1;
