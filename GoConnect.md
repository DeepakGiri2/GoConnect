# GoConnect

> A production-ready, microservices-based authentication backend written in Go.

GoConnect provides a complete authentication platform with email/password login, OTP-based email verification, TOTP two-factor authentication, OAuth 2.0 (Google, Facebook, GitHub), JWT-based session management, and a dedicated API Gateway — all designed for horizontal scalability.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Services](#2-services)
   - [API Gateway](#21-api-gateway)
   - [Auth Service](#22-auth-service)
3. [Service Communication](#3-service-communication)
4. [Request Lifecycle Walkthrough](#4-request-lifecycle-walkthrough)
5. [Database Schemas](#5-database-schemas)
6. [Redis Key Design](#6-redis-key-design)
7. [API Reference](#7-api-reference)
8. [Core Packages (`pkg/`)](#8-core-packages-pkg)
9. [Internal Packages (`internal/`)](#9-internal-packages-internal)
10. [Configuration Reference](#10-configuration-reference)
11. [Security Design](#11-security-design)
12. [Deployment](#12-deployment)
13. [Development Workflow](#13-development-workflow)
14. [Project Structure](#14-project-structure)

---

## 1. Architecture Overview

```
                          ┌──────────────────────┐
         HTTP Clients ───►│    API Gateway        │  :8080
                          │  (Gin HTTP Server)    │
                          │                       │
                          │  • CORS               │
                          │  • JWT Middleware      │
                          │  • Rate Limiting       │
                          │  • Request Routing     │
                          └──────────┬────────────┘
                                     │ gRPC (port 50051)
                          ┌──────────▼────────────┐
                          │    Auth Service        │  :50051
                          │  (gRPC Server)        │
                          │                       │
                          │  • Registration       │
                          │  • Email Verification │
                          │  • TOTP / 2FA         │
                          │  • OAuth 2.0          │
                          │  • Password Reset     │
                          │  • JWT Token Mgmt     │
                          └─────┬──────────┬──────┘
                                │          │
              ┌─────────────────▼──┐  ┌───▼────────────────┐
              │   PostgreSQL       │  │   Redis             │
              │   :5432            │  │   :6379             │
              │                   │  │                     │
              │  • users           │  │  • OTP codes        │
              │  • oauth_accounts  │  │  • Pending users    │
              │  • refresh_tokens  │  │  • Rate limit ctrs  │
              │  • unverified_users│  │  • Resend cooldowns │
              └────────────────────┘  │  • Block state      │
                                      └─────────────────────┘
```

### Key Design Decisions

| Decision | Rationale |
|---|---|
| **gRPC between services** | Strongly-typed, binary-encoded, low-latency internal communication |
| **API Gateway pattern** | Single public entry point; clean separation of HTTP concerns from business logic |
| **Dual storage for pending users** | Redis (fast TTL) + PostgreSQL (persistence/failover) prevents lost registrations on Redis restart |
| **Bloom filter for usernames** | Eliminates most DB reads for username-availability checks — O(1) in-memory lookup |
| **Sliding window rate limiting** | Per-IP and per-email rate limits backed by Redis; prevents registration abuse |
| **AES-256-GCM TOTP encryption** | TOTP secrets are encrypted at rest in the database |

---

## 2. Services

### 2.1 API Gateway

**Entrypoint:** `cmd/gateway/main.go`  
**Port:** `8080` (HTTP)  
**Framework:** [Gin](https://github.com/gin-gonic/gin)

The Gateway is the **only public-facing service**. It is a stateless HTTP reverse proxy that:

1. Applies middleware globally: CORS, structured logging (`zap`), `gin.Recovery()`
2. Extracts the real client IP via `X-Forwarded-For` / `X-Real-IP` headers
3. Enforces **sliding-window rate limits** for registration and username check endpoints
4. Validates JWT access tokens for protected routes via middleware
5. Proxies all auth operations to the **Auth Service** via gRPC

**Startup sequence:**
```
1. Load config from .env
2. Initialize zap logger
3. Dial Auth Service (gRPC, insecure)
4. Connect to Redis (for rate limiting)
5. Create sliding-window rate limiter
6. Register Gin middleware + routes
7. Start HTTP server
8. Block on OS signal (SIGINT/SIGTERM) → graceful shutdown
```

---

### 2.2 Auth Service

**Entrypoint:** `cmd/auth/main.go`  
**Port:** `50051` (gRPC)  
**Protocol:** Protocol Buffers v3

The Auth Service is the **brain of the system** — it owns all business logic and all database/Redis interactions. It is **not publicly reachable** (ClusterIP in Kubernetes).

**Internal sub-services initialized at startup:**

| Sub-service | Type | Purpose |
|---|---|---|
| `UserRepository` | Repository | CRUD on `users` table with configurable retry/backoff |
| `OAuthRepository` | Repository | CRUD on `oauth_accounts` table |
| `TokenRepository` | Repository | CRUD on `refresh_tokens` table |
| `UnverifiedUserRepository` | Repository | CRUD on `unverified_users` table |
| `BloomFilterService` | In-memory | O(1) username availability screening |
| `OTPService` | pkg | Generate, send, verify time-limited OTP codes |
| `PendingRegistrationService` | Service | Dual-write (Redis + DB) for unverified users |
| `CleanupService` | Background goroutine | Periodically delete expired `unverified_users` rows |
| `AuthService` | Service | Core auth logic (register, login, tokens, TOTP, password reset) |
| `OAuthService` | Service | OAuth 2.0 flows (Google, Facebook, GitHub) |

**Startup sequence:**
```
1. Load config from .env
2. Initialize zap logger
3. Connect to PostgreSQL (with pool: max 25 open, 10 idle)
4. Connect to Redis (x2: go-redis client + internal RedisClient wrapper)
5. Initialize EmailService (SMTP)
6. Initialize OTPService
7. Initialize all Repositories
8. Seed BloomFilter from existing usernames in DB
9. Initialize PendingRegistrationService
10. Initialize CleanupService → start background goroutine
11. Initialize AuthService + OAuthService
12. Register gRPC server with AuthService implementation
13. Listen on :50051
14. Block on OS signal → grpc.GracefulStop()
```

---

## 3. Service Communication

### gRPC Contract

Defined in: `api/shared/proto/auth.proto`  
Generated code: `api/shared/proto_gen/`

```protobuf
service AuthService {
  rpc Register(RegisterRequest)                     returns (RegisterResponse);
  rpc VerifyEmail(VerifyEmailRequest)               returns (VerifyEmailResponse);
  rpc ResendVerificationOTP(ResendOTPRequest)        returns (ResendOTPResponse);
  rpc GetBlockStatus(GetBlockStatusRequest)          returns (GetBlockStatusResponse);
  rpc SetupTOTP(SetupTOTPRequest)                   returns (SetupTOTPResponse);
  rpc VerifyTOTP(VerifyTOTPRequest)                 returns (VerifyTOTPResponse);
  rpc Login(LoginRequest)                           returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest)           returns (ValidateTokenResponse);
  rpc RefreshToken(RefreshTokenRequest)             returns (RefreshTokenResponse);
  rpc GenerateOTP(GenerateOTPRequest)               returns (GenerateOTPResponse);
  rpc VerifyOTP(VerifyOTPRequest)                   returns (VerifyOTPResponse);
  rpc ResetPassword(ResetPasswordRequest)           returns (ResetPasswordResponse);
  rpc OAuthLogin(OAuthLoginRequest)                 returns (OAuthLoginResponse);
  rpc CheckUsernameAvailability(CheckUsernameRequest) returns (CheckUsernameResponse);
}
```

### Regenerating the gRPC stubs

```bash
make proto
# or directly:
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/shared/proto/auth.proto
```

---

## 4. Request Lifecycle Walkthrough

### Registration Flow

```
Client POST /api/auth/register
  │
  ▼
[Gateway] Rate-limit check (IP + email sliding-window)
  │  reject → 429 Too Many Requests
  │  pass ↓
  ▼
[Gateway] → gRPC Register(username, email, password)
  │
  ▼
[AuthService.Register]
  ├─ Validate username / email / password format
  ├─ Check users table (email + username uniqueness)
  ├─ Check pending registrations (Redis + DB)
  │    └─ If email already pending → resend OTP, return "check email"
  ├─ bcrypt hash password
  ├─ PendingRegistrationService.CreatePendingUser()
  │    ├─ Write Redis hash  key=pending:user:{email}  TTL=48h
  │    └─ Insert into unverified_users table           expires_at=+48h
  └─ OTPService.GenerateAndSendEmailOTP(email)
       ├─ Check block key in Redis
       ├─ Check resend cooldown key in Redis
       ├─ Generate cryptographically-secure 6-digit OTP
       ├─ Store  otp:{email}  in Redis  TTL=5min
       ├─ Send OTP via SMTP
       └─ Set resend cooldown key in Redis
  │
  ▼
[Gateway] → HTTP 200 { success: true, email: "..." }
```

### Email Verification Flow

```
Client POST /api/auth/verify-email  { email, otp }
  │
  ▼
[AuthService.VerifyEmailAndCreateUser]
  ├─ OTPService.VerifyEmailOTP(email, otp)
  │    ├─ Check block key (Redis)
  │    ├─ Increment attempt counter (Redis)
  │    ├─ Compare stored OTP
  │    └─ On success: delete otp:*, resend:*, cooldown:* keys
  ├─ PendingRegistrationService.GetPendingUser(email)
  │    ├─ Check Redis first (fast)
  │    └─ Fallback: unverified_users table
  ├─ Create verified User row in users table
  │    (is_verified=true, email_verified_at=NOW())
  ├─ BloomFilter.AddUsername(username)
  ├─ PendingRegistrationService.DeletePendingUser(email)
  │    ├─ Delete Redis key
  │    └─ Delete unverified_users row
  └─ generateTokenPair(user)
       ├─ Sign JWT access token  (15min)
       ├─ Sign JWT refresh token (7 days)
       └─ Insert refresh_tokens row
  │
  ▼
HTTP 200 { access_token, refresh_token, user_id, username, email }
```

### Login Flow

```
Client POST /api/auth/login  { username, password }
  │
  ▼
[AuthService.Login]
  ├─ GetUserByUsername (DB)
  ├─ Check is_active / is_verified
  ├─ bcrypt CompareHash
  └─ If user.totp_enabled == true
       └─ Return { requires_totp: true }  ← client must call /verify-totp
      Else
       └─ generateTokenPair → return tokens

Client POST /api/auth/verify-totp  { username, totp_code }
  │
  ▼
[AuthService.VerifyTOTPAndLogin]
  ├─ GetUserByUsername
  ├─ Decrypt TOTP secret (AES-256-GCM)
  ├─ utils.VerifyTOTPCode(secret, code)   (pquerna/otp TOTP RFC 6238)
  └─ generateTokenPair → return tokens
```

### Password Reset Flow

```
POST /api/auth/forgot-password  { email }
  └─ AuthService.GenerateOTP(email)
       ├─ Lookup user by email
       └─ OTPService.GenerateAndSend() → SMTP

POST /api/auth/verify-otp  { email, otp }
  └─ AuthService.VerifyOTP() → returns user entity

POST /api/auth/reset-password  { email, otp, new_password }
  └─ AuthService.ResetPassword()
       ├─ VerifyOTP
       ├─ Validate new password strength
       ├─ bcrypt hash + UpdatePassword (DB)
       └─ RevokeAllUserTokens (DB)
```

### OAuth Flow

```
Client GET /api/auth/oauth/google
  └─ OAuthService.GetAuthURL("google", state) → redirect to Google

Google redirects to → GET /api/auth/callback/google?code=...
  └─ OAuthHandler.OAuthCallback
       └─ gRPC OAuthLogin(provider="google", code=...)
            └─ OAuthService.HandleOAuthCallback
                 ├─ Exchange code for token (Google API)
                 ├─ Fetch userinfo from googleapis.com
                 ├─ Lookup oauth_accounts by (provider, provider_user_id)
                 │    ├─ Not found → CreateUser + CreateOAuthAccount
                 │    └─ Found     → GetUser + UpdateOAuthAccount tokens
                 └─ generateTokenPair → return tokens
```

### Token Refresh Flow

```
POST /api/auth/refresh  { refresh_token }
  └─ AuthService.RefreshTokens(refreshToken)
       ├─ GetRefreshToken (DB) → check is_revoked + expiry
       ├─ GetUserByID
       ├─ RevokeRefreshToken (old one)
       └─ generateTokenPair → new access + refresh tokens
```

---

## 5. Database Schemas

GoConnect uses **PostgreSQL** with four migrations applied in order.

### Migration 001 — Initial Schema

```sql
-- ─────────────────────────────────────────────
-- TABLE: users
-- Primary entity table for verified accounts
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50)  UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),                    -- NULL for OAuth-only users
    created_at    TIMESTAMP    DEFAULT NOW(),
    updated_at    TIMESTAMP    DEFAULT NOW(),
    is_active     BOOLEAN      DEFAULT true,
    is_verified   BOOLEAN      DEFAULT false
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email    ON users(email);

-- ─────────────────────────────────────────────
-- TABLE: oauth_accounts
-- One user can have multiple OAuth providers
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS oauth_accounts (
    id               SERIAL       PRIMARY KEY,
    user_id          UUID         REFERENCES users(id) ON DELETE CASCADE,
    provider         VARCHAR(20)  NOT NULL,          -- 'google' | 'facebook' | 'github'
    provider_user_id VARCHAR(255) NOT NULL,
    access_token     TEXT,
    refresh_token    TEXT,
    expires_at       TIMESTAMP,
    created_at       TIMESTAMP    DEFAULT NOW(),
    updated_at       TIMESTAMP    DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_oauth_user_id  ON oauth_accounts(user_id);
CREATE INDEX idx_oauth_provider ON oauth_accounts(provider, provider_user_id);

-- ─────────────────────────────────────────────
-- TABLE: refresh_tokens
-- Persistent refresh token store (rotated on use)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         SERIAL       PRIMARY KEY,
    user_id    UUID         REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP    NOT NULL,
    created_at TIMESTAMP    DEFAULT NOW(),
    is_revoked BOOLEAN      DEFAULT false
);

CREATE INDEX idx_refresh_tokens_token   ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_oauth_accounts_updated_at
    BEFORE UPDATE ON oauth_accounts FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

### Migration 002 — TOTP Fields

```sql
-- Adds TOTP (Time-based One-Time Password) support to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret      VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled     BOOLEAN   DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMP;

-- totp_secret is stored AES-256-GCM encrypted (see pkg/crypto)
CREATE INDEX IF NOT EXISTS idx_users_totp_enabled ON users(id, totp_enabled);
```

---

### Migration 003 — Unverified Users Table

```sql
-- Temporary staging table for registrations pending email verification.
-- Isolated from users table to prevent DDoS from bloating the primary table.
-- Rows expire after 48 hours and are cleaned by the CleanupService goroutine.
CREATE TABLE IF NOT EXISTS unverified_users (
    id            VARCHAR(36)  PRIMARY KEY,       -- UUID generated at registration
    username      VARCHAR(50)  UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    expires_at    TIMESTAMP    NOT NULL            -- CleanupService deletes WHERE expires_at < NOW()
);

CREATE INDEX IF NOT EXISTS idx_unverified_users_email      ON unverified_users(email);
CREATE INDEX IF NOT EXISTS idx_unverified_users_username   ON unverified_users(username);
CREATE INDEX IF NOT EXISTS idx_unverified_users_expires_at ON unverified_users(expires_at);
```

---

### Migration 004 — Email Verified At

```sql
-- Audit column tracking when a user first verified their email address
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_users_verified ON users(is_verified);
```

---

### Complete Schema Summary

```
users
├── id                UUID        PK
├── username          VARCHAR(50) UNIQUE
├── email             VARCHAR(255) UNIQUE
├── password_hash     VARCHAR(255)          ← bcrypt, NULL for OAuth users
├── created_at        TIMESTAMP
├── updated_at        TIMESTAMP             ← auto-updated via trigger
├── is_active         BOOLEAN
├── is_verified       BOOLEAN
├── email_verified_at TIMESTAMP             ← set when OTP verified
├── totp_secret       VARCHAR(255)          ← AES-256-GCM encrypted
├── totp_enabled      BOOLEAN
└── totp_verified_at  TIMESTAMP

oauth_accounts
├── id               SERIAL       PK
├── user_id          UUID         FK → users.id
├── provider         VARCHAR(20)  ('google'|'facebook'|'github')
├── provider_user_id VARCHAR(255)
├── access_token     TEXT
├── refresh_token    TEXT
├── expires_at       TIMESTAMP
├── created_at       TIMESTAMP
└── updated_at       TIMESTAMP

refresh_tokens
├── id         SERIAL       PK
├── user_id    UUID         FK → users.id
├── token      VARCHAR(255) UNIQUE
├── expires_at TIMESTAMP
├── created_at TIMESTAMP
└── is_revoked BOOLEAN

unverified_users                            ← temporary; auto-cleaned after 48h
├── id            VARCHAR(36) PK
├── username      VARCHAR(50) UNIQUE
├── email         VARCHAR(255) UNIQUE
├── password_hash VARCHAR(255)
├── created_at    TIMESTAMP
└── expires_at    TIMESTAMP
```

---

## 6. Redis Key Design

All Redis keys used by the system:

| Key Pattern | Type | TTL | Purpose |
|---|---|---|---|
| `pending:user:{email}` | Hash | 48h | Dual-store for unverified registration |
| `otp:{email}` | String | 5min | Stores the active OTP code |
| `otp:attempts:{email}` | String | 20min | Failed OTP verification attempt counter |
| `otp:resend:{email}` | String | 20min | Resend attempt counter |
| `otp:cooldown:{email}` | String | configured | Cooldown between OTP resends |
| `otp:block:{email}` | String | configured | Account temporarily blocked after max failures |
| `ratelimit:{ip}:{window}` | String/ZSet | varies | Sliding window rate-limit counters (per-IP) |
| `ratelimit:email:{email}:{window}` | String/ZSet | varies | Sliding window rate-limit counters (per-email) |

---

## 7. API Reference

Base URL: `http://localhost:8080`

### Public Routes (no auth required)

#### `GET /health`
Service health check.

**Response:**
```json
{ "status": "healthy" }
```

---

#### `POST /api/auth/register`
Register a new account. Sends OTP to email.

**Request:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Validation rules:**
- `username`: 3–50 chars, alphanumeric + underscore
- `password`: min 8 chars, requires uppercase, lowercase, digit, special char

**Response (200):**
```json
{
  "success": true,
  "message": "Registration successful! Please check your email for the verification code.",
  "email": "john@example.com"
}
```

**Rate limited by:** IP (configurable limit) + email (configurable limit)

---

#### `POST /api/auth/verify-email`
Verify email with OTP received after registration.

**Request:**
```json
{ "email": "john@example.com", "otp": "123456" }
```

**Response (200):**
```json
{
  "success": true,
  "message": "Email verified successfully! Your account has been created.",
  "user_id": "uuid",
  "username": "john_doe",
  "email": "john@example.com",
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

---

#### `POST /api/auth/resend-otp`
Resend verification OTP (subject to cooldown and max resend limit).

**Request:**
```json
{ "email": "john@example.com" }
```

---

#### `GET /api/auth/block-status?email={email}`
Check if an email is blocked from OTP attempts.

**Response (200):**
```json
{
  "is_blocked": false,
  "remaining_seconds": 0,
  "remaining_attempts": 3
}
```

---

#### `POST /api/auth/login`
Login with username and password.

**Request:**
```json
{ "username": "john_doe", "password": "SecurePass123!" }
```

**Response — normal login (200):**
```json
{
  "success": true,
  "user_id": "uuid",
  "username": "john_doe",
  "email": "john@example.com",
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "requires_totp": false,
  "is_verified": true
}
```

**Response — TOTP required (200):**
```json
{
  "success": true,
  "user_id": "uuid",
  "message": "TOTP verification required",
  "requires_totp": true,
  "is_verified": true
}
```

---

#### `POST /api/auth/setup-totp`
Initialize TOTP for a user. Returns QR code URL for authenticator apps.

**Request:**
```json
{ "user_id": "uuid" }
```

**Response (200):**
```json
{
  "success": true,
  "secret": "BASE32SECRET",
  "qr_code": "otpauth://totp/...",
  "issuer": "GoConnect",
  "account_name": "john@example.com"
}
```

---

#### `POST /api/auth/verify-totp`
Complete login when TOTP is enabled.

**Request:**
```json
{ "username": "john_doe", "totp_code": "123456" }
```

**Response (200):**
```json
{
  "success": true,
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

---

#### `POST /api/auth/refresh`
Rotate token pair using a valid refresh token.

**Request:**
```json
{ "refresh_token": "eyJ..." }
```

**Response (200):**
```json
{
  "success": true,
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

---

#### `POST /api/auth/forgot-password`
Initiate password reset — sends OTP to email.

**Request:**
```json
{ "email": "john@example.com" }
```

---

#### `POST /api/auth/verify-otp`
Verify OTP for password reset.

**Request:**
```json
{ "email": "john@example.com", "otp": "123456" }
```

---

#### `POST /api/auth/reset-password`
Set a new password after OTP verification.

**Request:**
```json
{
  "email": "john@example.com",
  "otp": "123456",
  "new_password": "NewSecurePass123!"
}
```

---

#### `GET /api/auth/check-username?username={username}`
Check if a username is available (Bloom filter + DB).

**Response (200):**
```json
{ "available": true }
```

**Rate limited by:** IP (configurable limit)

---

#### `GET /api/auth/oauth/:provider`
Initiate OAuth login. `:provider` = `google` | `facebook` | `github`.

Redirects the browser to the provider's authorization page.

---

#### `GET /api/auth/callback/:provider`
OAuth callback endpoint. Called by the OAuth provider after user authorization.

Returns the same token structure as `/login`.

---

### Protected Routes (JWT required)

Add header: `Authorization: Bearer {access_token}`

#### `GET /api/me`
Get the currently authenticated user's profile.

**Response (200):**
```json
{
  "user_id": "uuid",
  "username": "john_doe",
  "email": "john@example.com"
}
```

---

## 8. Core Packages (`pkg/`)

### `pkg/config`
Loads configuration from `.env` using [Viper](https://github.com/spf13/viper). Provides strongly-typed config structs:

| Config Struct | Fields |
|---|---|
| `DatabaseConfig` | Host, Port, Name, User, Password, MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime |
| `RedisConfig` | Host, Port, Password |
| `JWTConfig` | Secret, AccessExpiry (15m), RefreshExpiry (168h) |
| `OTPConfig` | Secret, Expiry (5m), Length (6), MaxVerifyAttempts, MaxResendAttempts, ResendCooldown, BlockDuration, EncryptionKey |
| `RateLimitConfig` | RegistrationIP, RegistrationEmail, UsernameCheck, PendingRegTTL, UnverifiedUserCleanup |
| `OAuthConfig` | Google, Facebook, GitHub (ClientID, ClientSecret, RedirectURL) |
| `ServerConfig` | Port, Host, LogLevel, Env |
| `AuthServiceConfig` | Host, Port |
| `EmailConfig` | SMTPHost, SMTPPort, SMTPUsername, SMTPPassword, FromEmail, FromName |
| `RetryConfig` | MaxRetries (3), InitialBackoff (100ms), MaxBackoff (5s) |

---

### `pkg/db`
PostgreSQL connection factory using `lib/pq`. Configures connection pooling with all parameters from `DatabaseConfig`.

**Migrations** are SQL files in `pkg/db/migrations/` and must be applied in order:
```
001_initial_schema.sql
002_add_totp_fields.sql
003_create_unverified_users_table.sql
004_add_email_verified_at.sql
```

---

### `pkg/redis`
Thin wrapper around `go-redis/v8` exposing:
- `Get`, `Set`, `Delete`, `Exists`, `TTL`, `Expire`
- `Increment`, `HSet`, `HGetAll`

Used by `OTPService` and `PendingRegistrationService`.

---

### `pkg/models`
Domain model structs that map directly to DB tables:

- `User` — verified user entity
- `UnverifiedUser` — pending registration entity
- `OAuthAccount` — linked social login
- `RefreshToken` — persisted refresh token record

Plus response DTOs: `RegistrationResponse`, `LoginResponse`, `EmailVerificationResponse`.

---

### `pkg/middleware`
Reusable Gin middleware:

| File | Middleware | Description |
|---|---|---|
| `cors.go` | `CORS()` | Permissive CORS for development; tighten in production |
| `jwt.go` | `JWTAuth(secret)` | Validates Bearer token; sets `user_id`, `username`, `email` in context |
| `logger.go` | `LoggerMiddleware(log)` | Structured zap request logging |
| `ratelimit.go` | `RateLimit(limiter)` | Sliding-window rate limiting middleware |

---

### `pkg/ratelimit`
`sliding_window.go` — Redis-backed sliding window algorithm for rate limiting. Used by Gateway middleware to throttle:
- **Registration by IP** (default: 5 req/window)
- **Registration by email** (default: 3 req/window)
- **Username check by IP** (default: 30 req/window)

---

### `pkg/notification`
| File | Purpose |
|---|---|
| `email.go` | SMTP email sending with HTML templates for OTP and block notifications |
| `otp_service.go` | OTP lifecycle: generate (crypto/rand), send, verify, rate-limit, block |

OTP security model:
- Cryptographically secure generation (`crypto/rand`)
- Stored in Redis with TTL (default 5 min)
- Max 3 verify attempts before block
- Max 5 resends per registration window (20 min)
- Resend cooldown enforced per send
- Block notification sent via email when blocked

---

### `pkg/crypto`
AES-256-GCM encryption/decryption for TOTP secrets stored in the database. Key is set via `TOTP_ENCRYPTION_KEY` env variable (must be exactly 32 bytes).

---

### `pkg/logger`
Wraps `go.uber.org/zap`. Two modes:
- `development` → human-readable console output
- `production` → structured JSON output

---

### `pkg/utils`
General utilities:
- `HashPassword` / `CheckPassword` — bcrypt
- `GenerateAccessToken` / `GenerateRefreshToken` / `ValidateToken` — JWT (golang-jwt/v5)
- `IsValidUsername` / `IsValidEmail` / `IsValidPassword` — input validation regexes
- `GenerateTOTPSecret` / `VerifyTOTPCode` — TOTP via `pquerna/otp`
- `GenerateGUID` — UUID v4 via `google/uuid`

---

### `pkg/retry`
Exponential backoff retry helper used by repositories for transient DB errors.

---

## 9. Internal Packages (`internal/`)

### `internal/auth/repository/`

| File | Repository | Key Operations |
|---|---|---|
| `user_repository.go` | `UserRepository` | CreateUser, GetUserByID, GetUserByUsername, GetUserByEmail, UsernameExists, EmailExists, UpdatePassword, UpdateTOTPSecret, EnableTOTP, DisableTOTP, GetAllUsernames |
| `token_repository.go` | `TokenRepository` | CreateRefreshToken, GetRefreshToken, RevokeRefreshToken, RevokeAllUserTokens |
| `oauth_repository.go` | `OAuthRepository` | GetOAuthAccount, CreateOAuthAccount, UpdateOAuthAccount |
| `unverified_user_repository.go` | `UnverifiedUserRepository` | Create, GetByEmail, DeleteByEmail, EmailExists, UsernameExists, CleanupExpired |
| `totp_repository.go` | `TOTPRepository` | UpdateTOTPSecret, EnableTOTP, DisableTOTP (subset of UserRepository) |

---

### `internal/auth/service/`

| File | Service | Responsibility |
|---|---|---|
| `auth_service.go` | `AuthService` | Register, VerifyEmailAndCreateUser, ResendVerificationOTP, Login, RefreshTokens, GenerateOTP, VerifyOTP, ResetPassword, SetupTOTP, VerifyAndEnableTOTP, VerifyTOTPLogin |
| `oauth_service.go` | `OAuthService` | GetAuthURL, HandleOAuthCallback, fetchUserInfo, createUserFromOAuth |
| `bloom_filter.go` | `BloomFilterService` | IsUsernamePossiblyTaken, AddUsername, CheckUsernameAvailability |
| `pending_registration.go` | `PendingRegistrationService` | CreatePendingUser (dual-write), GetPendingUser (Redis-first), DeletePendingUser, EmailExists, UsernameExists |
| `cleanup_service.go` | `CleanupService` | Background goroutine — periodically calls `UnverifiedUserRepository.CleanupExpired()` |
| `totp_service.go` | `TOTPService` | (helpers for TOTP setup data) |

---

### `internal/auth/grpc/`

`server.go` — `AuthGRPCServer` implements all 14 gRPC RPC methods by delegating to `AuthService`, `OAuthService`, and `BloomFilterService`. Maps Go errors to gRPC-friendly responses.

---

### `internal/gateway/handlers/`

| File | Handler | HTTP Methods |
|---|---|---|
| `auth_handler.go` | `AuthHandler` | Register, VerifyEmail, ResendOTP, GetBlockStatus, Login, RefreshToken, ForgotPassword, VerifyOTP, ResetPassword, CheckUsername, SetupTOTP, VerifyTOTP |
| `oauth_handler.go` | `OAuthHandler` | InitiateOAuth, OAuthCallback |

All handlers decode the HTTP request body, call the corresponding gRPC method on the Auth Service client, and encode the response.

---

### `internal/gateway/middleware/`

| File | Middleware | Purpose |
|---|---|---|
| `ip_extractor.go` | `ExtractIP()` | Reads `X-Forwarded-For`, `X-Real-IP`, or RemoteAddr; stores in Gin context |
| `rate_limit.go` | `RateLimiter` | `RegistrationRateLimit(ipLimit, emailLimit)` and `UsernameCheckRateLimit(limit)` |

---

### `internal/gateway/routes/`

`routes.go` — Assembles the complete route table:

```
GET  /health
     /api/auth/register           [rate-limited: IP + email]
POST /api/auth/verify-email
     /api/auth/resend-otp
GET  /api/auth/block-status
POST /api/auth/setup-totp
     /api/auth/verify-totp
     /api/auth/login
     /api/auth/refresh
     /api/auth/forgot-password
     /api/auth/verify-otp
     /api/auth/reset-password
GET  /api/auth/check-username     [rate-limited: IP]
     /api/auth/oauth/:provider
     /api/auth/callback/:provider
GET  /api/me                      [JWT required]
```

---

## 10. Configuration Reference

📖 **For detailed configuration information, see [docs/CONFIGURATION.md](docs/CONFIGURATION.md)**

Copy `.env.example` to `.env` and fill in all values before running:

```env
# ── Database ──────────────────────────────
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=goconnect
DATABASE_USER=postgres
DATABASE_PASSWORD=your_secure_password

# Connection pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=10m

# ── Redis ─────────────────────────────────
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# ── JWT ───────────────────────────────────
JWT_SECRET=<min 32 chars, random>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# ── OTP ───────────────────────────────────
OTP_SECRET=<random>
OTP_EXPIRY=5m
OTP_LENGTH=6
OTP_MAX_VERIFY_ATTEMPTS=3
OTP_MAX_RESEND_ATTEMPTS=5
OTP_RESEND_COOLDOWN=60s
OTP_BLOCK_DURATION=5m

# ── TOTP Encryption ───────────────────────
# Must be EXACTLY 32 bytes for AES-256
TOTP_ENCRYPTION_KEY=<32-byte-random-key>

# Generate with: scripts/generate-encryption-key.bat

# ── Rate Limiting ─────────────────────────
RATE_LIMIT_REGISTRATION_IP=5
RATE_LIMIT_REGISTRATION_EMAIL=3
RATE_LIMIT_USERNAME_CHECK=30
PENDING_REGISTRATION_TTL=48h
UNVERIFIED_USER_CLEANUP_TTL=1h

# ── OAuth ─────────────────────────────────
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback/google

FACEBOOK_CLIENT_ID=
FACEBOOK_CLIENT_SECRET=
FACEBOOK_REDIRECT_URL=http://localhost:8080/api/auth/callback/facebook

GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_REDIRECT_URL=http://localhost:8080/api/auth/callback/github

# ── Server ────────────────────────────────
SERVER_HOST=localhost
SERVER_PORT=8080          # Gateway listens here
AUTH_SERVICE_HOST=localhost
AUTH_SERVICE_PORT=50051   # Auth Service gRPC port

ENV=development           # or "production"
LOG_LEVEL=debug

# ── Email (SMTP) ──────────────────────────
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your@email.com
SMTP_PASSWORD=your_app_password
FROM_EMAIL=noreply@goconnect.com
FROM_NAME=GoConnect

# ── Retry ─────────────────────────────────
RETRY_MAX_RETRIES=3
RETRY_INITIAL_BACKOFF=100ms
RETRY_MAX_BACKOFF=5s
```

---

## 11. Security Design

### Token Security
- **Access tokens**: JWT, signed HS256, 15-minute expiry. Never stored in DB.
- **Refresh tokens**: JWT, signed HS256, 7-day expiry. Stored in `refresh_tokens` table. Rotated on every use (old token revoked, new token issued).
- On password reset → all refresh tokens for that user are revoked.

### OTP Security
- Generated using `crypto/rand` (cryptographically secure)
- 6 digits, 5-minute TTL stored in Redis
- Brute-force protection: 3 max attempts → block for 5 minutes
- Block notification email sent to user
- Resend cooldown + max resend limit prevents enumeration/flooding

### TOTP (2FA) Security
- RFC 6238 compliant, 30-second window via `pquerna/otp`
- Secret stored **AES-256-GCM encrypted** in PostgreSQL; decrypted only in-process at validation time
- QR code generated with `otpauth://` URL scheme for authenticator apps

### Password Security
- bcrypt with default cost (≥10 rounds)
- Minimum password requirements enforced on registration and reset

### Rate Limiting
- Sliding-window algorithm backed by Redis (no local state = works across multiple Gateway replicas)
- Registration: per-IP + per-email limits
- Username checks: per-IP limit

### Username Enumeration Prevention
- Bloom filter: in-memory pre-screening (1,000,000 bucket, 1% false-positive rate)
- Positive results always confirmed with a DB read
- Seeds from full `users` table on startup; updated on every successful registration

### Unverified User Isolation
- Pre-verified registrations live in a separate `unverified_users` table (not `users`)
- Also cached in Redis with 48h TTL for fast retrieval
- Background `CleanupService` goroutine periodically deletes expired rows

---

## 12. Deployment

### Local Development

```bash
# 1. Install dependencies
make setup

# 2. Start PostgreSQL + Redis (Docker)
make docker-dev

# 3. Apply DB migrations
make init-db
# or manually:
scripts/init-db.bat

# 4. Run services (two terminals)
go run cmd/auth/main.go      # Terminal 1 (Auth Service :50051)
go run cmd/gateway/main.go   # Terminal 2 (Gateway :8080)

# Or combined:
make dev-all
```

### Building Binaries

```bash
make build
# Produces: bin/auth-service.exe, bin/gateway.exe
```

### Docker Compose

```bash
# Development (with mock OAuth servers)
make docker-dev

# Production
make docker-prod

# Stop
make docker-stop
```

### Kubernetes

```bash
# Deploy to K8s cluster
make k8s-deploy

# Applies in order:
# deployments/k8s/namespace.yaml      → namespace: goconnect
# deployments/k8s/secrets.yaml        → DB credentials, JWT secret, OTP secret
# deployments/k8s/configmap.yaml      → JWT/OTP expiry, env, log level
# deployments/k8s/postgres.yaml       → StatefulSet + ClusterIP Service
# deployments/k8s/redis.yaml          → Deployment + ClusterIP Service
# deployments/k8s/auth-service.yaml   → Deployment (3 replicas) + ClusterIP :50051
# deployments/k8s/gateway.yaml        → Deployment + LoadBalancer :8080
# deployments/k8s/ingress.yaml        → Ingress rules

# Teardown
make k8s-delete
```

**K8s resource limits** (Auth Service):
```yaml
requests: { memory: 256Mi, cpu: 250m }
limits:   { memory: 512Mi, cpu: 500m }
```

### Mock OAuth (Development Only)

```bash
make mock-oauth
# Starts mock OAuth servers on ports 9000-9002 (Google, Facebook, GitHub)

make dev-test
# Starts local test UI at http://localhost:3000/index.html
```

> ⚠️ **Never use mock OAuth in production.**

---

## 13. Development Workflow

### Run Tests

```bash
make test
# go test -v ./...
```

### Regenerate Proto Stubs

```bash
make proto
# Requires protoc + protoc-gen-go + protoc-gen-go-grpc
# Install: scripts/install-protoc.bat
```

### Generate TOTP Encryption Key

```bash
scripts/generate-encryption-key.bat
# Outputs a 32-byte base64 key for TOTP_ENCRYPTION_KEY
```

### Clean Build Artifacts

```bash
make clean
```

---

## 14. Project Structure

```
GoConnect/
├── api/
│   └── shared/
│       ├── proto/
│       │   └── auth.proto              ← gRPC service definition
│       └── proto_gen/                  ← generated Go stubs (do not edit)
│
├── cmd/
│   ├── auth/
│   │   └── main.go                     ← Auth Service entrypoint
│   ├── gateway/
│   │   └── main.go                     ← API Gateway entrypoint
│   └── dev-server/
│       └── main.go                     ← Dev-only HTTP server for test UI
│
├── internal/
│   ├── auth/
│   │   ├── grpc/
│   │   │   └── server.go               ← gRPC handler implementations
│   │   ├── repository/
│   │   │   ├── user_repository.go
│   │   │   ├── token_repository.go
│   │   │   ├── oauth_repository.go
│   │   │   ├── unverified_user_repository.go
│   │   │   └── totp_repository.go
│   │   └── service/
│   │       ├── auth_service.go         ← Core authentication logic
│   │       ├── oauth_service.go        ← OAuth 2.0 flows
│   │       ├── bloom_filter.go         ← Username availability bloom filter
│   │       ├── pending_registration.go ← Dual-store pending user logic
│   │       ├── cleanup_service.go      ← Background cleanup goroutine
│   │       └── totp_service.go         ← TOTP setup helpers
│   └── gateway/
│       ├── handlers/
│       │   ├── auth_handler.go         ← HTTP → gRPC auth handlers
│       │   └── oauth_handler.go        ← HTTP → gRPC OAuth handlers
│       ├── middleware/
│       │   ├── ip_extractor.go         ← Real IP extraction
│       │   └── rate_limit.go           ← Gateway rate limit middleware
│       └── routes/
│           └── routes.go               ← All route registrations
│
├── pkg/
│   ├── config/config.go                ← Viper config loader
│   ├── crypto/                         ← AES-256-GCM TOTP encryption
│   ├── db/
│   │   ├── database.go                 ← PostgreSQL connection factory
│   │   └── migrations/                 ← SQL migration files (001–004)
│   ├── logger/                         ← Zap logger wrapper
│   ├── middleware/                     ← Reusable Gin middleware
│   ├── models/                         ← Domain model structs
│   ├── notification/
│   │   ├── email.go                    ← SMTP email sender
│   │   └── otp_service.go              ← OTP lifecycle management
│   ├── ratelimit/
│   │   └── sliding_window.go           ← Redis sliding-window algorithm
│   ├── redis/redis.go                  ← Redis client wrapper
│   ├── retry/                          ← Exponential backoff retry
│   └── utils/                          ← JWT, bcrypt, validation, UUID, TOTP
│
├── deployments/
│   └── k8s/                            ← Kubernetes YAML manifests
│
├── scripts/                            ← Setup, init-db, proto gen scripts
├── build/docker/                       ← Docker Compose files (dev + prod)
├── docs/                               ← Additional documentation
├── temp/                               ← Dev test UI (HTML/JS)
├── .env                                ← Local configuration (not committed)
├── go.mod                              ← Go module: github.com/goconnect
├── Makefile                            ← All build/run targets
└── README.md                           ← This file
```

---

## Key Dependencies

| Package | Version | Purpose |
|---|---|---|
| `gin-gonic/gin` | v1.9.1 | HTTP framework (Gateway) |
| `google.golang.org/grpc` | v1.67.1 | gRPC framework |
| `google.golang.org/protobuf` | v1.36.1 | Protocol Buffers runtime |
| `go-redis/redis/v8` | v8.11.5 | Redis client |
| `lib/pq` | v1.10.9 | PostgreSQL driver |
| `golang-jwt/jwt/v5` | v5.2.0 | JWT signing/verification |
| `golang.org/x/crypto` | v0.26.0 | bcrypt password hashing |
| `golang.org/x/oauth2` | v0.22.0 | OAuth 2.0 client (Google/Facebook/GitHub) |
| `bits-and-blooms/bloom/v3` | v3.6.0 | In-memory Bloom filter |
| `pquerna/otp` | v1.5.0 | TOTP (RFC 6238) implementation |
| `spf13/viper` | v1.18.2 | Configuration management |
| `go.uber.org/zap` | v1.26.0 | Structured logging |
| `google/uuid` | v1.6.0 | UUID v4 generation |
