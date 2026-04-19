# GoConnect API Architecture and Flow Visualization

This document provides a detailed breakdown of the `GoConnect` backend architecture, explaining how different features and services interact with each other.

The application follows a **microservices-oriented architecture** using an **API Gateway pattern**. The primary communication between the Gateway and internal microservices (like the Auth service) is handled over **gRPC**.

## 1. High-Level Architecture

The core of the system is divided into moving parts:
- **Client App**: Makes HTTP/REST requests.
- **API Gateway (Gin)**: Acts as the single entry point. Handles HTTP routing, rate-limiting, CORS, and JWT authentication for protected routes.
- **Auth Service (gRPC)**: Handles business logic for user management, OTP, TOTP, and sessions.
- **Redis**: Used for Session tokens, Rate Limiting, Bloom Filters (username availability check), and OTP storage.
- **PostgreSQL Database**: Persistent storage for Users, Tokens, and Unverified Registrations.
- **Email Provider**: External SMTP server to send OTPs and verification emails.

```mermaid
graph TD
    Client[Client Devices] -->|HTTP REST| Gateway(API Gateway\nGin Framework)
    Gateway -->|Rate Limiting| Redis[(Redis)]
    Gateway -->|gRPC| AuthService(Auth Service\ngRPC Server)
    
    AuthService -->|Read/Write| Postgres[(PostgreSQL DB)]
    AuthService -->|Store OTP & Bloom Filter| Redis
    AuthService -->|Send Notification| EmailService[Email Service/SMTP]
```

---

## 2. Feature Interaction Flows

Below are detailed, step-by-step sequence diagrams of the major features in the platform.

### 2.1 User Registration and Email Verification Flow

When a user signs up, they are not immediately created as a verified user. They are temporarily buffered as a pending registration.

```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant A as Auth Service
    participant R as Redis
    participant DB as PostgreSQL
    participant E as Email Service

    %% Step 1: Registration
    C->>G: POST /api/auth/register (IP & Email rate limited)
    G->>R: Check Rate Limits
    R-->>G: Allowed
    G->>A: gRPC Register(Username, Email, Password)
    A->>DB: Save to Unverified Users / Pending Registration
    A->>R: Generate & Store OTP
    A->>E: Send Verification Email with OTP
    A-->>G: RegisterResponse (Success)
    G-->>C: 200 OK (OTP Sent to Email)

    %% Step 2: Verification
    C->>G: POST /api/auth/verify-email
    G->>A: gRPC VerifyEmail(Email, OTP)
    A->>R: Validate OTP
    R-->>A: OTP is Valid
    A->>DB: Move Unverified User -> Actual Users Table
    A->>DB: Remove from Unverified
    A->>A: Generate Access & Refresh Tokens
    A-->>G: VerifyEmailResponse (Tokens included)
    G-->>C: 200 OK (User Authorized & Tokens sent)
```

### 2.2 User Login Flow with Conditional TOTP (2FA)

If a user has 2FA enabled, the standard login acts as a pre-authorization step before full tokens are dispensed.

```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant A as Auth Service
    participant DB as PostgreSQL

    C->>G: POST /api/auth/login
    G->>A: gRPC Login(Username, Password)
    A->>DB: Fetch User & Hash Check
    DB-->>A: Valid Credentials
    
    alt User has TOTP enabled
        A-->>G: LoginResponse (requires_totp: true)
        G-->>C: 200 OK (Requires TOTP)
        
        C->>G: POST /api/auth/verify-totp
        G->>A: gRPC VerifyTOTP(Username, TOTP Code)
        A->>A: Validate TOTP crypto code
        A->>A: Generate Access & Refresh Tokens
        A-->>G: VerifyTOTPResponse (Tokens)
        G-->>C: 200 OK (Tokens sent)
    else User does NOT have TOTP enabled
        A->>A: Generate Access & Refresh Tokens
        A-->>G: LoginResponse (Tokens)
        G-->>C: 200 OK (Tokens sent)
    end
```

### 2.3 Password Reset Flow

The password reset relies on the OTP service validating identity via Email. Redis manages the expiry and lockout mechanism to prevent abuse.

