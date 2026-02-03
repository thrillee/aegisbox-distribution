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
