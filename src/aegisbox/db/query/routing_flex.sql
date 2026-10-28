-- Queries for MsisdnPrefixGroup

-- name: CreateMsisdnPrefixGroup :one
INSERT INTO msisdn_prefix_groups (
    name, description, reference
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetMsisdnPrefixGroupByID :one
SELECT * FROM msisdn_prefix_groups
WHERE id = $1 LIMIT 1;

-- name: GetMsisdnPrefixGroupByReference :one
SELECT * FROM msisdn_prefix_groups
WHERE reference = $1 LIMIT 1;

-- name: ListMsisdnPrefixGroups :many
SELECT * FROM msisdn_prefix_groups
ORDER BY name
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountMsisdnPrefixGroups :one
SELECT count(*) FROM msisdn_prefix_groups;

-- name: UpdateMsisdnPrefixGroup :one
UPDATE msisdn_prefix_groups
SET
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    reference = COALESCE(sqlc.narg(reference), reference),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMsisdnPrefixGroup :exec
-- Note: This will fail if the group is referenced in routing_assignments due to ON DELETE RESTRICT.
-- Handle this in application logic (e.g., unassign first or provide a force option).
DELETE FROM msisdn_prefix_groups WHERE id = $1;


-- Queries for MsisdnPrefixEntry

-- name: CreateMsisdnPrefixEntry :one
INSERT INTO msisdn_prefix_entries (
    msisdn_prefix_group_id, msisdn_prefix
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetMsisdnPrefixEntryByID :one
SELECT * FROM msisdn_prefix_entries
WHERE id = $1 LIMIT 1;

-- name: ListMsisdnPrefixEntriesByGroup :many
SELECT mpe.*
FROM msisdn_prefix_entries mpe
WHERE mpe.msisdn_prefix_group_id = $1
ORDER BY mpe.msisdn_prefix
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountMsisdnPrefixEntriesByGroup :one
SELECT count(*)
FROM msisdn_prefix_entries
WHERE msisdn_prefix_group_id = $1;

-- name: GetMsisdnPrefixEntryByPrefix :one
SELECT *
FROM msisdn_prefix_entries
WHERE msisdn_prefix = $1 LIMIT 1;

-- name: UpdateMsisdnPrefixEntry :one
UPDATE msisdn_prefix_entries
SET
    msisdn_prefix = COALESCE(sqlc.narg(msisdn_prefix), msisdn_prefix),
    -- msisdn_prefix_group_id = COALESCE(sqlc.narg(msisdn_prefix_group_id), msisdn_prefix_group_id), -- Typically you wouldn't change the group_id, but delete and recreate
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMsisdnPrefixEntry :exec
DELETE FROM msisdn_prefix_entries WHERE id = $1;

-- name: DeleteMsisdnPrefixEntriesByGroup :exec
DELETE FROM msisdn_prefix_entries WHERE msisdn_prefix_group_id = $1;

-- name: FindMatchingPrefixGroupForMSISDN :one
-- Finds the msisdn_prefix_group_id for a given MSISDN by longest prefix match.
-- $1: full_msisdn (e.g., '2348031234567')
SELECT
    mpe.msisdn_prefix_group_id,
    mpe.msisdn_prefix AS matched_prefix
FROM msisdn_prefix_entries mpe
WHERE $1 LIKE mpe.msisdn_prefix || '%'
ORDER BY length(mpe.msisdn_prefix) DESC
LIMIT 1;


-- Queries for RoutingGroup

-- name: CreateRoutingGroup :one
INSERT INTO routing_groups (
    name, description, reference
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetRoutingGroupByID :one
SELECT * FROM routing_groups
WHERE id = $1 LIMIT 1;

-- name: GetRoutingGroupByReference :one
SELECT * FROM routing_groups
WHERE reference = $1 LIMIT 1;

-- name: ListRoutingGroups :many
SELECT * FROM routing_groups
ORDER BY name
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountRoutingGroups :one
SELECT count(*) FROM routing_groups;

-- name: UpdateRoutingGroup :one
UPDATE routing_groups
SET
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    reference = COALESCE(sqlc.narg(reference), reference),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteRoutingGroup :exec
-- Note: This will fail if the group is referenced in routing_assignments or sp_credentials.
DELETE FROM routing_groups WHERE id = $1;

-- name: AssignRoutingGroupToSpCredential :one
UPDATE sp_credentials
SET routing_group_id = sqlc.arg(routing_group_id), updated_at = NOW()
WHERE id = sqlc.arg(sp_credential_id)
RETURNING *;

-- name: RemoveRoutingGroupFromSpCredential :one
UPDATE sp_credentials
SET routing_group_id = NULL, updated_at = NOW()
WHERE id = sqlc.arg(sp_credential_id)
RETURNING *;

-- name: GetRoutingGroupIdForSpCredential :one
-- Returns the routing_group_id (can be NULL) and the sp_credential_id itself for context.
SELECT id as sp_credential_id, routing_group_id
FROM sp_credentials
WHERE id = $1;


-- Queries for RoutingAssignment (the new routing rules)

-- name: CreateRoutingAssignment :one
INSERT INTO routing_assignments (
    routing_group_id, msisdn_prefix_group_id, mno_id, status, priority, comment
) VALUES (
    $1, $2, $3, COALESCE(sqlc.narg(status), 'active'), COALESCE(sqlc.narg(priority), 0), sqlc.narg(comment)
) RETURNING *;

-- name: GetRoutingAssignmentByID :one
SELECT
    ra.*,
    rg.name as routing_group_name,
    rg.reference as routing_group_reference,
    mpg.name as prefix_group_name,
    mpg.reference as prefix_group_reference,
    m.name as mno_name
FROM routing_assignments ra
JOIN routing_groups rg ON ra.routing_group_id = rg.id
JOIN msisdn_prefix_groups mpg ON ra.msisdn_prefix_group_id = mpg.id
JOIN mnos m ON ra.mno_id = m.id
WHERE ra.id = $1 LIMIT 1;

-- name: ListRoutingAssignmentsByRoutingGroup :many
-- Lists assignments for a specific routing group.
SELECT
    ra.*,
    mpg.name as prefix_group_name,
    mpg.reference as prefix_group_reference,
    m.name as mno_name
FROM routing_assignments ra
JOIN msisdn_prefix_groups mpg ON ra.msisdn_prefix_group_id = mpg.id
JOIN mnos m ON ra.mno_id = m.id
WHERE ra.routing_group_id = $1
ORDER BY ra.priority ASC, mpg.name ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountRoutingAssignmentsByRoutingGroup :one
SELECT count(*)
FROM routing_assignments
WHERE routing_group_id = $1;

-- name: GetRoutingAssignmentByRoutingGroupAndPrefixGroup :one
SELECT ra.*
FROM routing_assignments ra
WHERE ra.routing_group_id = $1 AND ra.msisdn_prefix_group_id = $2
LIMIT 1;

-- name: UpdateRoutingAssignment :one
UPDATE routing_assignments
SET
    -- routing_group_id = COALESCE(sqlc.narg(routing_group_id), routing_group_id), -- Usually not changed, delete/recreate
    -- msisdn_prefix_group_id = COALESCE(sqlc.narg(msisdn_prefix_group_id), msisdn_prefix_group_id), -- Usually not changed
    mno_id = COALESCE(sqlc.narg(mno_id), mno_id),
    status = COALESCE(sqlc.narg(status), status),
    priority = COALESCE(sqlc.narg(priority), priority),
    comment = COALESCE(sqlc.narg(comment), comment),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
-- Can also add routing_group_id to WHERE if updates are always scoped to a group passed in context
-- WHERE id = sqlc.arg(id) AND routing_group_id = sqlc.arg(routing_group_id_context)
RETURNING *;

-- name: DeleteRoutingAssignment :exec
DELETE FROM routing_assignments WHERE id = $1;


-- name: GetApplicableMnoForRouting :one
-- Core routing query:
-- Given a routing_group_id (from sp_credential or a default) and an msisdn_prefix_group_id (from MSISDN lookup),
-- find the active MNO to route to. Returns MNO with protocol preference.
-- $1: routing_group_id
-- $2: msisdn_prefix_group_id
SELECT
    ra.id AS routing_assignment_id,
    ra.mno_id,
    m.name AS mno_name,
    mc.id AS mno_connection_id,
    mc.system_id AS mno_system_id,
    mc.protocol AS mno_protocol
FROM routing_assignments ra
JOIN mnos m ON ra.mno_id = m.id
LEFT JOIN mno_connections mc ON m.id = mc.mno_id AND mc.status = 'active'
WHERE ra.routing_group_id = $1
  AND ra.msisdn_prefix_group_id = $2
  AND ra.status = 'active'
ORDER BY
    ra.priority ASC,
    mc.priority ASC NULLS LAST,
    mc.updated_at DESC
LIMIT 1;

-- name: GetSpCredentialRoutingGroupId :one
SELECT routing_group_id FROM sp_credentials WHERE id = $1;

-- name: GetRoutingGroupReferenceById :one
SELECT reference FROM routing_groups WHERE id = $1;


-- =========================================
-- INTELLIGENT ROUTING QUERIES
-- =========================================

-- name: GetAllApplicableRoutesForRouting :many
-- Returns ALL active routes for intelligent route selection (not just one)
-- Ordered by priority, health, and weight for best-first selection
-- $1: routing_group_id
-- $2: msisdn_prefix_group_id
SELECT
    ra.id AS routing_assignment_id,
    ra.mno_id,
    ra.priority,
    ra.weight,
    ra.max_retry_attempts,
    ra.status AS assignment_status,
    ra.is_active,
    m.name AS mno_name,
    m.status AS mno_status,
    mc.id AS mno_connection_id,
    mc.protocol AS mno_protocol,
    mc.status AS connection_status,
    mc.priority AS connection_priority,
    COALESCE(rh.health_score, 100.00) AS route_health_score,
    COALESCE(rh.is_available, true) AS route_is_available,
    COALESCE(rh.consecutive_failures, 0) AS consecutive_failures,
    COALESCE(rh.last_success_at, NOW()) AS last_route_success
FROM routing_assignments ra
JOIN mnos m ON ra.mno_id = m.id AND m.status = 'active'
LEFT JOIN mno_connections mc ON m.id = mc.mno_id AND mc.status = 'active'
LEFT JOIN route_health rh ON ra.id = rh.routing_assignment_id AND rh.sender_id IS NULL
WHERE ra.routing_group_id = $1
  AND ra.msisdn_prefix_group_id = $2
  AND ra.status = 'active'
  AND ra.is_active = true
ORDER BY
    ra.priority ASC,
    route_health_score DESC,
    ra.weight DESC,
    mc.priority ASC NULLS LAST,
    mc.updated_at DESC;

-- name: UpdateRoutingAssignmentFields :one
-- Update routing assignment with new intelligent routing fields
UPDATE routing_assignments
SET
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    weight = COALESCE(sqlc.narg(weight), weight),
    max_retry_attempts = COALESCE(sqlc.narg(max_retry_attempts), max_retry_attempts),
    priority = COALESCE(sqlc.narg(priority), priority),
    status = COALESCE(sqlc.narg(status), status),
    comment = COALESCE(sqlc.narg(comment), comment),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateMessageRoutingAssignment :exec
-- Track which route was used for a message
UPDATE messages
SET routing_assignment_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateSegmentRoutingAssignment :exec
-- Track which route was used for a segment
UPDATE message_segments
SET routing_assignment_id = $2
WHERE id = $1;


-- =========================================
-- SENDER ID HEALTH QUERIES
-- =========================================

-- name: RecordSenderDeliveryAttempt :one
-- Record individual delivery attempt in ring buffer
INSERT INTO sender_id_delivery_attempts (
    sender_id, mno_id, message_segment_id, status, error_code,
    sent_at, dlr_received_at, recorded_date
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetSenderIdHealthForMNO :one
-- Get sender ID health for a specific MNO
SELECT *
FROM sender_id_health
WHERE sender_id = $1 AND mno_id = $2
LIMIT 1;

-- name: CalculateSenderHealthFromRingBuffer :one
-- Calculate health metrics from last N attempts in ring buffer
-- $1: sender_id
-- $2: mno_id  
-- $3: window_size (e.g., 1000)
WITH recent_attempts AS (
    SELECT status
    FROM sender_id_delivery_attempts
    WHERE sender_id = $1 
      AND mno_id = $2
    ORDER BY recorded_at DESC
    LIMIT $3
),
aggregated AS (
    SELECT
        COALESCE(COUNT(CASE WHEN status = 'delivered' THEN 1 END), 0) as delivered,
        COALESCE(COUNT(CASE WHEN status = 'failed' THEN 1 END), 0) as failed,
        COALESCE(COUNT(CASE WHEN status = 'no_dlr' THEN 1 END), 0) as no_dlr,
        COALESCE(COUNT(*), 0) as total
    FROM recent_attempts
)
SELECT
    total::INT as total_sent,
    delivered::INT as total_delivered,
    failed::INT as total_failed,
    no_dlr::INT as total_no_dlr,
    CASE WHEN total > 0 THEN (delivered::NUMERIC / total::NUMERIC * 100.0)::NUMERIC(5,2) ELSE 100.00 END as delivery_rate,
    CASE WHEN total > 0 THEN ((delivered::NUMERIC + failed::NUMERIC) / total::NUMERIC * 100.0)::NUMERIC(5,2) ELSE 100.00 END as dlr_response_rate,
    CASE WHEN total > 0 THEN 
        ((delivered::NUMERIC / total::NUMERIC * 0.7 + (delivered::NUMERIC + failed::NUMERIC) / total::NUMERIC * 0.3) * 100.0)::NUMERIC(5,2)
    ELSE 100.00 END as health_score
FROM aggregated;

-- name: UpsertSenderIdHealth :one
-- Create or update sender ID health record
INSERT INTO sender_id_health (
    sender_id, mno_id, total_sent, total_delivered, total_failed, total_no_dlr,
    delivery_rate, dlr_response_rate, health_score, is_blocked, blocked_reason,
    auto_blocked, last_success_at, last_failure_at, last_no_dlr_at, last_attempt_at,
    window_size, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())
ON CONFLICT (sender_id, mno_id) 
DO UPDATE SET
    total_sent = EXCLUDED.total_sent,
    total_delivered = EXCLUDED.total_delivered,
    total_failed = EXCLUDED.total_failed,
    total_no_dlr = EXCLUDED.total_no_dlr,
    delivery_rate = EXCLUDED.delivery_rate,
    dlr_response_rate = EXCLUDED.dlr_response_rate,
    health_score = EXCLUDED.health_score,
    is_blocked = EXCLUDED.is_blocked,
    blocked_reason = EXCLUDED.blocked_reason,
    blocked_at = CASE WHEN EXCLUDED.is_blocked AND NOT sender_id_health.is_blocked THEN NOW() ELSE sender_id_health.blocked_at END,
    auto_blocked = EXCLUDED.auto_blocked,
    last_success_at = COALESCE(EXCLUDED.last_success_at, sender_id_health.last_success_at),
    last_failure_at = COALESCE(EXCLUDED.last_failure_at, sender_id_health.last_failure_at),
    last_no_dlr_at = COALESCE(EXCLUDED.last_no_dlr_at, sender_id_health.last_no_dlr_at),
    last_attempt_at = COALESCE(EXCLUDED.last_attempt_at, sender_id_health.last_attempt_at),
    window_size = EXCLUDED.window_size,
    updated_at = NOW()
RETURNING *;

-- name: GetBlockedSendersForAutoUnblock :many
-- Find blocked senders with no activity in 7 days (for auto-unblock)
SELECT 
    sih.*,
    si.sender_id_string,
    si.service_provider_id
FROM sender_id_health sih
JOIN sender_ids si ON sih.sender_id = si.id
WHERE sih.is_blocked = true
  AND sih.auto_blocked = true
  AND sih.last_attempt_at < NOW() - INTERVAL '7 days'
  AND (sih.last_unblock_check_at IS NULL OR sih.last_unblock_check_at < NOW() - INTERVAL '1 day');

-- name: AutoUnblockSenderAfterInactivity :exec
-- Auto-unblock sender after 7 days of inactivity
UPDATE sender_id_health
SET is_blocked = false,
    blocked_reason = 'Auto-unblocked after 7 days of inactivity',
    auto_blocked = false,
    last_unblock_check_at = NOW(),
    updated_at = NOW()
WHERE sender_id = $1 AND mno_id = $2 AND is_blocked = true AND auto_blocked = true;

-- name: ManuallyUnblockSenderIdForMNO :exec
-- Manually unblock a sender ID (admin override)
UPDATE sender_id_health
SET is_blocked = false,
    blocked_reason = NULL,
    auto_blocked = false,
    updated_at = NOW()
WHERE sender_id = $1 AND mno_id = $2;

-- name: ManuallyBlockSenderIdForMNO :exec
-- Manually block a sender ID (admin action)
UPDATE sender_id_health
SET is_blocked = true,
    blocked_at = NOW(),
    blocked_reason = $3,
    auto_blocked = false,
    updated_at = NOW()
WHERE sender_id = $1 AND mno_id = $2;

-- name: ListSenderHealthWithDetails :many
-- List all sender health records with details for dashboard
SELECT 
    sih.*,
    si.sender_id_string,
    si.service_provider_id,
    m.name as mno_name
FROM sender_id_health sih
JOIN sender_ids si ON sih.sender_id = si.id
JOIN mnos m ON sih.mno_id = m.id
ORDER BY sih.health_score ASC, sih.updated_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: GetRecentSenderDeliveryAttempts :many
-- Get recent delivery attempts for a sender-MNO combination
SELECT *
FROM sender_id_delivery_attempts
WHERE sender_id = $1 AND mno_id = $2
ORDER BY recorded_at DESC
LIMIT $3;


-- =========================================
-- ROUTE HEALTH QUERIES
-- =========================================

-- name: RecordRouteDeliveryAttempt :one
-- Record route delivery attempt in ring buffer
INSERT INTO route_delivery_attempts (
    routing_assignment_id, sender_id, message_id, success, error_code, error_description, recorded_date
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetRouteHealth :one
-- Get route health record
SELECT *
FROM route_health
WHERE routing_assignment_id = $1
  AND (sender_id = $2 OR ($2 IS NULL AND sender_id IS NULL))
LIMIT 1;

-- name: CalculateRouteHealthFromRingBuffer :one
-- Calculate route health from last N attempts
-- $1: routing_assignment_id
-- $2: sender_id (nullable)
-- $3: window_size
WITH recent_attempts AS (
    SELECT success
    FROM route_delivery_attempts
    WHERE routing_assignment_id = $1 
      AND (sender_id = $2 OR ($2 IS NULL AND sender_id IS NULL))
    ORDER BY attempted_at DESC
    LIMIT $3
),
aggregated AS (
    SELECT
        COALESCE(COUNT(*), 0) as total,
        COALESCE(COUNT(CASE WHEN success THEN 1 END), 0) as successes
    FROM recent_attempts
)
SELECT
    total::INT as total_attempts,
    successes::INT as successful_attempts,
    (total - successes)::INT as failed_attempts,
    CASE WHEN total > 0 THEN (successes::NUMERIC / total::NUMERIC * 100.0)::NUMERIC(5,2) ELSE 100.00 END as success_rate,
    CASE WHEN total > 0 THEN (successes::NUMERIC / total::NUMERIC * 100.0)::NUMERIC(5,2) ELSE 100.00 END as health_score
FROM aggregated;

-- name: UpsertRouteHealth :one
-- Create or update route health record
INSERT INTO route_health (
    routing_assignment_id, sender_id, total_attempts, successful_attempts, failed_attempts,
    success_rate, health_score, is_available, unavailable_reason, auto_unavailable,
    consecutive_failures, consecutive_successes, last_attempt_at, last_success_at, last_failure_at,
    window_size, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW())
ON CONFLICT (routing_assignment_id, sender_id)
DO UPDATE SET
    total_attempts = EXCLUDED.total_attempts,
    successful_attempts = EXCLUDED.successful_attempts,
    failed_attempts = EXCLUDED.failed_attempts,
    success_rate = EXCLUDED.success_rate,
    health_score = EXCLUDED.health_score,
    is_available = EXCLUDED.is_available,
    unavailable_reason = EXCLUDED.unavailable_reason,
    unavailable_since = CASE WHEN EXCLUDED.is_available THEN NULL ELSE COALESCE(route_health.unavailable_since, NOW()) END,
    auto_unavailable = EXCLUDED.auto_unavailable,
    consecutive_failures = EXCLUDED.consecutive_failures,
    consecutive_successes = EXCLUDED.consecutive_successes,
    last_attempt_at = EXCLUDED.last_attempt_at,
    last_success_at = COALESCE(EXCLUDED.last_success_at, route_health.last_success_at),
    last_failure_at = COALESCE(EXCLUDED.last_failure_at, route_health.last_failure_at),
    window_size = EXCLUDED.window_size,
    updated_at = NOW()
RETURNING *;

-- name: ListRouteHealthWithDetails :many
-- List all route health records with details for dashboard
SELECT 
    rh.*,
    ra.priority,
    ra.weight,
    m.name as mno_name,
    rg.name as routing_group_name,
    mpg.name as prefix_group_name
FROM route_health rh
JOIN routing_assignments ra ON rh.routing_assignment_id = ra.id
JOIN mnos m ON ra.mno_id = m.id
JOIN routing_groups rg ON ra.routing_group_id = rg.id
JOIN msisdn_prefix_groups mpg ON ra.msisdn_prefix_group_id = mpg.id
WHERE rh.sender_id IS NULL  -- Route-level health only
ORDER BY rh.is_available DESC, rh.health_score DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');


-- =========================================
-- ROUTING CONFIGURATION QUERIES
-- =========================================

-- name: GetRoutingConfigForGroup :one
-- Get routing config for a specific group, fallback to global
SELECT * FROM routing_config rc
WHERE rc.routing_group_id = $1
   OR (rc.routing_group_id IS NULL AND NOT EXISTS (SELECT 1 FROM routing_config rc2 WHERE rc2.routing_group_id = $1))
ORDER BY rc.routing_group_id NULLS LAST
LIMIT 1;

-- name: GetGlobalRoutingConfig :one
-- Get global routing configuration
SELECT * FROM routing_config
WHERE routing_group_id IS NULL
LIMIT 1;

-- name: ListGroupRoutingConfigs :many
-- Per routing group configs (the global default row has a NULL group and is read separately)
SELECT rc.*, rg.name AS routing_group_name
FROM routing_config rc
JOIN routing_groups rg ON rg.id = rc.routing_group_id
ORDER BY rg.name;

-- name: GetRoutingConfigByGroupID :one
-- Config of one routing group, without falling back to the global defaults
SELECT rc.*, rg.name AS routing_group_name
FROM routing_config rc
JOIN routing_groups rg ON rg.id = rc.routing_group_id
WHERE rc.routing_group_id = $1;

-- name: UpdateGlobalRoutingConfig :one
-- Update the global default row (routing_group_id IS NULL)
UPDATE routing_config
SET sender_id_block_threshold = sqlc.arg(sender_id_block_threshold),
    sender_id_unblock_threshold = sqlc.arg(sender_id_unblock_threshold),
    sender_id_dlr_timeout_secs = sqlc.arg(sender_id_dlr_timeout_secs),
    sender_id_health_window_size = sqlc.arg(sender_id_health_window_size),
    route_failure_threshold = sqlc.arg(route_failure_threshold),
    route_recovery_threshold = sqlc.arg(route_recovery_threshold),
    route_health_window_size = sqlc.arg(route_health_window_size),
    max_route_retries = sqlc.arg(max_route_retries),
    selection_strategy = sqlc.arg(selection_strategy),
    updated_at = NOW()
WHERE id = (SELECT id FROM routing_config WHERE routing_group_id IS NULL ORDER BY id LIMIT 1)
RETURNING *;

-- name: UpsertRoutingConfig :one
-- Create or update routing configuration
INSERT INTO routing_config (
    routing_group_id, sender_id_block_threshold, sender_id_unblock_threshold,
    sender_id_dlr_timeout_secs, sender_id_health_window_size,
    route_availability_threshold, route_failure_threshold, route_recovery_threshold,
    route_health_window_size, enable_automatic_failover, max_route_retries,
    selection_strategy, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
ON CONFLICT (routing_group_id)
DO UPDATE SET
    sender_id_block_threshold = EXCLUDED.sender_id_block_threshold,
    sender_id_unblock_threshold = EXCLUDED.sender_id_unblock_threshold,
    sender_id_dlr_timeout_secs = EXCLUDED.sender_id_dlr_timeout_secs,
    sender_id_health_window_size = EXCLUDED.sender_id_health_window_size,
    route_availability_threshold = EXCLUDED.route_availability_threshold,
    route_failure_threshold = EXCLUDED.route_failure_threshold,
    route_recovery_threshold = EXCLUDED.route_recovery_threshold,
    route_health_window_size = EXCLUDED.route_health_window_size,
    enable_automatic_failover = EXCLUDED.enable_automatic_failover,
    max_route_retries = EXCLUDED.max_route_retries,
    selection_strategy = EXCLUDED.selection_strategy,
    updated_at = NOW()
RETURNING *;


-- =========================================
-- DLR TIMEOUT QUERIES
-- =========================================

-- name: GetPendingDLRSegments :many
-- Find segments waiting for DLR beyond timeout threshold
SELECT 
    ms.id,
    ms.message_id,
    ms.mno_message_id,
    ms.mno_connection_id,
    ms.sent_to_mno_at,
    ms.routing_assignment_id,
    m.approved_sender_id,
    m.routed_mno_id
FROM message_segments ms
JOIN messages m ON ms.message_id = m.id
WHERE ms.dlr_status IS NULL
  AND ms.sent_to_mno_at IS NOT NULL
  AND ms.sent_to_mno_at < NOW() - make_interval(secs => sqlc.arg(timeout_seconds)::numeric)
  AND m.processing_status NOT IN ('failed_validation', 'failed_routing', 'failed_pricing')
ORDER BY ms.sent_to_mno_at ASC
LIMIT sqlc.arg(max_results);

-- name: MarkSegmentDLRMissing :exec
-- Mark a segment as having missing DLR
UPDATE message_segments
SET dlr_status = 'no_dlr',
    error_description = 'DLR not received within timeout threshold'
WHERE id = $1;


-- =========================================
-- RING BUFFER CLEANUP QUERIES
-- =========================================

-- name: CleanupOldDeliveryAttempts :exec
-- Cleanup sender delivery attempts older than retention period
DELETE FROM sender_id_delivery_attempts
WHERE recorded_date < CURRENT_DATE - make_interval(days => sqlc.arg(retention_days)::int);

-- name: CleanupOldRouteAttempts :exec
-- Cleanup route delivery attempts older than retention period
DELETE FROM route_delivery_attempts
WHERE recorded_date < CURRENT_DATE - make_interval(days => sqlc.arg(retention_days)::int);

-- name: GetRingBufferCleanupConfig :many
-- Get cleanup configuration for ring buffer tables
SELECT * FROM health_ring_buffer_config
WHERE cleanup_enabled = true
ORDER BY table_name;

-- name: UpdateRingBufferCleanupTimestamp :exec
-- Update last cleanup timestamp
UPDATE health_ring_buffer_config
SET last_cleanup_at = NOW(), updated_at = NOW()
WHERE table_name = $1;
