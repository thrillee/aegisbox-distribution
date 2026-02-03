-- Migration: Add retry tracking and message retry support
-- Run this manually: psql -d your_database -f 00006_retry_tracking_run.sql
-- Or use: make db/migrate (if Docker is running)

BEGIN;

-- Add retry tracking columns to dlr_forwarding_queue table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'dlr_forwarding_queue' AND column_name = 'next_retry_at'
    ) THEN
        ALTER TABLE dlr_forwarding_queue
        ADD COLUMN next_retry_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Add retry tracking columns to messages table for message-level retry
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'messages' AND column_name = 'retry_count'
    ) THEN
        ALTER TABLE messages
        ADD COLUMN retry_count INTEGER DEFAULT 0 NOT NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'messages' AND column_name = 'next_retry_at'
    ) THEN
        ALTER TABLE messages
        ADD COLUMN next_retry_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Create dead_letter_queue table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'dead_letter_queue'
    ) THEN
        CREATE TABLE dead_letter_queue (
            id SERIAL PRIMARY KEY,
            message_id BIGINT NOT NULL,
            original_status VARCHAR(50),
            error_code VARCHAR(50),
            error_description TEXT,
            retry_count INTEGER DEFAULT 0,
            failed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            metadata JSONB
        );

        -- Indexes
        CREATE INDEX idx_dead_letter_queue_message_id ON dead_letter_queue(message_id);
        CREATE INDEX idx_dead_letter_queue_failed_at ON dead_letter_queue(failed_at);
    END IF;
END $$;

-- Add indexes for efficient polling (if they don't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes WHERE indexname = 'idx_dlr_forward_queue_next_retry'
    ) THEN
        CREATE INDEX idx_dlr_forward_queue_next_retry
        ON dlr_forwarding_queue(next_retry_at)
        WHERE status = 'pending';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes WHERE indexname = 'idx_messages_next_retry'
    ) THEN
        CREATE INDEX idx_messages_next_retry
        ON messages(next_retry_at)
        WHERE processing_status = 'retry_pending';
    END IF;
END $$;

COMMIT;

-- Verify migration
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'dlr_forwarding_queue'
ORDER BY ordinal_position;

SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'messages'
AND column_name IN ('retry_count', 'next_retry_at')
ORDER BY column_name;

SELECT table_name FROM information_schema.tables WHERE table_name = 'dead_letter_queue';
