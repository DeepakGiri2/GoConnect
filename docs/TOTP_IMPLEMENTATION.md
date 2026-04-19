# TOTP Implementation Guide

## Overview

TOTP (Time-based One-Time Password) has been integrated into the GoConnect authentication system to provide two-factor authentication during user registration and login.

## Features

### 1. **Automatic TOTP Setup on Registration**
- When a user registers, a TOTP secret is automatically generated
- The secret is encrypted using AES-256-GCM and stored in the database
- Registration response includes TOTP setup data (secret and QR code URL)
- TOTP is **not** enabled by default - user must verify and enable it

### 2. **TOTP Verification Flow**
Users must complete these steps to enable TOTP:
1. Register and receive TOTP setup data
2. Scan QR code or manually enter secret in authenticator app (Google Authenticator, Authy, etc.)
3. Call `VerifyAndEnableTOTP` with a valid code to activate TOTP

### 3. **TOTP-Protected Login**
- If TOTP is enabled, login requires two steps:
  - Step 1: Username/password validation (returns `requires_totp: true`)
  - Step 2: TOTP code verification (returns access tokens)

## API Methods

### AuthService Methods

#### `Register(username, email, password) (*RegistrationResponse, error)`
Returns:
```go
type RegistrationResponse struct {
    User      *User
    Tokens    *TokenPair
    TOTPSetup *TOTPSetupData {
        Secret    string  // Base32-encoded secret
        QRCodeURL string  // otpauth:// URL for QR code
    }
}
```

#### `GetTOTPSetup(userID) (*TOTPSetupData, error)`
Retrieves TOTP setup information for users who need to reconfigure their authenticator app.

#### `VerifyAndEnableTOTP(userID, code) error`
Verifies the TOTP code and enables TOTP for the user account.

#### `Login(username, password) (*LoginResponse, error)`
Returns:
```go
type LoginResponse struct {
    User         *User
    Tokens       *TokenPair  // nil if TOTP required
    RequiresTOTP bool
}
```

#### `VerifyTOTPLogin(username, code) (*TokenPair, error)`
Verifies TOTP code during login and returns access tokens.

#### `DisableTOTP(userID) error`
Disables TOTP for a user account.

## Database Schema

### Users Table TOTP Fields
```sql
totp_secret      VARCHAR(255)   -- Encrypted TOTP secret
totp_enabled     BOOLEAN        -- Whether TOTP 2FA is enabled
totp_verified_at TIMESTAMP      -- When TOTP was successfully verified
```

## Configuration

Add to `.env`:
```bash
TOTP_ENCRYPTION_KEY=your-32-character-encryption-key-here
```

**Important**: The encryption key must be exactly 32 characters for AES-256.

Generate a secure key:
```bash
# Linux/Mac
openssl rand -base64 32 | cut -c1-32

# Windows PowerShell
-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 32 | % {[char]$_})
```

## Usage Example

### Registration Flow
```go
// 1. User registers
regResp, err := authService.Register("john", "john@example.com", "StrongPass123!")

// 2. Frontend displays QR code from regResp.TOTPSetup.QRCodeURL
// 3. User scans with authenticator app
// 4. User enters code from authenticator
err = authService.VerifyAndEnableTOTP(regResp.User.ID, "123456")
```

### Login Flow (TOTP Enabled)
```go
// 1. User attempts login
loginResp, err := authService.Login("john", "StrongPass123!")

// 2. Check if TOTP is required
if loginResp.RequiresTOTP {
    // 3. Prompt user for TOTP code
    // 4. Verify TOTP and get tokens
    tokens, err := authService.VerifyTOTPLogin("john", "123456")
} else {
    // No TOTP required, tokens already in loginResp.Tokens
}
```

## Security Features

1. **Encryption**: TOTP secrets are encrypted at rest using AES-256-GCM
2. **Time Window**: TOTP codes are validated with ±1 time window (90 seconds total)
3. **6-digit codes**: Standard TOTP configuration
4. **30-second intervals**: Standard TOTP time step

## Testing TOTP

### Manual Testing
1. Install an authenticator app (Google Authenticator, Authy, Microsoft Authenticator)
2. Register a new user account
3. Use the QR code URL or manual secret to add to authenticator
4. Verify with generated code
5. Enable TOTP
6. Test login with TOTP

### Generate QR Code
The QR code URL format:
```
otpauth://totp/GoConnect:user@example.com?secret=BASE32SECRET&issuer=GoConnect
```

You can generate QR codes from this URL using libraries or online tools.

## Migration

The database migration is already in place at:
`pkg/db/migrations/002_add_totp_fields.sql`

Run migrations:
```bash
make migrate-up
```

## Notes

- TOTP is optional by default - users can choose to enable it for extra security
- Once enabled, TOTP verification is required for every login
- Users should save backup codes (future enhancement)
- Consider implementing TOTP bypass codes for account recovery
