-- name: UpsertConcatPart :one
-- Stores a part; a resent part (same group and number, not yet joined) replaces the stored copy and
-- keeps its message ID.
INSERT INTO incoming_concat_parts (
    service_provider_id, sp_credential_id, source_addr, destination_addr, concat_ref, total_parts,
    part_num, content, client_message_id, request_dlr, is_flash
) VALUES (
    sqlc.arg(service_provider_id), sqlc.arg(sp_credential_id), sqlc.arg(source_addr), sqlc.arg(destination_addr),
    sqlc.arg(concat_ref), sqlc.arg(total_parts), sqlc.arg(part_num), sqlc.arg(content),
    sqlc.arg(client_message_id), sqlc.arg(request_dlr), sqlc.arg(is_flash)
)
ON CONFLICT (sp_credential_id, source_addr, destination_addr, concat_ref, total_parts, part_num) WHERE claimed_at IS NULL
DO UPDATE SET content = EXCLUDED.content, received_at = NOW()
RETURNING client_message_id;

-- name: ClaimCompleteConcatGroup :many
-- Atomically claims a group's parts once all of them are in. A concurrent call for the same group
-- waits on the row locks, then finds them claimed and returns nothing, so only one caller joins it.
UPDATE incoming_concat_parts p
SET claimed_at = NOW()
WHERE p.sp_credential_id = sqlc.arg(sp_credential_id) AND p.source_addr = sqlc.arg(source_addr)
  AND p.destination_addr = sqlc.arg(destination_addr) AND p.concat_ref = sqlc.arg(concat_ref)
  AND p.total_parts = sqlc.arg(total_parts) AND p.claimed_at IS NULL
  AND (SELECT count(*) FROM incoming_concat_parts q
       WHERE q.sp_credential_id = p.sp_credential_id AND q.source_addr = p.source_addr
         AND q.destination_addr = p.destination_addr AND q.concat_ref = p.concat_ref
         AND q.total_parts = p.total_parts AND q.claimed_at IS NULL) >= p.total_parts
RETURNING p.id, p.part_num, p.content, p.client_message_id;

-- name: ClaimStaleConcatGroup :many
-- Claims the parts of one incomplete group whose first part is older than the cutoff, so what
-- arrived is still sent. Picks the oldest such group.
WITH g AS (
    SELECT sp_credential_id, source_addr, destination_addr, concat_ref, total_parts
    FROM incoming_concat_parts
    WHERE claimed_at IS NULL
    GROUP BY sp_credential_id, source_addr, destination_addr, concat_ref, total_parts
    HAVING min(received_at) < sqlc.arg(cutoff)::TIMESTAMPTZ
    ORDER BY min(received_at)
    LIMIT 1
)
UPDATE incoming_concat_parts p
SET claimed_at = NOW()
FROM g
WHERE p.claimed_at IS NULL AND p.sp_credential_id = g.sp_credential_id AND p.source_addr = g.source_addr
  AND p.destination_addr = g.destination_addr AND p.concat_ref = g.concat_ref AND p.total_parts = g.total_parts
RETURNING p.id, p.service_provider_id, p.sp_credential_id, p.source_addr, p.destination_addr,
          p.total_parts, p.part_num, p.content, p.client_message_id, p.request_dlr, p.is_flash, p.received_at;

-- name: SetConcatPartsMessage :exec
UPDATE incoming_concat_parts SET message_id = sqlc.arg(message_id) WHERE id = ANY(sqlc.arg(ids)::BIGINT[]);

-- name: ReleaseConcatParts :exec
-- Gives claimed parts back when the joined message could not be created, so it is retried.
UPDATE incoming_concat_parts SET claimed_at = NULL WHERE id = ANY(sqlc.arg(ids)::BIGINT[]) AND message_id IS NULL;

-- name: DeleteOldConcatParts :execrows
-- Retention: joined parts are only kept for a while.
DELETE FROM incoming_concat_parts WHERE message_id IS NOT NULL AND received_at < sqlc.arg(cutoff)::TIMESTAMPTZ;
