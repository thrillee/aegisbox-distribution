-- name: IsSimulatorMNO :one
SELECT COALESCE((SELECT is_simulator FROM mnos WHERE id = sqlc.arg(mno_id)), FALSE)::BOOLEAN AS is_simulator;

-- name: RecordMessageLinkExtraction :exec
-- Records that a message was checked for links (also when it had none). Written after its jobs;
-- a second send attempt after failover leaves the first record.
INSERT INTO message_link_extractions (message_id, outcome, link_count)
VALUES (sqlc.arg(message_id), sqlc.arg(outcome), sqlc.arg(link_count))
ON CONFLICT (message_id) DO NOTHING;

-- name: CountRecentVisitsForGroup :one
-- Visits queued in the last hour for links of the same customer that share a group key, used to
-- visit only a sample of per recipient links.
SELECT COUNT(*)::INT AS visits
FROM link_scan_jobs
WHERE service_provider_id = sqlc.arg(service_provider_id)
  AND group_key = sqlc.arg(group_key)
  AND scan_mode = 'visit'
  AND created_at > NOW() - INTERVAL '1 hour';

-- name: InsertLinkScanJob :exec
INSERT INTO link_scan_jobs (
    message_id, service_provider_id, url, normalized_url, host, group_key,
    link_kind, scan_mode, mode_reason, status, completed_at
) VALUES (
    sqlc.arg(message_id), sqlc.arg(service_provider_id), sqlc.arg(url), sqlc.arg(normalized_url),
    sqlc.arg(host), sqlc.arg(group_key), sqlc.arg(link_kind), sqlc.arg(scan_mode),
    sqlc.narg(mode_reason), sqlc.arg(status)::TEXT,
    CASE WHEN sqlc.arg(status)::TEXT = 'skipped' THEN NOW() ELSE NULL END
)
ON CONFLICT (message_id, normalized_url) DO NOTHING;

-- name: ClaimLinkScanJobs :many
-- Takes due jobs for scanning. A job left in processing for 5 minutes (a crashed worker) is taken
-- again.
UPDATE link_scan_jobs j
SET status = 'processing', attempts = j.attempts + 1, locked_at = NOW(), updated_at = NOW()
WHERE j.id IN (
    SELECT c.id FROM link_scan_jobs c
    WHERE (c.status = 'queued' AND c.next_attempt_at <= NOW())
       OR (c.status = 'processing' AND c.locked_at < NOW() - INTERVAL '5 minutes')
    ORDER BY c.next_attempt_at
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
RETURNING j.*;

-- name: CompleteLinkScanJob :exec
UPDATE link_scan_jobs
SET status = 'done', last_error = NULL, locked_at = NULL, completed_at = NOW(), updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'processing';

-- name: FailLinkScanJob :exec
-- A failed attempt: retried after 30 s, 60 s, 120 s... until max_attempts, then failed.
UPDATE link_scan_jobs
SET status = CASE WHEN attempts >= sqlc.arg(max_attempts)::INT THEN 'failed' ELSE 'queued' END,
    last_error = sqlc.arg(last_error),
    locked_at = NULL,
    next_attempt_at = NOW() + make_interval(secs => 30 * power(2, GREATEST(attempts - 1, 0))),
    completed_at = CASE WHEN attempts >= sqlc.arg(max_attempts)::INT THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'processing';

-- name: ListLinkScanJobsForMessage :many
SELECT * FROM link_scan_jobs WHERE message_id = sqlc.arg(message_id) ORDER BY id;