```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant A as Auth Service
    participant R as Redis
    participant E as Email Service

    C->>G: POST /api/auth/forgot-password
    G->>A: gRPC GenerateOTP(Email)
    A->>R: Store Reset OTP with TTL
    A->>E: Send Password Reset OTP
    A-->>G: GenerateOTPResponse
    G-->>C: 200 OK
    
    C->>G: POST /api/auth/verify-otp
    G->>A: gRPC VerifyOTP(Email, OTP)
    A->>R: Validate OTP against Redis
    R-->>A: OTP Match
    A-->>G: VerifyOTPResponse (User ID)
    G-->>C: 200 OK (Token or Allow Next Step)
    
    C->>G: POST /api/auth/reset-password
    G->>A: gRPC ResetPassword(Email, OTP, NewPassword)
    A->>R: Validate OTP again (or verify short-lived permission)
    A->>DB: Update Password Hash
    A-->>G: ResetPasswordResponse
    G-->>C: 200 OK (Password changed)
```

### 2.4 Protected Routes and JWT Validation

Once authenticated, the Client attaches a JWT to their requests. The API Gateway verifies this token without needing to constantly ping the Auth Service unless deep verification or invalidation is required.

```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant A as Auth Service (Optional check)

    C->>G: GET /api/me (Header: Authorization Bearer [Token])
    G->>G: JWT Middleware intercepts request
    G->>G: Validates JWT signature locally using Shared Secret
    
    alt Token is Valid
        G->>G: Extract user_id, username, email locally
        G->>Gateway Handler: Process Route
        Gateway Handler-->>C: 200 OK {user_id, username, email}
    else Token is Invalid / Expired
        G-->>C: 401 Unauthorized
        
        %% Token Refresh Flow
        C->>G: POST /api/auth/refresh (Requires Refresh Token)
        G->>A: gRPC RefreshToken(RefreshToken)
        A->>DB: Validate Refresh Token in DB
        A->>A: Generate New Access & Refresh Tokens
        A-->>G: RefreshTokenResponse
        G-->>C: 200 OK (New Tokens)
    end
```

### 2.5 Real-Time Username Check (Bloom Filter Optimization)

To prevent hammering the database on username availability lookups, the system utilizes a **Redis Bloom Filter**.

```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant R as Redis (Rate Limiter)
    participant A as Auth Service
    participant BF as Redis (Bloom Filter)

    C->>G: GET /api/auth/check-username
    G->>R: Rate Limit Check (Prevents scraping)
    G->>A: gRPC CheckUsernameAvailability(Username)
    A->>BF: Probably Exists? (O(1) Check)
    
    alt Bloom Filter returns YES
        A->>DB: Verifies against actual DB table (rare cache collision check)
        DB-->>A: True/False
    else Bloom Filter returns NO
        A-->>G: Definitely Available
    end
    G-->>C: 200 OK {available: boolean}
```

---

## 3. Key Components in Detail

1. **API Gateway (`cmd/gateway`)**:
   - Built on `gin-gonic/gin`.
   - Protects the system using Redis sliding-window Rate Limiting middleware.
   - Converts HTTP/REST into strict Protocol Buffer requests for the inner backend logic.

2. **Auth Service (`cmd/auth`)**:
   - A standalone process completely detached from HTTP rules. Only exposes a gRPC server defined in `auth.proto`.
   - Connects to the Database layer (`pkg/db`).
   - Runs a background task `CleanupService` to constantly clear out unverified users and dead OTPs from the database to keep the system clean.
   
3. **API Shared Contracts (`api/shared`)**:
   - Single source of truth. Contains the `.proto` files allowing Gateway to understand what Auth Service expects.
   
4. **Resilience & Rate Limiting Strategy**:
   - **Signup Spam Protection**: Specifically hard limits IPs and Emails attempting to register iteratively.
   - **OTP Brute Force Protection**: Stores attempts in Redis. Generates cooldown lockouts via custom `otpService` when limits are reached.
   - **Scrape Protection**: Bloom Filters stop users from brute-forcing millions of usernames. Rate limits specifically throttle the username endpoint.
