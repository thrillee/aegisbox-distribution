-- name: CreateSMPPSession :one
INSERT INTO smpp_sessions (sp_credential_id, service_provider_id, system_id, bind_type, remote_addr, status, ended_at, end_reason)
VALUES (sqlc.narg(sp_credential_id), sqlc.narg(service_provider_id), sqlc.arg(system_id), sqlc.arg(bind_type), sqlc.arg(remote_addr), sqlc.arg(status)::text,
        CASE WHEN sqlc.arg(status)::text = 'refused' THEN NOW() END, sqlc.narg(end_reason))
RETURNING id;

-- name: EndSMPPSession :exec
UPDATE smpp_sessions
SET status = 'closed', ended_at = NOW(), end_reason = sqlc.arg(end_reason)
WHERE id = sqlc.arg(id) AND status = 'connected';

-- name: CloseOrphanedSMPPSessions :execrows
-- Sessions left 'connected' by a gateway that stopped without closing them.
UPDATE smpp_sessions
SET status = 'closed', ended_at = NOW(), end_reason = 'gateway restarted'
WHERE status = 'connected';
