# OAuth, JWT, OTP & Database Setup Guide

Complete step-by-step guide to configure OAuth providers, JWT authentication, OTP verification, and database for GoConnect.

---

## Table of Contents
1. [Database Setup](#1-database-setup)
2. [Redis Setup](#2-redis-setup)
3. [JWT Configuration](#3-jwt-configuration)
4. [OTP Configuration](#4-otp-configuration)
5. [OAuth Setup](#5-oauth-setup)
   - [Google OAuth](#google-oauth)
   - [GitHub OAuth](#github-oauth)
   - [Facebook OAuth](#facebook-oauth)
6. [Email Configuration](#6-email-configuration)
7. [Running the Application](#7-running-the-application)
8. [Testing](#8-testing)

---

## 1. Database Setup

### PostgreSQL Installation

**Windows:**
```powershell
# Download from: https://www.postgresql.org/download/windows/
# Or use Chocolatey:
choco install postgresql
```

**Linux/Mac:**
```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# macOS
brew install postgresql
```

### Create Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE goconnect;

# Create user (optional)
CREATE USER goconnect_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect_user;
```

### Run Migrations

```bash
# Apply migrations
psql -U postgres -d goconnect -f pkg/db/migrations/001_initial_schema.sql
```

### Update .env File

```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=goconnect
DATABASE_USER=postgres
DATABASE_PASSWORD=your_secure_password
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=10m
```

---

## 2. Redis Setup

### Redis Installation

**Windows:**
```powershell
# Download from: https://github.com/microsoftarchive/redis/releases
# Or use Chocolatey:
choco install redis-64

# Start Redis
redis-server
```

**Linux:**
```bash
sudo apt-get install redis-server
sudo systemctl start redis
```

**macOS:**
```bash
brew install redis
brew services start redis
```

### Test Redis Connection

```bash
redis-cli ping
# Expected output: PONG
```

### Update .env File

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

---

## 3. JWT Configuration

### Generate Secure JWT Secret

```bash
# Generate a random 64-character secret
openssl rand -base64 64
```

### Update .env File

```env
JWT_SECRET=your_generated_jwt_secret_key_change_this_in_production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
```

**Security Best Practices:**
- Use a strong, random secret (minimum 32 characters)
- Never commit secrets to version control
- Rotate secrets periodically in production
- Use different secrets for different environments

---

## 4. OTP Configuration

### Generate OTP Secret

```bash
# Generate a random secret
openssl rand -base64 32
```

### Update .env File

```env
OTP_SECRET=your_otp_secret_key_change_this_in_production
OTP_EXPIRY=5m
```

**OTP Features:**
- 6-digit codes
- 5-minute expiration
- Rate limiting (max 5 requests per 15 minutes)
- Max 3 verification attempts per OTP
- Stored in Redis for quick access

---

## 5. OAuth Setup

### Google OAuth

#### Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Navigate to "APIs & Services" > "Credentials"

#### Step 2: Create OAuth 2.0 Credentials

1. Click "Create Credentials" > "OAuth 2.0 Client ID"
2. Configure OAuth consent screen (if prompted)
   - User Type: External
   - App name: GoConnect
   - User support email: your-email@example.com
   - Developer contact: your-email@example.com
3. Application type: Web application
4. Name: GoConnect Web Client
5. Authorized redirect URIs:
   - `http://localhost:8080/api/auth/callback/google`
   - Add production URLs when deploying

#### Step 3: Update .env File

```env
GOOGLE_CLIENT_ID=your_google_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback/google
```

---

### GitHub OAuth

#### Step 1: Register OAuth Application

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"

#### Step 2: Configure Application

- Application name: `GoConnect`
- Homepage URL: `http://localhost:8080`
- Authorization callback URL: `http://localhost:8080/api/auth/callback/github`

#### Step 3: Update .env File

```env
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=http://localhost:8080/api/auth/callback/github
```

---

### Facebook OAuth

#### Step 1: Create Facebook App

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Click "My Apps" > "Create App"
3. Select "Consumer" as app type
4. Fill in app details

#### Step 2: Configure Facebook Login

1. In app dashboard, go to "Add Product" > "Facebook Login"
2. Select "Web"
3. Add OAuth Redirect URIs:
   - `http://localhost:8080/api/auth/callback/facebook`

#### Step 3: Get App Credentials

1. Go to Settings > Basic
2. Copy App ID and App Secret

#### Step 4: Update .env File

```env
FACEBOOK_CLIENT_ID=your_facebook_app_id
FACEBOOK_CLIENT_SECRET=your_facebook_app_secret
FACEBOOK_REDIRECT_URL=http://localhost:8080/api/auth/callback/facebook
```

---

## 6. Email Configuration

### Option 1: Gmail (For Development)

#### Step 1: Enable 2-Factor Authentication
1. Go to Google Account settings
2. Enable 2-Factor Authentication

#### Step 2: Create App Password
1. Go to [App Passwords](https://myaccount.google.com/apppasswords)
2. Select "Mail" and "Other (Custom name)"
3. Name it "GoConnect"
4. Copy the generated password

#### Step 3: Update .env File

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_16_char_app_password
FROM_EMAIL=noreply@goconnect.com
FROM_NAME=GoConnect
```

### Option 2: SendGrid (For Production)

```env
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your_sendgrid_api_key
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=GoConnect
```

### Option 3: AWS SES (For Production)

```env
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=your_aws_smtp_username
SMTP_PASSWORD=your_aws_smtp_password
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=GoConnect
```

---

## 7. Running the Application

### Step 1: Copy Environment File

```bash
cp .env.example .env
# Edit .env with your actual values
```

### Step 2: Install Dependencies

```bash
go mod download
go mod tidy
```

### Step 3: Start Services

**Start PostgreSQL and Redis** (if not running):
```bash
# PostgreSQL
sudo systemctl start postgresql  # Linux
# or start via Services on Windows

# Redis
redis-server  # or: sudo systemctl start redis
```

### Step 4: Run Auth Service

```bash
cd cmd/auth
go run main.go
```

### Step 5: Run Gateway Service

```bash
cd cmd/gateway
go run main.go
```

---

## 8. Testing

### Test Database Connection

```bash
psql -U postgres -d goconnect -c "SELECT 1;"
```

### Test Redis Connection

```bash
redis-cli ping
```

### Test API Endpoints

**Register User:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePass123!"
  }'
```

**Request OTP:**
```bash
curl -X POST http://localhost:8080/api/auth/otp/generate \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

**OAuth Login (Google):**
```bash
# Open in browser:
http://localhost:8080/api/auth/oauth/google
```

---

## Environment Variables Summary

Create a `.env` file in the root directory with all these variables:

```env
# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=goconnect
DATABASE_USER=postgres
DATABASE_PASSWORD=your_secure_password
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=10m

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your_jwt_secret_key_change_this_in_production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# OTP
OTP_SECRET=your_otp_secret_key_change_this_in_production
OTP_EXPIRY=5m

# OAuth - Google
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback/google

# OAuth - Facebook
FACEBOOK_CLIENT_ID=your_facebook_client_id
FACEBOOK_CLIENT_SECRET=your_facebook_client_secret
FACEBOOK_REDIRECT_URL=http://localhost:8080/api/auth/callback/facebook

# OAuth - GitHub
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=http://localhost:8080/api/auth/callback/github

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_specific_password
FROM_EMAIL=noreply@goconnect.com
FROM_NAME=GoConnect

# Service Ports
GATEWAY_PORT=8080
AUTH_SERVICE_HOST=localhost
AUTH_SERVICE_PORT=50051

# Environment
ENV=development
LOG_LEVEL=debug
```

---

## Troubleshooting

### Database Connection Issues
- Verify PostgreSQL is running: `sudo systemctl status postgresql`
- Check credentials in `.env` file
- Ensure database exists: `psql -U postgres -l`

### Redis Connection Issues
- Verify Redis is running: `redis-cli ping`
- Check port 6379 is not blocked

### OAuth Errors
- Verify callback URLs match exactly (including http/https)
- Check client ID and secret are correct
- Ensure OAuth app is not in testing mode (for production)

### Email Sending Issues
- Verify SMTP credentials
- For Gmail, ensure App Password is used (not regular password)
- Check firewall isn't blocking port 587

---

## Security Checklist

- [ ] Use strong, unique secrets for JWT and OTP
- [ ] Never commit `.env` file to version control
- [ ] Use HTTPS in production
- [ ] Rotate secrets regularly
- [ ] Enable rate limiting on auth endpoints
- [ ] Use environment-specific OAuth redirect URLs
- [ ] Implement proper CORS configuration
- [ ] Use prepared statements to prevent SQL injection
- [ ] Hash all passwords with bcrypt
- [ ] Implement account lockout after failed attempts
- [ ] Log authentication events for monitoring

---

## Next Steps

1. Review the API documentation in `docs/API.md`
2. Set up monitoring and logging
3. Configure production environment
4. Implement rate limiting middleware
5. Set up CI/CD pipeline
6. Configure backup strategy for database

For more information, see:
- [Architecture Documentation](ARCHITECTURE.md)
- [API Documentation](API.md)
- [Deployment Guide](DEPLOYMENT.md)
