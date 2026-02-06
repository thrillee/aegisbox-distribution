# Migration Guide - Retry Tracking & Performance Fixes

## Apply Migrations

### Option 1: Using Docker (Recommended)
```bash
make db/migrate
```

### Option 2: Manual Execution
```bash
# Run the migration script directly with psql
psql -d your_database -f db/migration/00006_retry_tracking_run.sql
```

### Verify Migration
```sql
-- Check new columns in messages table
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'messages' 
AND column_name IN ('retry_count', 'next_retry_at');

-- Check DLR queue has next_retry_at
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'dlr_forwarding_queue' 
AND column_name = 'next_retry_at';

-- Check dead_letter_queue exists
SELECT COUNT(*) FROM dead_letter_queue;
```

## New Features

### 1. Circuit Breaker for MNO Connections
- Location: `internal/mno/circuit_breaker.go`
- Logs: `circuit_breaker_state_change` events
- States: closed → half-open → open
- Default thresholds: 5 failures → open, 30s timeout → half-open

### 2. Proactive Reconnection
- Location: `internal/mno/manager.go`
- Runs every 10 seconds
- Exponential backoff: 5s → 5min max
- Logs: `Reconnection attempt`, `Reconnection failed`, `Reconnection successful`

### 3. Message Retry with Exponential Backoff
- Retry delay: `LEAST(GREATEST(5 * 2^attempts, 5), 300) seconds`
- Max retries: 5 per message
- Dead letter queue for permanent failures

### 4. DLR Retry with Exponential Backoff
- Same exponential backoff formula
- Max retries: 10 per DLR job
- Automatic next_retry_at calculation

### 5. DLR-Based Wallet Reversal
- Automatic credit reversal when DLR indicates failure
- Logs: `CRITICAL: Failed to reverse wallet debit on DLR failure`

### 6. Rate Limiting
- Token bucket per MNO
- Configurable: 50 req/s, 100 burst
- Logs: `Rate limit exceeded for MNO`

## Key Log Events

| Event | Level | Description |
|-------|-------|-------------|
| `circuit_breaker_state_change` | INFO | State transitions (open/closed/half-open) |
| `Reconnection attempt` | INFO | MNO reconnection retry |
| `Reconnection failed` | WARN | Reconnection error |
| `Reconnection successful` | INFO | Reconnection succeeded |
| `Rate limit exceeded for MNO` | WARN | MNO throttled, message queued for retry |
| `DLR forwarding batch completed` | INFO | Batch processing summary |
| `Wallet debit reversed due to DLR failure` | INFO | Automatic refund issued |
| `CRITICAL: Failed to reverse wallet debit` | ERROR | Refund failed - manual intervention needed |

## Monitoring Queries

```sql
-- View circuit breaker states
SELECT * FROM mno_connections WHERE status != 'bound';

-- Check messages pending retry
SELECT id, retry_count, next_retry_at, error_description
FROM messages
WHERE processing_status = 'retry_pending'
ORDER BY next_retry_at ASC;

-- Check dead letter queue
SELECT * FROM dead_letter_queue ORDER BY failed_at DESC LIMIT 100;

-- Check DLR jobs pending retry
SELECT id, message_id, attempts, next_retry_at, error_message
FROM dlr_forwarding_queue
WHERE status = 'pending' AND next_retry_at IS NOT NULL
ORDER BY next_retry_at ASC;
```

## Rollback (if needed)

```sql
-- Remove dead_letter_queue (optional, safe to leave)
DROP TABLE IF EXISTS dead_letter_queue;

-- Remove retry columns (NOT RECOMMENDED - data loss)
ALTER TABLE messages DROP COLUMN IF EXISTS retry_count;
ALTER TABLE messages DROP COLUMN IF EXISTS next_retry_at;
ALTER TABLE dlr_forwarding_queue DROP COLUMN IF EXISTS next_retry_at;

-- Remove indexes
DROP INDEX IF EXISTS idx_dlr_forward_queue_next_retry;
DROP INDEX IF EXISTS idx_messages_next_retry;
DROP INDEX IF EXISTS idx_dead_letter_queue_message_id;
DROP INDEX IF EXISTS idx_dead_letter_queue_failed_at;
```
