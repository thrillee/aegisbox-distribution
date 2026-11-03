-- +goose Up
-- +goose StatementBegin

-- Link scanning step 4 (rest) and step 5: admin allow and block lists, verdict overrides, domain
-- ages from RDAP, and a log of who looked at scan results.

-- Admin lists. A domain matches itself and its subdomains. service_provider_id NULL applies to
-- every customer; a customer's own entry wins over a global one, and block wins over allow.
CREATE TABLE IF NOT EXISTS link_domain_lists (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(253) NOT NULL,        -- lowercase, punycode, no scheme or path
    list_type VARCHAR(10) NOT NULL CHECK (list_type IN ('allow', 'block')),
    service_provider_id INT REFERENCES service_providers(id) ON DELETE CASCADE,
    note TEXT,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_link_domain_lists
ON link_domain_lists(domain, COALESCE(service_provider_id, 0));

-- An admin's correction of a result (false positive or missed threat). The effective verdict is
-- COALESCE(override_verdict, verdict).
ALTER TABLE link_scans ADD COLUMN IF NOT EXISTS override_verdict VARCHAR(20);
ALTER TABLE link_scans ADD COLUMN IF NOT EXISTS override_note TEXT;
ALTER TABLE link_scans ADD COLUMN IF NOT EXISTS override_by TEXT;
ALTER TABLE link_scans ADD COLUMN IF NOT EXISTS override_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_link_scans_final_host ON link_scans(final_host);

-- Registration dates of registrable domains, from RDAP. status: known, unknown (no RDAP for the
-- TLD or no registration date; not scored), error (retried soon).
CREATE TABLE IF NOT EXISTS link_domain_ages (
    domain VARCHAR(253) PRIMARY KEY,
    registered_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Who opened scan details in the portal (message text, numbers and screenshots are personal data).
CREATE TABLE IF NOT EXISTS link_scan_view_log (
    id BIGSERIAL PRIMARY KEY,
    viewer TEXT NOT NULL,
    message_id BIGINT,
    link_scan_id BIGINT,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_link_scan_view_log_viewed_at ON link_scan_view_log(viewed_at);

-- Jobs of messages waiting on a verdict (Phase 2 hold and block) are scanned first.
ALTER TABLE link_scan_jobs ADD COLUMN IF NOT EXISTS priority SMALLINT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_link_scan_jobs_message ON link_scan_jobs(message_id);

-- Phase 2: per customer scanning of real traffic. No row means off. Nothing changes for a
-- customer until an admin sets a mode.
--   off      no scanning of real messages
--   monitor  links are scanned after the message is sent; never delays or blocks
--   hold     the message waits up to hold_seconds for verdicts; a bad verdict sends it to admin
--            review (release or reject in the portal)
--   block    the message waits up to hold_seconds; a bad verdict rejects it (REJECTD DLR)
CREATE TABLE IF NOT EXISTS link_scan_settings (
    service_provider_id INT PRIMARY KEY REFERENCES service_providers(id) ON DELETE CASCADE,
    mode VARCHAR(10) NOT NULL DEFAULT 'off' CHECK (mode IN ('off', 'monitor', 'hold', 'block')),
    hold_seconds INT NOT NULL DEFAULT 60 CHECK (hold_seconds BETWEEN 5 AND 600),
    -- When no verdict arrives in time (scanner down or slow): true sends the message anyway,
    -- false treats it as bad (review in hold mode, rejected in block mode).
    fail_open BOOLEAN NOT NULL DEFAULT TRUE,
    -- Treat suspicious like malicious (default: only malicious stops a message).
    block_suspicious BOOLEAN NOT NULL DEFAULT FALSE,
    -- Keep the charge for rejected messages (default: the debit is reversed).
    charge_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    updated_by TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One row per real message held for link verdicts.
--   waiting   verdicts pending (message processing_status 'held_link_scan')
--   review    a bad verdict in hold mode, waiting for an admin (rejected after 24 hours)
--   released  sent on (decision says why)
--   rejected  not sent; REJECTD DLR forwarded
CREATE TABLE IF NOT EXISTS message_link_holds (
    message_id BIGINT PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    mode VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    decision VARCHAR(30),       -- clean, timeout_released, admin_released, malicious, suspicious,
                                -- timeout_rejected, admin_rejected, review_expired
    worst_verdict VARCHAR(20),
    hold_until TIMESTAMPTZ NOT NULL,
    fail_open BOOLEAN NOT NULL,
    block_suspicious BOOLEAN NOT NULL,
    charge_blocked BOOLEAN NOT NULL,
    decided_by TEXT,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_message_link_holds_open ON message_link_holds(status, hold_until)
WHERE status IN ('waiting', 'review');
CREATE INDEX IF NOT EXISTS idx_message_link_holds_created_at ON message_link_holds(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS message_link_holds;
DROP TABLE IF EXISTS link_scan_settings;
DROP INDEX IF EXISTS idx_link_scan_jobs_message;
ALTER TABLE link_scan_jobs DROP COLUMN IF EXISTS priority;
DROP TABLE IF EXISTS link_scan_view_log;
DROP TABLE IF EXISTS link_domain_ages;
DROP INDEX IF EXISTS idx_link_scans_final_host;
ALTER TABLE link_scans DROP COLUMN IF EXISTS override_at;
ALTER TABLE link_scans DROP COLUMN IF EXISTS override_by;
ALTER TABLE link_scans DROP COLUMN IF EXISTS override_note;
ALTER TABLE link_scans DROP COLUMN IF EXISTS override_verdict;
DROP TABLE IF EXISTS link_domain_lists;
-- +goose StatementEnd
