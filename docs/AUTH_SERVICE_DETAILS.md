# Auth Service - Deep Dive Architecture and Feature Breakdown

The **Auth Service** (`cmd/auth`) is a robust, security-focused gRPC server responsible for handling user identities, multi-factor authentication, distributed session tracking, and advanced defensive blocking mechanisms. 

It does not expose an HTTP/REST interface; instead, it relies exclusively on a strongly-typed `pb.AuthServiceClient` gRPC protocol.

## 1. Comprehensive Feature Breakdown

### 1.1 Dual-Storage Pending Registration
Instead of writing partially verified users directly into the primary `users` repository, the system protects core datasets using a "Dual-Storage" strategy (`internal/auth/service/pending_registration.go`):
- **Redis Layer**: Pending registrations are aggressively cached in Redis hashes (`pending:user:{email}`). It benefits from native Redis TTLs, which will automatically eject dormant registrations after 48 hours without expensive database scans.
- **PostgreSQL Fallback**: Mirrors the pending data into an `unverified_users` table to guarantee users do not lose their sign-up state if the Redis cache is flushed or restarts.
- **Async Cleanup Worker**: The `CleanupService` continuously polls the database for expired `unverified_users` rows on an async goroutine to sweep orphaned database files.

### 1.2 Defensive OTP & Threat Blocking
The `pkg/notification/otp_service.go` is heavily armed against enumeration, brute-forcing, and spam:
- **Numerical Generative Logic**: Utilizes `crypto/rand` bound to `math/big` exponents to generate mathematically secure OTP nonces.
- **Hard Rate-Limiting**: 
  - Restricts OTP resend requests (tracked via `otp:resend:{email}`).
  - Enforces mandatory cooldowns (`otp:cooldown:{email}`) preventing malicious actors from spamming SMTP pipelines.
- **Dynamic Lockouts**: If a user exceeds incorrect verification guesses, the system injects an `otp:block:{email}` key into Redis. During this timeout duration, API calls fail instantly at the cache boundary.

### 1.3 Probabilistic Bloom Filter
Handling `/check-username` requests can be disastrous for a database during a DDoS event. The Auth Service deploys a `BloomFilterService` (`github.com/bits-and-blooms/bloom`):
- Initializes on boot by scraping all usernames from Postgres.
- Checks against the Bloom Filter taking microseconds. If the filter declares a username is available, the Database is *never* queried. 
- Only if the filter flags "possibly taken" does it fall back to a Postgres validation.

### 1.4 Cryptographic Multi-Factor Authentication (TOTP)
Handled via `internal/auth/service/totp_service.go`:
- Calculates time-based one-time passwords via `pquerna/otp`.
- **Pre-Validation Safety**: New TOTP secret hashes are written to a temporary Redis pipeline (`totp:pending:{userID}`). Before committing to Postgres, the user *must* successfully synthesize an OTP against it.
- **Replay Protection**: Successfully utilized OTPs are pushed into an exclusion hash map (`totp:used:{userID}:{code}`) in Redis for 90 seconds, making replay attacks geometrically impossible.
- **Secret Encryption**: Stored keys within the database are AES-encrypted at rest by `crypto.Encryptor`.

### 1.5 Secure JWT & Token Lifecycle
- Yields a dual `TokenPair` (AccessToken / RefreshToken).
- Expired Refresh Tokens check the database `token_repository.go` and ensure the token flag `IsRevoked` guarantees manual session revocation is respected system-wide.

---

## 2. Infrastructure Dependencies and Justifications

### A. PostgreSQL Server
The canonical source of truth for the Auth Service.
- **What it does**: Stores `UserRepository`, `OAuthRepository`, `TokenRepository`, and `UnverifiedUser` logs.
- **Why it's used**: ACID compliance. Once an account is validated, it must be durably stored. Relational schemas map perfectly to `users` -> `tokens` 1:N relations.

### B. Redis Cluster
Essentially acts as the architectural glue for real-time defensive evaluations.
- **What it does**: Bloom Filter acceleration, TOTP temporary states, OTP spam throttling, Account Brute-force Lockouts, Token Blacklisting, and Dual-Storage pending caching.
- **Why it's used**: Redis's atomic operations (`INCR`, `HSETNX`) and native `EXPIRE` attributes allow GoConnect to automatically expire security policies (like a 15-minute lock on an account) without writing background-polling loop logic. Constant `O(1)` memory lookup times make it formidable under heavy load.

### C. SMTP Exchange (Email)
- **What it does**: Dispatches numeric verification payloads and lockout warnings.
- **Why it's used**: External user validation. Wrapped cleanly through `notification.EmailConfig` ensuring synchronous OTP issuance.
