# Configuration Guide

This document provides detailed information about configuring the GoConnect application.

## Table of Contents

- [Environment Variables](#environment-variables)
- [Database Configuration](#database-configuration)
- [Email/SMTP Configuration](#emailsmtp-configuration)
- [OAuth Configuration](#oauth-configuration)
- [Security Configuration](#security-configuration)
- [Production Considerations](#production-considerations)

## Environment Variables

All configuration is managed through environment variables, typically stored in a `.env` file. Use `.env.example` as a template.

### Database Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_HOST` | PostgreSQL host address | `localhost` | Yes |
| `DATABASE_PORT` | PostgreSQL port | `5432` | Yes |
| `DATABASE_NAME` | Database name | `goconnect` | Yes |
| `DATABASE_USER` | Database user | `postgres` | Yes |
| `DATABASE_PASSWORD` | Database password | - | Yes |
| `DB_MAX_OPEN_CONNS` | Max open connections | `25` | No |
| `DB_MAX_IDLE_CONNS` | Max idle connections | `10` | No |
| `DB_CONN_MAX_LIFETIME` | Connection max lifetime | `5m` | No |
| `DB_CONN_MAX_IDLE_TIME` | Idle connection timeout | `10m` | No |

### Redis Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `REDIS_HOST` | Redis host address | `localhost` | Yes |
| `REDIS_PORT` | Redis port | `6379` | Yes |
| `REDIS_PASSWORD` | Redis password (if auth enabled) | - | No |

### JWT Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET` | Secret key for JWT signing | - | Yes |
| `JWT_ACCESS_EXPIRY` | Access token expiry duration | `15m` | Yes |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry duration | `168h` (7 days) | Yes |

**Important:** Generate a strong random secret for production:
```bash
openssl rand -base64 32
```

### OTP Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `OTP_SECRET` | Secret key for OTP generation | - | Yes |
| `OTP_EXPIRY` | OTP expiration time | `5m` | Yes |
| `OTP_LENGTH` | Length of OTP code | `6` | No |
| `OTP_MAX_VERIFY_ATTEMPTS` | Max verification attempts | `5` | No |
| `OTP_MAX_RESEND_ATTEMPTS` | Max resend attempts | `3` | No |
| `OTP_RESEND_COOLDOWN` | Cooldown between resends | `1m` | No |
| `OTP_BLOCK_DURATION` | Block duration after max attempts | `15m` | No |
| `TOTP_ENCRYPTION_KEY` | 32-byte key for TOTP encryption | - | Yes |

**Generate encryption key:**
```bash
# Windows
.\scripts\generate-encryption-key.bat

# Linux/Mac
./scripts/generate-encryption-key.sh
```

### Email/SMTP Configuration

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `SMTP_HOST` | SMTP server hostname | `smtp.gmail.com` | Yes |
| `SMTP_PORT` | SMTP server port | `587` | Yes |
| `SMTP_USERNAME` | SMTP authentication username | `your_email@gmail.com` | Yes |
| `SMTP_PASSWORD` | SMTP authentication password | App password | Yes |
| `FROM_EMAIL` | Sender email address | `noreply@goconnect.com` | Yes |
| `FROM_NAME` | Sender display name | `GoConnect` | Yes |

#### Popular SMTP Providers

**Gmail:**
- Host: `smtp.gmail.com`
- Port: `587`
- Note: Use [App Password](https://support.google.com/accounts/answer/185833) instead of regular password
- Enable 2FA and generate an app-specific password

**SendGrid:**
- Host: `smtp.sendgrid.net`
- Port: `587`
- Username: `apikey`
- Password: Your SendGrid API key

**Mailgun:**
- Host: `smtp.mailgun.org`
- Port: `587`
- Credentials available in Mailgun dashboard

**AWS SES:**
- Host: `email-smtp.{region}.amazonaws.com`
- Port: `587`
- Use SMTP credentials from AWS SES console

### OAuth Configuration

#### Google OAuth

| Variable | Description | Required |
|----------|-------------|----------|
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | Yes |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | Yes |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL | Yes |

Setup: [Google Cloud Console](https://console.cloud.google.com/apis/credentials)

#### Facebook OAuth

| Variable | Description | Required |
|----------|-------------|----------|
| `FACEBOOK_CLIENT_ID` | Facebook app ID | Yes |
| `FACEBOOK_CLIENT_SECRET` | Facebook app secret | Yes |
| `FACEBOOK_REDIRECT_URL` | OAuth callback URL | Yes |

Setup: [Facebook Developers](https://developers.facebook.com/apps/)

#### GitHub OAuth

| Variable | Description | Required |
|----------|-------------|----------|
| `GITHUB_CLIENT_ID` | GitHub OAuth app client ID | Yes |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth app client secret | Yes |
| `GITHUB_REDIRECT_URL` | OAuth callback URL | Yes |

Setup: [GitHub Developer Settings](https://github.com/settings/developers)

### Rate Limiting Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `RATE_LIMIT_REGISTRATION_IP` | Max registrations per IP | `5` | No |
| `RATE_LIMIT_REGISTRATION_EMAIL` | Max attempts per email | `3` | No |
| `RATE_LIMIT_USERNAME_CHECK` | Username check rate limit | `10` | No |
| `PENDING_REGISTRATION_TTL` | Pending registration expiry | `15m` | No |
| `UNVERIFIED_USER_CLEANUP_TTL` | Cleanup unverified users after | `24h` | No |

### Server Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` | No |
| `SERVER_PORT` | Gateway server port | `8080` | No |
| `GATEWAY_PORT` | Gateway port (legacy) | `8080` | No |
| `AUTH_SERVICE_HOST` | Auth gRPC service host | `localhost` | Yes |
| `AUTH_SERVICE_PORT` | Auth gRPC service port | `50051` | Yes |
| `ENV` | Environment mode | `development` | Yes |
| `LOG_LEVEL` | Logging level | `debug` | Yes |

**Environment modes:**
- `development` - Verbose logging, no SSL requirements
- `staging` - Production-like with debug logging
- `production` - Minimal logging, strict security

**Log levels:**
- `debug` - All logs including debug info
- `info` - General informational messages
- `warn` - Warning messages
- `error` - Error messages only

### Retry Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `RETRY_MAX_RETRIES` | Maximum retry attempts | `3` | No |
| `RETRY_INITIAL_BACKOFF` | Initial backoff duration | `100ms` | No |
| `RETRY_MAX_BACKOFF` | Maximum backoff duration | `5s` | No |

## Production Considerations

### Security Checklist

- [ ] Generate strong random secrets for all `*_SECRET` variables
- [ ] Use secure SMTP credentials (app passwords, API keys)
- [ ] Enable TLS/SSL for database connections
- [ ] Set `ENV=production`
- [ ] Set `LOG_LEVEL=info` or `warn`
- [ ] Use strong database passwords
- [ ] Enable Redis authentication
- [ ] Configure proper OAuth redirect URLs (HTTPS)
- [ ] Set appropriate rate limits
- [ ] Review and adjust connection pool settings

### Environment-Specific Settings

**Development:**
```bash
ENV=development
LOG_LEVEL=debug
SERVER_HOST=localhost
```

**Production:**
```bash
ENV=production
LOG_LEVEL=info
SERVER_HOST=0.0.0.0
# Use environment secrets management (AWS Secrets Manager, HashiCorp Vault, etc.)
```

### Docker Considerations

When running with Docker Compose:
- Database host should be service name: `DATABASE_HOST=postgres`
- Redis host should be service name: `REDIS_HOST=redis`
- Auth service host: `AUTH_SERVICE_HOST=auth-service`

### Secrets Management

**Development:** Use `.env` file (never commit to git)

**Production:** Use secure secrets management:
- Kubernetes Secrets
- AWS Secrets Manager
- HashiCorp Vault
- Azure Key Vault
- Google Secret Manager

### Monitoring & Logging

Configure external logging services:
- Datadog
- New Relic
- Sentry
- ELK Stack
- CloudWatch

## Quick Start

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Generate encryption key:
   ```bash
   # Windows
   .\scripts\generate-encryption-key.bat
   
   # Linux/Mac
   ./scripts/generate-encryption-key.sh
   ```

3. Update required fields:
   - Database credentials
   - JWT secret
   - SMTP credentials
   - OAuth credentials (optional)

4. Start the services:
   ```bash
   docker-compose -f build/docker/docker-compose.dev.yml up
   ```

## Troubleshooting

### Common Issues

**SMTP Connection Failed:**
- Verify SMTP credentials
- Check firewall/security group rules
- For Gmail: Enable "Less secure app access" or use App Password
- Verify port is correct (587 for TLS, 465 for SSL)

**Database Connection Failed:**
- Verify database is running
- Check host and port settings
- Verify credentials
- Check network connectivity

**Redis Connection Failed:**
- Verify Redis is running
- Check if authentication is required
- Verify host and port

**OAuth Redirect Mismatch:**
- Ensure redirect URLs match exactly in provider settings
- Use HTTPS in production
- Check for trailing slashes

## Additional Resources

- [PostgreSQL Configuration](https://www.postgresql.org/docs/current/runtime-config.html)
- [Redis Configuration](https://redis.io/docs/management/config/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OAuth 2.0 Specification](https://oauth.net/2/)
