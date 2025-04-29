-- name: GetWalletForUpdate :one
SELECT id, balance
FROM wallets
WHERE service_provider_id = $1 AND currency_code = $2
FOR UPDATE; -- Lock the row for atomic update

-- name: UpdateWalletBalance :one
UPDATE wallets
SET balance = $1,
    updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: CreateWalletTransaction :one
INSERT INTO wallet_transactions (
    wallet_id, message_id, transaction_type, amount, balance_before, balance_after, description, reference_transaction_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetWalletsBelowThreshold :many
SELECT w.id as wallet_id, w.balance, w.low_balance_threshold, w.currency_code, sp.id as service_provider_id, sp.email
FROM wallets w
JOIN service_providers sp ON w.service_provider_id = sp.id
WHERE w.balance < w.low_balance_threshold
  AND (w.low_balance_notified_at IS NULL OR w.low_balance_notified_at < NOW() - INTERVAL '24 hours'); -- Check threshold and notification time

-- name: UpdateLowBalanceNotifiedAt :exec
UPDATE wallets
SET low_balance_notified_at = NOW()
WHERE id = $1;

-- name: FindDebitTransactionForMessage :one
SELECT id, wallet_id, amount
FROM wallet_transactions
WHERE message_id = $1
  AND transaction_type = 'debit'
LIMIT 1;

-- name: GetWalletForUpdateByID :one
SELECT *
FROM wallets
WHERE id = $1
FOR UPDATE;

-- name: ListWalletsBySP :many
-- Lists wallets for a specific Service Provider, paginated.
SELECT * FROM wallets
WHERE service_provider_id = $1
ORDER BY currency_code
LIMIT $2 OFFSET $3;

-- name: CountWalletsBySP :one
-- Counts wallets for a specific Service Provider.
SELECT count(*) FROM wallets
WHERE service_provider_id = $1;

-- name: GetWalletByID :one
-- Gets a specific wallet by its primary ID.
SELECT * FROM wallets WHERE id = $1 LIMIT 1;

