-- +goose Up
-- +goose StatementBegin

-- =========================================
-- ROUTE HEALTH: ONE ROUTE-WIDE ROW PER ASSIGNMENT
-- =========================================

-- UNIQUE (routing_assignment_id, sender_id) treats NULLs as distinct, so the
-- UpsertRouteHealth ON CONFLICT never matched route-wide rows (sender_id NULL) and every
-- update inserted another row. The router joins on that row, so a route could have several
-- rows with conflicting availability. Keep the most recent row per assignment and make the
-- constraint treat NULLs as equal (Postgres 15+) so the upsert updates in place.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'route_health' AND column_name = 'routing_assignment_id'
    ) THEN
        RAISE NOTICE 'route_health has no routing_assignment_id column, skipping';
        RETURN;
    END IF;

    DELETE FROM route_health rh
    USING route_health newer
    WHERE rh.sender_id IS NULL
      AND newer.sender_id IS NULL
      AND rh.routing_assignment_id = newer.routing_assignment_id
      AND (rh.updated_at, rh.id) < (newer.updated_at, newer.id);

    IF current_setting('server_version_num')::int >= 150000 THEN
        ALTER TABLE route_health DROP CONSTRAINT IF EXISTS route_health_routing_assignment_id_sender_id_key;
        ALTER TABLE route_health
            ADD CONSTRAINT route_health_routing_assignment_id_sender_id_key
            UNIQUE NULLS NOT DISTINCT (routing_assignment_id, sender_id);
    ELSE
        RAISE NOTICE 'Postgres older than 15: duplicates removed, but route-wide route_health rows can still be duplicated';
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DO $$
BEGIN
    IF current_setting('server_version_num')::int >= 150000 AND EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'route_health' AND column_name = 'routing_assignment_id'
    ) THEN
        ALTER TABLE route_health DROP CONSTRAINT IF EXISTS route_health_routing_assignment_id_sender_id_key;
        ALTER TABLE route_health
            ADD CONSTRAINT route_health_routing_assignment_id_sender_id_key
            UNIQUE (routing_assignment_id, sender_id);
    END IF;
END $$;

-- +goose StatementEnd
