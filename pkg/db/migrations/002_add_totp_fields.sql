-- Add TOTP fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMP;

-- Create index for faster TOTP lookups
CREATE INDEX IF NOT EXISTS idx_users_totp_enabled ON users(id, totp_enabled);

-- Add comment for documentation
COMMENT ON COLUMN users.totp_secret IS 'Encrypted TOTP secret using AES-256-GCM';
COMMENT ON COLUMN users.totp_enabled IS 'Whether TOTP 2FA is enabled for this user';
COMMENT ON COLUMN users.totp_verified_at IS 'Timestamp when user successfully verified TOTP setup';
