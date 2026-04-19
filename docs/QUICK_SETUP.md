# Quick Setup Guide

Get GoConnect up and running in minutes!

## Prerequisites

- Docker & Docker Compose
- Git
- Windows (PowerShell) or Linux/Mac (Bash)

## Setup Steps

### 1. Clone and Navigate

```bash
git clone <repository-url>
cd GoConnect
```

### 2. Configure Environment

```bash
# Copy the example configuration
cp .env.example .env
```

### 3. Generate Encryption Key

**Windows:**
```powershell
.\scripts\generate-encryption-key.bat
```

**Linux/Mac:**
```bash
chmod +x scripts/generate-encryption-key.sh
./scripts/generate-encryption-key.sh
```

This will generate a secure 32-byte encryption key and update your `.env` file.

### 4. Configure SMTP (Required for OTP)

Open `.env` and update the SMTP section:

**For Gmail:**
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
FROM_EMAIL=noreply@goconnect.com
FROM_NAME=GoConnect
```

> **Important:** For Gmail, use an [App Password](https://support.google.com/accounts/answer/185833), not your regular password.

**For Other Providers:**
- See [CONFIGURATION.md](CONFIGURATION.md#emailsmtp-configuration) for SendGrid, Mailgun, AWS SES, etc.

### 5. Update Secrets

Generate strong random secrets:

```bash
# For JWT_SECRET
openssl rand -base64 32

# For OTP_SECRET
openssl rand -base64 32
```

Update these in `.env`:
```env
JWT_SECRET=<generated_jwt_secret>
OTP_SECRET=<generated_otp_secret>
```

### 6. (Optional) Configure OAuth

If you want to use OAuth providers:

**Google:**
1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create OAuth 2.0 credentials
3. Add redirect URL: `http://localhost:8080/api/auth/callback/google`
4. Update `.env` with Client ID and Secret

**Facebook:**
1. Go to [Facebook Developers](https://developers.facebook.com/apps/)
2. Create a new app
3. Add redirect URL: `http://localhost:8080/api/auth/callback/facebook`
4. Update `.env` with App ID and Secret

**GitHub:**
1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Create OAuth App
3. Add callback URL: `http://localhost:8080/api/auth/callback/github`
4. Update `.env` with Client ID and Secret

### 7. Validate Configuration

**Windows:**
```powershell
.\scripts\validate-env.bat
```

**Linux/Mac:**
```bash
chmod +x scripts/validate-env.sh
./scripts/validate-env.sh
```

This script checks if all required variables are set.

### 8. Initialize Database

**Windows:**
```powershell
.\scripts\init-db.bat
```

**Linux/Mac:**
```bash
chmod +x scripts/init-db.sh
./scripts/init-db.sh
```

### 9. Start Services

```bash
docker-compose -f build/docker/docker-compose.dev.yml up -d
```

### 10. Verify Services

Check if all services are running:

```bash
docker-compose -f build/docker/docker-compose.dev.yml ps
```

You should see:
- `goconnect-postgres` (healthy)
- `goconnect-redis` (healthy)
- `goconnect-auth` (running)
- `goconnect-gateway` (running)
- `goconnect-dev-frontend` (running)

### 11. Test the API

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

## Next Steps

- **API Documentation:** See [API.md](API.md)
- **Architecture:** See [ARCHITECTURE.md](ARCHITECTURE.md)
- **Deployment:** See [DEPLOYMENT.md](DEPLOYMENT.md)
- **Full Configuration:** See [CONFIGURATION.md](CONFIGURATION.md)

## Test the Registration Flow

Use the dev frontend at `http://localhost:3000` or test with curl:

```bash
# Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

You should receive an OTP via email!

## Troubleshooting

### SMTP Not Working

1. **Gmail:** Enable 2FA and create an [App Password](https://support.google.com/accounts/answer/185833)
2. **Check credentials:** Verify SMTP_USERNAME and SMTP_PASSWORD
3. **Check port:** Use 587 for TLS, 465 for SSL
4. **Firewall:** Ensure outbound SMTP traffic is allowed

### Database Connection Failed

1. Check if PostgreSQL is running: `docker-compose -f build/docker/docker-compose.dev.yml ps postgres`
2. Verify credentials in `.env`
3. Check logs: `docker logs goconnect-postgres`

### Redis Connection Failed

1. Check if Redis is running: `docker-compose -f build/docker/docker-compose.dev.yml ps redis`
2. Check logs: `docker logs goconnect-redis`

### Auth Service Not Starting

1. Check logs: `docker logs goconnect-auth`
2. Verify TOTP_ENCRYPTION_KEY is exactly 32 bytes
3. Ensure database is healthy and initialized

## Common Configuration Mistakes

| Issue | Solution |
|-------|----------|
| OTP emails not sending | Configure SMTP settings correctly |
| TOTP encryption error | Generate key with `generate-encryption-key` script |
| JWT validation fails | Ensure JWT_SECRET is the same across all services |
| OAuth redirect mismatch | Callback URLs must match exactly in provider settings |
| Database connection refused | Use service names in Docker: `postgres`, `redis` |

## Development vs Production

**Development (.env):**
```env
ENV=development
LOG_LEVEL=debug
SERVER_HOST=localhost
```

**Production:**
```env
ENV=production
LOG_LEVEL=info
SERVER_HOST=0.0.0.0
# Use secrets management (AWS Secrets Manager, Vault, etc.)
```

## Stop Services

```bash
docker-compose -f build/docker/docker-compose.dev.yml down
```

To remove volumes (⚠️ deletes all data):
```bash
docker-compose -f build/docker/docker-compose.dev.yml down -v
```

## Need Help?

- 📖 [Full Configuration Guide](CONFIGURATION.md)
- 🏗️ [Architecture Documentation](ARCHITECTURE.md)
- 🚀 [Deployment Guide](DEPLOYMENT.md)
- 📡 [API Reference](API.md)
