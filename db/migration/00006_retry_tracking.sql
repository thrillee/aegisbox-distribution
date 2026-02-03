-- Migration: Add retry tracking and message retry support
-- Created: 2026-02-03

-- Add retry tracking columns to dlr_forwarding_queue table
ALTER TABLE dlr_forwarding_queue
ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMP WITH TIME ZONE;

-- Add retry tracking columns to messages table for message-level retry
ALTER TABLE messages
ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0 NOT NULL,
ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMP WITH TIME ZONE;

-- Add dead_letter_queue table for permanently failed messages
CREATE TABLE IF NOT EXISTS dead_letter_queue (
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

-- Index for quick lookup by message_id
CREATE INDEX IF NOT EXISTS idx_dead_letter_queue_message_id ON dead_letter_queue(message_id);

-- Index for failed_at for cleanup jobs
CREATE INDEX IF NOT EXISTS idx_dead_letter_queue_failed_at ON dead_letter_queue(failed_at);

-- Index for next_retry_at on dlr_forwarding_queue for efficient polling
CREATE INDEX IF NOT EXISTS idx_dlr_forward_queue_next_retry ON dlr_forwarding_queue(next_retry_at) WHERE status = 'pending';

-- Index for next_retry_at on messages for efficient polling
CREATE INDEX IF NOT EXISTS idx_messages_next_retry ON messages(next_retry_at) WHERE processing_status = 'retry_pending';
