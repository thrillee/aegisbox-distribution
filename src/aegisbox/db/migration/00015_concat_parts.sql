-- +goose Up
-- +goose StatementBegin

-- Parts of a long SMS sent over SMPP as separate submit_sm PDUs with a concatenation UDH. Each part
-- is answered with its own message ID; when all parts are in (or after a timeout) they are joined
-- into one message, so the text is sent, billed and reported once.
CREATE TABLE IF NOT EXISTS incoming_concat_parts (
    id BIGSERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    sp_credential_id INT NOT NULL REFERENCES sp_credentials(id) ON DELETE CASCADE,
    source_addr VARCHAR(32) NOT NULL,
    destination_addr VARCHAR(32) NOT NULL,
    concat_ref INT NOT NULL,
    total_parts INT NOT NULL,
    part_num INT NOT NULL,
    content TEXT NOT NULL,
    client_message_id VARCHAR(64) NOT NULL,
    request_dlr BOOLEAN NOT NULL DEFAULT TRUE,
    is_flash BOOLEAN NOT NULL DEFAULT FALSE,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    claimed_at TIMESTAMPTZ,  -- set while the joined message is being created
    message_id BIGINT REFERENCES messages(id) ON DELETE SET NULL
);
-- One pending copy of each part per group (a resent part replaces it).
CREATE UNIQUE INDEX IF NOT EXISTS uq_incoming_concat_parts_pending
ON incoming_concat_parts(sp_credential_id, source_addr, destination_addr, concat_ref, total_parts, part_num)
WHERE claimed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_incoming_concat_parts_unclaimed ON incoming_concat_parts(received_at) WHERE claimed_at IS NULL;

-- The customer's message IDs for the other parts of a joined message; each gets the message's DLR.
ALTER TABLE messages ADD COLUMN IF NOT EXISTS extra_client_message_ids TEXT[];

-- A message's final real DLR is forwarded once, when its last segment gets a final status, even if
-- two segment DLRs arrive at the same moment: the job carries a unique key.
ALTER TABLE dlr_forwarding_queue ADD COLUMN IF NOT EXISTS dedupe_key TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS uq_dlr_forwarding_queue_dedupe_key ON dlr_forwarding_queue(dedupe_key) WHERE dedupe_key IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_dlr_forwarding_queue_dedupe_key;
ALTER TABLE dlr_forwarding_queue DROP COLUMN IF EXISTS dedupe_key;
ALTER TABLE messages DROP COLUMN IF EXISTS extra_client_message_ids;
DROP TABLE IF EXISTS incoming_concat_parts;
-- +goose StatementEnd
