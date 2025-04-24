-- name: EnqueueDLRForForwarding :exec
-- Inserts a new DLR forwarding job into the queue.
INSERT INTO dlr_forwarding_queue (
    message_id, payload, max_attempts, created_at, status
) VALUES (
    $1, $2, $3, NOW(), 'pending'
);

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
-- Marks a job as failed for this attempt, increments attempts, and potentially sets to permanent failure.
UPDATE dlr_forwarding_queue
SET status = CASE WHEN attempts + 1 >= max_attempts THEN 'failed' ELSE 'pending' END, -- Set to 'failed' on max attempts, else 'pending' for retry
    attempts = attempts + 1,
    error_message = $1, -- Error message from the attempt
    last_attempt_at = NOW(),
    locked_at = NULL, -- Unlock for next attempt or final state
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
