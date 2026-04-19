-- Migration: Create unverified_users table for pending registrations
-- This table isolates unverified users to prevent DDoS attacks from bloating the main users table
-- Can be moved to a separate database instance in the future for horizontal scaling

CREATE TABLE IF NOT EXISTS unverified_users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_unverified_users_email ON unverified_users(email);
CREATE INDEX IF NOT EXISTS idx_unverified_users_username ON unverified_users(username);
CREATE INDEX IF NOT EXISTS idx_unverified_users_expires_at ON unverified_users(expires_at);

-- Comment
COMMENT ON TABLE unverified_users IS 'Temporary storage for unverified user registrations. Auto-cleanup after 48 hours.';
COMMENT ON COLUMN unverified_users.expires_at IS 'Expiration timestamp for automatic cleanup by background job';
