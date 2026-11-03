-- +goose Up
-- +goose StatementBegin

-- Whitelisted numbers: a test customer's message to one of these numbers is not sent to the
-- simulator but through a real network in the customer's routing group, so the phone gets it and
-- the real DLR comes back. service_provider_id NULL applies to every test customer.
CREATE TABLE IF NOT EXISTS simulator_whitelist_numbers (
    id SERIAL PRIMARY KEY,
    msisdn VARCHAR(20) NOT NULL,             -- digits only, international format (2348031234567)
    service_provider_id INT REFERENCES service_providers(id) ON DELETE CASCADE,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_simulator_whitelist_numbers
ON simulator_whitelist_numbers(msisdn, COALESCE(service_provider_id, 0));

-- Link scanning, steps 3 and 4: one row per scan result. Jobs point to the result they used, so
-- several jobs (the same link in many messages) can share one visit.
CREATE TABLE IF NOT EXISTS link_scans (
    id BIGSERIAL PRIMARY KEY,
    normalized_url TEXT NOT NULL,
    group_key TEXT NOT NULL,
    scan_kind VARCHAR(20) NOT NULL,      -- visit (opened in the browser) or static (checks only)
    final_url TEXT,
    final_host TEXT,
    http_status INT,
    verdict VARCHAR(20) NOT NULL,        -- safe, suspicious, malicious
    score INT NOT NULL,
    reasons JSONB NOT NULL DEFAULT '[]', -- [{code, points, detail}]
    title TEXT,
    text_excerpt TEXT,
    redirects JSONB NOT NULL DEFAULT '[]',
    forms JSONB NOT NULL DEFAULT '[]',
    downloads JSONB NOT NULL DEFAULT '[]',
    blocked JSONB NOT NULL DEFAULT '[]', -- requests the SSRF guard refused
    tls_issuer TEXT,
    screenshot_path TEXT,                -- file on the screenshot volume; NULL after retention
    scan_error TEXT,
    duration_ms INT,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL      -- reused for the same URL until then
);
CREATE INDEX IF NOT EXISTS idx_link_scans_url ON link_scans(normalized_url, expires_at);
CREATE INDEX IF NOT EXISTS idx_link_scans_group ON link_scans(group_key, scanned_at);
CREATE INDEX IF NOT EXISTS idx_link_scans_scanned_at ON link_scans(scanned_at);

ALTER TABLE link_scan_jobs ADD COLUMN IF NOT EXISTS link_scan_id BIGINT REFERENCES link_scans(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_link_scan_jobs_link_scan_id ON link_scan_jobs(link_scan_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE link_scan_jobs DROP COLUMN IF EXISTS link_scan_id;
DROP TABLE IF EXISTS link_scans;
DROP TABLE IF EXISTS simulator_whitelist_numbers;
-- +goose StatementEnd
