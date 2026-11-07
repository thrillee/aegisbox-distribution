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
    link_kind, scan_mode, mode_reason, status, completed_at, priority, next_attempt_at
) VALUES (
    sqlc.arg(message_id), sqlc.arg(service_provider_id), sqlc.arg(url), sqlc.arg(normalized_url),
    sqlc.arg(host), sqlc.arg(group_key), sqlc.arg(link_kind), sqlc.arg(scan_mode),
    sqlc.narg(mode_reason), sqlc.arg(status)::TEXT,
    CASE WHEN sqlc.arg(status)::TEXT = 'skipped' THEN NOW() ELSE NULL END,
    sqlc.arg(priority)::SMALLINT,
    -- A random wait before the link is opened (LINKSCAN_VISIT_DELAY_MIN/MAX); 0 means now.
    NOW() + sqlc.arg(delay_ms)::INT * INTERVAL '1 millisecond'
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
    -- Jobs of held real messages (priority 1) first.
    ORDER BY c.priority DESC, c.next_attempt_at
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
RETURNING j.*;

-- name: CompleteLinkScanJob :exec
UPDATE link_scan_jobs
SET status = 'done', link_scan_id = sqlc.narg(link_scan_id), last_error = NULL, locked_at = NULL,
    completed_at = NOW(), updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'processing';

-- name: FailLinkScanJob :exec
-- A failed attempt: retried after base, 2 x base, 4 x base seconds... until max_attempts, then
-- failed. The worker uses a 30 s base, and a short one for jobs of held messages.
UPDATE link_scan_jobs
SET status = CASE WHEN attempts >= sqlc.arg(max_attempts)::INT THEN 'failed' ELSE 'queued' END,
    last_error = sqlc.arg(last_error),
    locked_at = NULL,
    next_attempt_at = NOW() + make_interval(secs => sqlc.arg(base_seconds)::INT * power(2, GREATEST(attempts - 1, 0))),
    completed_at = CASE WHEN attempts >= sqlc.arg(max_attempts)::INT THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'processing';

-- name: DeferLinkScanJob :exec
-- The scanner is down, busy or not configured: not the page's fault, so the attempt the claim
-- counted is given back and the job waits a minute. Links are never lost while the scanner is out.
UPDATE link_scan_jobs
SET status = 'queued', attempts = GREATEST(attempts - 1, 0), last_error = sqlc.arg(last_error),
    locked_at = NULL, next_attempt_at = NOW() + INTERVAL '1 minute', updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'processing';

-- name: ListLinkScanJobsForMessage :many
SELECT * FROM link_scan_jobs WHERE message_id = sqlc.arg(message_id) ORDER BY id;

-- name: InsertLinkScan :one
INSERT INTO link_scans (
    normalized_url, group_key, scan_kind, final_url, final_host, http_status, verdict, score,
    reasons, title, text_excerpt, redirects, forms, downloads, blocked, tls_issuer,
    screenshot_path, scan_error, duration_ms, expires_at
) VALUES (
    sqlc.arg(normalized_url), sqlc.arg(group_key), sqlc.arg(scan_kind), sqlc.narg(final_url),
    sqlc.narg(final_host), sqlc.narg(http_status), sqlc.arg(verdict), sqlc.arg(score),
    sqlc.arg(reasons), sqlc.narg(title), sqlc.narg(text_excerpt), sqlc.arg(redirects),
    sqlc.arg(forms), sqlc.arg(downloads), sqlc.arg(blocked), sqlc.narg(tls_issuer),
    sqlc.narg(screenshot_path), sqlc.narg(scan_error), sqlc.narg(duration_ms), sqlc.arg(expires_at)
)
RETURNING id;

-- name: FindCachedLinkScan :one
-- A recent visit of the same URL that is still valid (24 h, 7 days when malicious). verdict is
-- the effective one (an admin override wins).
SELECT id, COALESCE(override_verdict, verdict)::TEXT AS verdict, final_host FROM link_scans
WHERE normalized_url = sqlc.arg(normalized_url) AND scan_kind = 'visit' AND expires_at > NOW()
ORDER BY scanned_at DESC
LIMIT 1;

-- name: FindRecentGroupVisit :one
-- The latest visit of a link in the same group (per recipient links that differ only by a token),
-- used for links that were sampled out instead of visited.
SELECT id, COALESCE(override_verdict, verdict)::TEXT AS verdict, final_host FROM link_scans
WHERE group_key = sqlc.arg(group_key) AND scan_kind = 'visit' AND expires_at > NOW()
ORDER BY scanned_at DESC
LIMIT 1;

-- name: ListOldScreenshots :many
SELECT id, screenshot_path FROM link_scans
WHERE screenshot_path IS NOT NULL AND scanned_at < sqlc.arg(before)
ORDER BY id
LIMIT 500;

-- name: ClearScreenshotPath :exec
UPDATE link_scans SET screenshot_path = NULL WHERE id = sqlc.arg(id);

-- name: DeleteOldLinkScanData :one
-- Deletes scan results, finished jobs and extraction records older than the cutoff. Returns the
-- number of scan results deleted.
-- Jobs still queued by then (scanning was off) are deleted too: their messages are too old to matter.
WITH j AS (
    DELETE FROM link_scan_jobs lj
    WHERE lj.created_at < sqlc.arg(before) AND lj.status <> 'processing'
), e AS (
    DELETE FROM message_link_extractions WHERE extracted_at < sqlc.arg(before)
), h AS (
    DELETE FROM message_link_holds mh
    WHERE mh.created_at < sqlc.arg(before) AND mh.status IN ('released', 'rejected')
), s AS (
    DELETE FROM link_scans WHERE scanned_at < sqlc.arg(before) AND screenshot_path IS NULL
    RETURNING id
)
SELECT COUNT(*)::INT FROM s;

-- name: CountLinkScanBacklog :one
SELECT COUNT(*)::INT AS queued,
       COALESCE(EXTRACT(EPOCH FROM NOW() - MIN(created_at)), 0)::INT AS oldest_seconds
FROM link_scan_jobs
WHERE status IN ('queued', 'processing');

-- name: GetSimulatorMessageForLinkScan :one
-- A message whose segments went out through a simulator MNO, for queuing its links once it is
-- delivered. No row: it was not simulator traffic.
SELECT m.service_provider_id, m.short_message
FROM messages m
WHERE m.id = sqlc.arg(message_id)
  AND EXISTS (
      SELECT 1 FROM message_segments ms
      JOIN mno_connections c ON c.id = ms.mno_connection_id
      JOIN mnos n ON n.id = c.mno_id
      WHERE ms.message_id = m.id AND n.is_simulator
  );
