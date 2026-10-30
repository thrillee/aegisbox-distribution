-- +goose Up
-- +goose StatementBegin

-- The gateway writes error codes longer than 10 characters (MNO_TIMEOUT, MNO_WINDOW_FULL,
-- send_failed, INSUFFICIENT_FUNDS). With VARCHAR(10) those writes failed with "value too long",
-- so submits that timed out or failed were never recorded. Widening never loses data.
ALTER TABLE messages ALTER COLUMN error_code TYPE VARCHAR(50);
ALTER TABLE message_segments ALTER COLUMN error_code TYPE VARCHAR(50);
ALTER TABLE sender_id_delivery_attempts ALTER COLUMN error_code TYPE VARCHAR(50);
ALTER TABLE route_delivery_attempts ALTER COLUMN error_code TYPE VARCHAR(50);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Not narrowed back: rows with longer codes would make it fail.
SELECT 1;
-- +goose StatementEnd
