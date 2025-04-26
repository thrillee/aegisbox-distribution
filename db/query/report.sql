-- name: GetDailyReport :many
SELECT
    -- Grouping Keys
    DATE(m.submitted_at AT TIME ZONE 'UTC') AS report_date, -- Group by date part of submitted_at in UTC
    sp.id AS service_provider_id,
    sp.name AS service_provider_name,
    sp.email AS service_provider_email,
    COALESCE(mn.id, 0) AS mno_id,
    COALESCE(mn.name, 'N/A') AS mno_name,
    m.currency_code,
    -- Aggregations
    COUNT(m.id) AS total_received,
    COALESCE(SUM(m.total_segments), 0) AS total_segments_received,
    COALESCE(SUM(CASE WHEN m.processing_status NOT IN ('received', 'failed_validation', 'failed_routing', 'failed_pricing') THEN m.total_segments ELSE 0 END), 0) AS total_segments_submitted,
    COALESCE(SUM(CASE WHEN m.final_status = 'delivered' THEN m.total_segments ELSE 0 END), 0) AS total_segments_delivered,
    COALESCE(SUM(CASE WHEN m.final_status = 'failed' THEN m.total_segments ELSE 0 END), 0) AS total_segments_failed,
    COALESCE(SUM(CASE WHEN m.final_status = 'rejected' THEN m.total_segments ELSE 0 END), 0) AS total_segments_rejected,
    -- Use the stored cost directly
    COALESCE(SUM(m.cost), 0.0) AS net_cost -- Sum the stored cost
FROM messages m
JOIN service_providers sp ON m.service_provider_id = sp.id
LEFT JOIN mnos mn ON m.routed_mno_id = mn.id
-- Filter by date range (inclusive)
WHERE m.submitted_at AT TIME ZONE 'UTC' >= sqlc.arg(start_date)::timestamptz -- Start of day UTC
  AND m.submitted_at AT TIME ZONE 'UTC' < sqlc.arg(end_date)::timestamptz + INTERVAL '1 day' -- Up to, but not including, start of next day UTC
  AND ($1::INT IS NULL OR m.service_provider_id = $1) -- Filter by sp_id if provided
  AND ($2::TEXT IS NULL OR sp.email = $2) -- Filter by sp_email if provided
GROUP BY
    report_date,
    sp.id,
    mn.id,
    m.currency_code
ORDER BY
    report_date DESC,
    sp.name;
