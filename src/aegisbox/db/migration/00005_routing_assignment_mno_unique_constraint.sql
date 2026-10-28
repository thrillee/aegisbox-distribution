-- +goose Up
-- +goose StatementBegin

-- =========================================
-- ROUTING ASSIGNMENT UNIQUE CONSTRAINT UPDATE
-- =========================================
-- This migration allows multiple MNOs to be configured for the same 
-- routing_group + msisdn_prefix combination (required for MNO pool routing)
--
-- Old constraint: routing_assignments_routing_group_id_msisdn_prefix_group_id_key
--   UNIQUE (routing_group_id, msisdn_prefix_group_id)
--
-- New constraint: routing_assignments_routing_group_id_mno_id_msisdn_prefix_group_id_key
--   UNIQUE (routing_group_id, mno_id, msisdn_prefix_group_id)

-- Step 1: Drop old unique constraint (if it exists)
ALTER TABLE routing_assignments 
DROP CONSTRAINT IF EXISTS routing_assignments_routing_group_id_msisdn_prefix_group_id_key;

-- Step 2: Add new unique constraint with mno_id
ALTER TABLE routing_assignments 
ADD CONSTRAINT routing_assignments_routing_group_id_mno_id_msisdn_prefix_group_id_key 
UNIQUE (routing_group_id, mno_id, msisdn_prefix_group_id);

-- Step 3: Ensure is_active column exists (from migration 00003)
ALTER TABLE routing_assignments 
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

ALTER TABLE routing_assignments 
ADD COLUMN IF NOT EXISTS weight INT DEFAULT 100 CHECK (weight >= 0 AND weight <= 100);

ALTER TABLE routing_assignments 
ADD COLUMN IF NOT EXISTS max_retry_attempts INT DEFAULT 2;

-- Step 4: Update comments
COMMENT ON COLUMN routing_assignments.is_active IS 'Quick toggle for route without changing status';
COMMENT ON COLUMN routing_assignments.weight IS 'Load balancing weight (0-100). Higher weight gets more traffic within same priority';
COMMENT ON COLUMN routing_assignments.max_retry_attempts IS 'Max retry attempts for this specific route before trying next route';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverse the changes

-- Step 1: Drop new unique constraint
ALTER TABLE routing_assignments 
DROP CONSTRAINT IF EXISTS routing_assignments_routing_group_id_mno_id_msisdn_prefix_group_id_key;

-- Step 2: Restore old unique constraint (without mno_id)
ALTER TABLE routing_assignments 
ADD CONSTRAINT routing_assignments_routing_group_id_msisdn_prefix_group_id_key 
UNIQUE (routing_group_id, msisdn_prefix_group_id);

-- Note: We keep is_active, weight, max_retry_attempts columns as they were added in 00003

-- +goose StatementEnd
