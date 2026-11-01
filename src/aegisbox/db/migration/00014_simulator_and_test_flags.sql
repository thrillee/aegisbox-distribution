-- +goose Up
-- +goose StatementBegin

-- A simulator MNO (aegisbox mnosim) reports messages as sent and delivered without sending them.
-- Only customers marked as test may be routed to one; the gateway refuses everyone else.
ALTER TABLE mnos ADD COLUMN IF NOT EXISTS is_simulator BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE service_providers ADD COLUMN IF NOT EXISTS is_test BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE service_providers DROP COLUMN IF EXISTS is_test;
ALTER TABLE mnos DROP COLUMN IF EXISTS is_simulator;
-- +goose StatementEnd
