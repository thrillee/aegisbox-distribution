-- name: GetApplicablePrice :one
SELECT price_per_sms, currency_code
FROM pricing_rules
WHERE service_provider_id = $1
  AND currency_code = $2
  AND (mno_id = $3 OR mno_id IS NULL) -- Match specific MNO or default
ORDER BY mno_id DESC NULLS LAST -- Prioritize specific MNO rule over default (NULL)
LIMIT 1;
