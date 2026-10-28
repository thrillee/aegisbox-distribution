-- +goose Up
-- +goose StatementBegin

-- =========================================
-- ORG MAPPINGS: CLERK ORGANIZATION -> SERVICE PROVIDER
-- =========================================

-- The portal resolves which service provider a signed in Clerk organization belongs to from
-- this table. It was only defined in the portal's ORM schema, so databases built from these
-- migrations lacked it and portal signup and the provider pages failed.
CREATE TABLE IF NOT EXISTS org_mappings (
    id SERIAL PRIMARY KEY,
    clerk_org_id VARCHAR(64) NOT NULL UNIQUE,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A table created earlier by the ORM may be missing updated_at.
ALTER TABLE org_mappings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_org_mappings_service_provider_id ON org_mappings(service_provider_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS org_mappings;

-- +goose StatementEnd
