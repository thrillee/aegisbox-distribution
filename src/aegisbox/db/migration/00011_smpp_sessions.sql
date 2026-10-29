-- +goose Up
-- +goose StatementBegin

-- =========================================
-- SMPP SESSION HISTORY
-- =========================================

-- One row per bind attempt. Successful binds are 'connected' until the connection ends, then
-- 'closed' with ended_at and end_reason. Refused binds are 'refused' with the reason.
CREATE TABLE IF NOT EXISTS smpp_sessions (
    id BIGSERIAL PRIMARY KEY,
    sp_credential_id INT REFERENCES sp_credentials(id) ON DELETE SET NULL,
    service_provider_id INT REFERENCES service_providers(id) ON DELETE SET NULL,
    system_id VARCHAR(64) NOT NULL,
    bind_type VARCHAR(10) NOT NULL,
    remote_addr VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    end_reason VARCHAR(100)
);
CREATE INDEX IF NOT EXISTS idx_smpp_sessions_started_at ON smpp_sessions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_smpp_sessions_sp_started ON smpp_sessions(service_provider_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_smpp_sessions_open ON smpp_sessions(status) WHERE status = 'connected';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS smpp_sessions;
-- +goose StatementEnd
