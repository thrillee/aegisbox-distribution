-- +goose Up
-- +goose StatementBegin

-- Link scanning, step 2: links found in messages sent through a simulator MNO, queued for the
-- link scanner. One row per distinct link per message.
CREATE TABLE IF NOT EXISTS link_scan_jobs (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    url TEXT NOT NULL,             -- as written in the message
    normalized_url TEXT NOT NULL,  -- scheme added, host lowercased and in punycode, no fragment
    host TEXT NOT NULL,
    group_key TEXT NOT NULL,       -- host and path with per recipient tokens replaced by *
    link_kind VARCHAR(20) NOT NULL,   -- web, whatsapp, tel, intent, app_store
    scan_mode VARCHAR(20) NOT NULL,   -- visit (open in a browser), reputation (checks only), none
    mode_reason VARCHAR(30),          -- why not visited: non_web, action_link, sampled
    status VARCHAR(20) NOT NULL DEFAULT 'queued', -- queued, processing, done, failed, skipped
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT uq_link_scan_jobs_message_url UNIQUE (message_id, normalized_url)
);
CREATE INDEX IF NOT EXISTS idx_link_scan_jobs_due ON link_scan_jobs(next_attempt_at) WHERE status IN ('queued', 'processing');
CREATE INDEX IF NOT EXISTS idx_link_scan_jobs_group ON link_scan_jobs(service_provider_id, group_key, created_at);
CREATE INDEX IF NOT EXISTS idx_link_scan_jobs_created_at ON link_scan_jobs(created_at);

-- One row per message checked for links, so "no links" and "skipped as OTP" are recorded too.
CREATE TABLE IF NOT EXISTS message_link_extractions (
    message_id BIGINT PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    outcome VARCHAR(20) NOT NULL, -- links, no_links, skipped_otp
    link_count INT NOT NULL DEFAULT 0,
    extracted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_message_link_extractions_extracted_at ON message_link_extractions(extracted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS message_link_extractions;
DROP TABLE IF EXISTS link_scan_jobs;
-- +goose StatementEnd
