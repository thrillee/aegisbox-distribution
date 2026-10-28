-- +goose Up
-- +goose StatementBegin

-- =========================================
-- OTP ALTERNATIVE SENDERS COMPOSITE UNIQUE
-- =========================================

-- Allow same sender_id_string for different SPs and/or different MNOs
-- Drop existing unique constraint on just sender_id_string
ALTER TABLE otp_alternative_senders 
DROP CONSTRAINT IF EXISTS otp_alternative_senders_sender_id_string_key;

-- Add composite unique constraint on (service_provider_id, sender_id_string, mno_id)
-- This allows:
-- - Same sender_id for different service providers
-- - Same sender_id for different MNOs within same SP
-- - Multiple NULLs in service_provider_id and mno_id (separate rows)
ALTER TABLE otp_alternative_senders 
ADD CONSTRAINT uq_otp_alternative_senders_sp_sender_mno 
UNIQUE (service_provider_id, sender_id_string, mno_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverse: Drop composite constraint and restore original unique on sender_id_string
ALTER TABLE otp_alternative_senders 
DROP CONSTRAINT IF EXISTS uq_otp_alternative_senders_sp_sender_mno;

ALTER TABLE otp_alternative_senders 
ADD CONSTRAINT otp_alternative_senders_sender_id_string_key 
UNIQUE (sender_id_string);

-- +goose StatementEnd
