-- Link scanning: admin lists, domain ages, screenshots and Phase 2 holds.

-- name: ListLinkDomainLists :many
SELECT id, domain, list_type, service_provider_id FROM link_domain_lists ORDER BY id;

-- name: GetLinkDomainAge :one
SELECT domain, registered_at, status, checked_at FROM link_domain_ages WHERE domain = sqlc.arg(domain);

-- name: UpsertLinkDomainAge :exec
INSERT INTO link_domain_ages (domain, registered_at, status, checked_at)
VALUES (sqlc.arg(domain), sqlc.narg(registered_at), sqlc.arg(status), NOW())
ON CONFLICT (domain) DO UPDATE
SET registered_at = EXCLUDED.registered_at, status = EXCLUDED.status, checked_at = NOW();

-- name: GetLinkScanScreenshot :one
SELECT screenshot_path FROM link_scans WHERE id = sqlc.arg(id);

-- name: ListActiveLinkScanSettings :many
-- Customers whose real traffic is scanned (mode other than off).
SELECT service_provider_id, mode, hold_seconds, fail_open, block_suspicious, charge_blocked
FROM link_scan_settings
WHERE mode <> 'off';

-- name: GetMessageTextForLinkScan :one
SELECT short_message FROM messages WHERE id = sqlc.arg(id);

-- name: GetMessageLinkHold :one
SELECT * FROM message_link_holds WHERE message_id = sqlc.arg(message_id);

-- name: InsertMessageLinkHold :execrows
INSERT INTO message_link_holds (
    message_id, service_provider_id, mode, status, hold_until, fail_open, block_suspicious, charge_blocked
) VALUES (
    sqlc.arg(message_id), sqlc.arg(service_provider_id), sqlc.arg(mode), 'waiting',
    NOW() + make_interval(secs => sqlc.arg(hold_seconds)::INT),
    sqlc.arg(fail_open), sqlc.arg(block_suspicious), sqlc.arg(charge_blocked)
)
ON CONFLICT (message_id) DO NOTHING;

-- name: SetMessageProcessingStatus :execrows
-- Moves a message between processing states, only from the expected one.
UPDATE messages SET processing_status = sqlc.arg(to_status)::TEXT
WHERE id = sqlc.arg(id) AND processing_status = sqlc.arg(from_status)::TEXT;

-- name: ListOpenLinkHolds :many
-- Holds to decide: waiting ones (verdicts may be in, or the time is up) and reviews past 24 hours.
SELECT * FROM message_link_holds
WHERE status = 'waiting'
   OR (status = 'review' AND created_at < NOW() - INTERVAL '24 hours')
ORDER BY created_at
LIMIT sqlc.arg(batch_size);

-- name: ListHoldJobVerdicts :many
-- The links of a held message that get a verdict, with the effective verdict of finished ones.
SELECT j.id, j.status, j.normalized_url, s.verdict, s.override_verdict
FROM link_scan_jobs j
LEFT JOIN link_scans s ON s.id = j.link_scan_id
WHERE j.message_id = sqlc.arg(message_id) AND j.scan_mode <> 'none';

-- name: DecideMessageLinkHold :execrows
-- Records a decision, only from the expected status, so two deciders cannot both act.
UPDATE message_link_holds
SET status = sqlc.arg(to_status)::TEXT, decision = sqlc.arg(decision)::TEXT,
    worst_verdict = sqlc.narg(worst_verdict), decided_by = sqlc.narg(decided_by),
    decided_at = CASE WHEN sqlc.arg(to_status)::TEXT IN ('released', 'rejected') THEN NOW() ELSE decided_at END,
    updated_at = NOW()
WHERE message_id = sqlc.arg(message_id) AND status = sqlc.arg(from_status)::TEXT;

-- name: CountOpenLinkHolds :one
SELECT COUNT(*) FILTER (WHERE status = 'waiting')::INT AS waiting,
       COUNT(*) FILTER (WHERE status = 'review')::INT AS review
FROM message_link_holds WHERE status IN ('waiting', 'review');
