-- +goose Up
-- +goose StatementBegin

-- Table for grouping MSISDN prefixes (e.g., MTN Nigeria, Vodafone Ghana)
CREATE TABLE msisdn_prefix_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    reference VARCHAR(100) UNIQUE, -- User-friendly reference (e.g., MTNNG_PFX)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE msisdn_prefix_groups IS 'Groups of MSISDN prefixes, e.g., "MTN Nigeria Prefixes".';
COMMENT ON COLUMN msisdn_prefix_groups.reference IS 'A unique, user-friendly reference for the prefix group.';
CREATE INDEX idx_msisdn_prefix_groups_reference ON msisdn_prefix_groups(reference);

-- Table for individual MSISDN prefixes belonging to a group
CREATE TABLE msisdn_prefix_entries (
    id SERIAL PRIMARY KEY,
    msisdn_prefix_group_id INT NOT NULL REFERENCES msisdn_prefix_groups(id) ON DELETE CASCADE,
    msisdn_prefix VARCHAR(20) NOT NULL, -- The actual prefix, e.g., 234803
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (msisdn_prefix_group_id, msisdn_prefix),
    UNIQUE (msisdn_prefix) -- A prefix should ideally be unique globally to avoid ambiguity, or this constraint can be removed if prefixes can exist in different groups for different reasons (though routing logic would need care). For now, let's assume unique.
);
COMMENT ON TABLE msisdn_prefix_entries IS 'Individual MSISDN prefixes belonging to a specific msisdn_prefix_group.';
COMMENT ON COLUMN msisdn_prefix_entries.msisdn_prefix IS 'The MSISDN prefix string, e.g., "234703".';
-- Index for efficient prefix matching (LIKE 'prefix%')
CREATE INDEX idx_msisdn_prefix_entries_prefix_pattern ON msisdn_prefix_entries (msisdn_prefix varchar_pattern_ops);
CREATE INDEX idx_msisdn_prefix_entries_group_id ON msisdn_prefix_entries(msisdn_prefix_group_id);

-- Table for defining logical routing groups/strategies (e.g., "Standard Route", "Premium Route")
CREATE TABLE routing_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    reference VARCHAR(100) UNIQUE, -- User-friendly reference (e.g., STD_ROUTE_NG)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE routing_groups IS 'Defines a collection of routing rules forming a specific routing strategy.';
COMMENT ON COLUMN routing_groups.reference IS 'A unique, user-friendly reference for the routing group.';
CREATE INDEX idx_routing_groups_reference ON routing_groups(reference);

-- Add routing_group_id to sp_credentials to link them to a specific RoutingGroup
-- This allows different credentials (even for the same SP) to have different routing behaviors.
ALTER TABLE sp_credentials
ADD COLUMN routing_group_id INT REFERENCES routing_groups(id) ON DELETE SET NULL;
COMMENT ON COLUMN sp_credentials.routing_group_id IS 'The routing group assigned to this specific SP credential. NULL means system default may apply.';

CREATE INDEX idx_sp_credentials_routing_group_id ON sp_credentials(routing_group_id);

-- Add priority to mno_connections
ALTER TABLE mno_connections
ADD COLUMN priority INT default 1;

-- New table for routing rules, linking an (MsisdnPrefixGroup within a RoutingGroup) to an MNO
-- This table effectively replaces the old routing_rules functionality with more structure.
CREATE TABLE routing_assignments (
    id SERIAL PRIMARY KEY,
    routing_group_id INT NOT NULL REFERENCES routing_groups(id) ON DELETE CASCADE,
    msisdn_prefix_group_id INT NOT NULL REFERENCES msisdn_prefix_groups(id) ON DELETE RESTRICT, -- Prevent deleting a prefix group if it's used in an active rule
    mno_id INT NOT NULL REFERENCES mnos(id) ON DELETE RESTRICT, -- Prevent deleting an MNO if it's used in an active rule
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- e.g., active, inactive, testing
    priority INT NOT NULL DEFAULT 0, -- Lower number means higher priority for this assignment if multiple prefix groups match within the same routing group (less common scenario, primary sorting is on prefix length)
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (routing_group_id, msisdn_prefix_group_id) -- Within a routing strategy, a prefix group should map to one MNO. For failover, consider more advanced features or multiple rules with different priorities/statuses.
);
COMMENT ON TABLE routing_assignments IS 'Assigns an MNO to a specific MSISDN prefix group within a given routing group.';
COMMENT ON COLUMN routing_assignments.status IS 'Status of the routing assignment (e.g., active, inactive, testing).';
COMMENT ON COLUMN routing_assignments.priority IS 'Priority of this rule within the routing group for the same prefix group (lower is higher).';

CREATE INDEX idx_routing_assignments_routing_group_id ON routing_assignments(routing_group_id);
CREATE INDEX idx_routing_assignments_msisdn_prefix_group_id ON routing_assignments(msisdn_prefix_group_id);
CREATE INDEX idx_routing_assignments_mno_id ON routing_assignments(mno_id);
CREATE INDEX idx_routing_assignments_status ON routing_assignments(status);
CREATE INDEX idx_routing_assignments_rgroup_pgroup_status ON routing_assignments(routing_group_id, msisdn_prefix_group_id, status);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop in reverse order of creation, handling dependencies

ALTER TABLE sp_credentials DROP COLUMN IF EXISTS routing_group_id; -- Drop index idx_sp_credentials_routing_group_id automatically

ALTER TABLE mno_connections DROP COLUMN IF EXISTS priority;

DROP TABLE IF EXISTS routing_assignments; -- Drop indexes automatically
DROP TABLE IF EXISTS routing_groups; -- Drop index idx_routing_groups_reference automatically
DROP TABLE IF EXISTS msisdn_prefix_entries; -- Drop indexes automatically
DROP TABLE IF EXISTS msisdn_prefix_groups; -- Drop index idx_msisdn_prefix_groups_reference automatically

-- +goose StatementEnd
