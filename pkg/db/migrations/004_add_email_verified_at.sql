-- Migration: Add email_verified_at column to users table
-- This tracks when a user's email was verified for auditing purposes

ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP;

-- Create index on is_verified for faster queries when filtering by verification status
CREATE INDEX IF NOT EXISTS idx_users_verified ON users(is_verified);

-- Comment
COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when user verified their email address';
