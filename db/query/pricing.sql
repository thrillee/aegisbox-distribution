-- name: GetApplicablePrice :one
SELECT price_per_sms, currency_code
FROM pricing_rules
WHERE service_provider_id = $1
  AND currency_code = $2
  AND (mno_id = $3 OR mno_id IS NULL) -- Match specific MNO or default
ORDER BY mno_id DESC NULLS LAST -- Prioritize specific MNO rule over default (NULL)
LIMIT 1;

-- name: CreatePricingRule :one
INSERT INTO pricing_rules (
    service_provider_id, mno_id, currency_code, price_per_sms
) VALUES (
    $1, $2, $3, $4 -- $2 is sql.NullInt32, $4 is decimal.Decimal
) RETURNING *;

-- name: GetPricingRuleByID :one
-- Join MNO Name for display
SELECT pr.*, m.name as mno_name
FROM pricing_rules pr
LEFT JOIN mnos m ON pr.mno_id = m.id
WHERE pr.id = $1 LIMIT 1;

-- name: ListPricingRulesBySP :many
-- Lists pricing rules for a specific Service Provider, paginated.
SELECT pr.*, m.name as mno_name
FROM pricing_rules pr
LEFT JOIN mnos m ON pr.mno_id = m.id
WHERE pr.service_provider_id = $1 -- Required SP ID filter
ORDER BY pr.mno_id NULLS FIRST, pr.currency_code -- Show defaults first
LIMIT $2 OFFSET $3;

-- name: CountPricingRulesBySP :one
-- Counts pricing rules for a specific Service Provider.
SELECT count(*)
FROM pricing_rules
WHERE service_provider_id = $1;

-- name: DeletePricingRule :exec
DELETE FROM pricing_rules WHERE id = $1;
