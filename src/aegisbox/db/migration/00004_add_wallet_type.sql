-- +goose Up
-- +goose StatementBegin

-- Add wallet_type column to wallets table
ALTER TABLE wallets
ADD COLUMN IF NOT EXISTS wallet_type VARCHAR(20) DEFAULT 'prepaid' NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE wallets
DROP COLUMN IF EXISTS wallet_type;

-- +goose StatementEnd
