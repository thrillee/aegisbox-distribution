-- name: EnqueueDLRForForwarding :exec
-- Inserts a new DLR forwarding job into the queue.
INSERT INTO dlr_forwarding_queue (
    message_id, payload, max_attempts, created_at, status
) VALUES (
    $1, $2, $3, NOW(), 'pending'
);

-- name: GetPendingDLRsToForwardWithBackoff :many
-- Selects pending DLR jobs that are due for retry (next_retry_at <= NOW()), plus jobs left
-- in 'processing' for over 5 minutes by a worker that crashed or failed to record the result,
-- so no DLR is stranded. A recovered job may be forwarded twice (at least once delivery).
WITH candidates AS (
    SELECT id
    FROM dlr_forwarding_queue
    WHERE (
        (status = 'pending'
          AND (next_retry_at IS NULL OR next_retry_at <= NOW())
          AND (locked_at IS NULL OR locked_at < NOW() - INTERVAL '5 minutes'))
        OR (status = 'processing' AND locked_at < NOW() - INTERVAL '5 minutes')
      )
      AND attempts < max_attempts
    ORDER BY next_retry_at ASC NULLS FIRST, created_at ASC
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
UPDATE dlr_forwarding_queue q
SET
    status = 'processing',
    locked_at = NOW(),
    locked_by = $2
FROM candidates c
WHERE q.id = c.id
RETURNING q.id, q.message_id, q.payload, q.attempts, q.max_attempts;

-- name: GetPendingDLRsToForward :many
-- Selects pending DLR jobs and locks them for processing.
WITH candidates AS (
    SELECT id
    FROM dlr_forwarding_queue
    WHERE status = 'pending'
      AND (locked_at IS NULL OR locked_at < NOW() - INTERVAL '5 minutes') -- Pick unlocked or stale locks
      AND attempts < max_attempts
    ORDER BY created_at
    LIMIT $1 -- Batch size
    FOR UPDATE SKIP LOCKED
)
UPDATE dlr_forwarding_queue q
SET
    status = 'processing',
    locked_at = NOW(),
    locked_by = $2 -- Worker identifier
FROM candidates c
WHERE q.id = c.id
RETURNING q.id, q.message_id, q.payload, q.attempts, q.max_attempts;

-- name: MarkDLRForwardingSuccess :exec
-- Marks a job as successfully completed.
UPDATE dlr_forwarding_queue
SET status = 'success',
    error_message = NULL,
    last_attempt_at = NOW(),
    locked_at = NULL, -- Unlock
    locked_by = NULL
WHERE id = $1;

-- name: MarkDLRForwardingAttemptFailed :exec
-- Marks a job as failed for this attempt, increments attempts, sets next_retry_at for exponential backoff.
-- Calculates backoff: least(greatest(5 * 2^(attempts), 5 seconds), 5 minutes) seconds
UPDATE dlr_forwarding_queue
SET status = CASE WHEN attempts + 1 >= max_attempts THEN 'failed' ELSE 'pending' END,
    attempts = attempts + 1,
    error_message = $1,
    last_attempt_at = NOW(),
    next_retry_at = CASE
        WHEN attempts + 1 >= max_attempts THEN NULL
        ELSE NOW() + (LEAST(GREATEST(5 * (2 ^ attempts), 5), 300) || ' seconds')::interval
    END,
    locked_at = NULL,
    locked_by = NULL
WHERE id = $2;

-- name: UnlockStaleDLRs :exec
-- Optional: Periodically run to unlock jobs held by dead workers.
UPDATE dlr_forwarding_queue
SET status = 'pending',
    locked_at = NULL,
    locked_by = NULL
WHERE status = 'processing'
  AND locked_at < NOW() - INTERVAL '15 minutes'; -- Adjust stale timeout

-- name: RequeueFailedDLRForwarding :many
-- Puts failed DLR forwarding jobs back in the queue with a fresh set of attempts.
UPDATE dlr_forwarding_queue
SET status = 'pending',
    attempts = 0,
    max_attempts = GREATEST(max_attempts, 10), -- Older jobs were queued with fewer attempts
    next_retry_at = NULL,
    error_message = NULL,
    locked_at = NULL,
    locked_by = NULL
WHERE id = ANY(sqlc.arg(ids)::BIGINT[])
  AND status = 'failed'
RETURNING id;
