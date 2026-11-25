-- +goose Up
-- +goose StatementBegin

-- Per network sender rules: when a provider's message goes to a number in the given prefix group
-- (for example MTN Nigeria), it is submitted with replacement_sender instead. original_sender
-- limits the rule to messages sent with that sender; NULL means any sender. A rule naming the
-- sender wins over a NULL rule. Rules apply before the route sender pool.
CREATE TABLE IF NOT EXISTS sender_network_overrides (
    id SERIAL PRIMARY KEY,
    service_provider_id INT NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    msisdn_prefix_group_id INT NOT NULL REFERENCES msisdn_prefix_groups(id) ON DELETE CASCADE,
    original_sender VARCHAR(16),
    replacement_sender VARCHAR(16) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sender_network_overrides_rule
    ON sender_network_overrides (service_provider_id, msisdn_prefix_group_id, (UPPER(COALESCE(original_sender, ''))));
CREATE INDEX IF NOT EXISTS idx_sender_network_overrides_sp_status
    ON sender_network_overrides (service_provider_id, status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sender_network_overrides;
-- +goose StatementEnd
