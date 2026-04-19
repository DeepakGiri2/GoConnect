# Authentication and Gateway Server Communication Architecture

This document details how the standard Gateway Server and the Authentication (Auth) Server communicate, alongside comprehensive visual diagrams illustrating the flow of various critical APIs such as Login, Registration, and Rate Limiting.

## 1. System Communication Overview

The system employs a microservices architectural pattern where:
- **Gateway Server**: Acts as the public-facing entry point. It's an HTTP server written using the Gin framework. It handles routing, initial request validation, IP extraction, and rate limiting.
- **Auth Server**: An internal service handling the core business logic of user authentication, token generation, and account management. It exposes its services via **gRPC**.

### Gateway to Auth Communication

Communication between the Gateway and the Auth server is strictly **RPC-based** using gRPC.
1. The Gateway maintains a gRPC client connection to the Auth server.
2. Incoming HTTP JSON requests are parsed by the Gateway into internal models.
3. The Gateway invokes the appropriate gRPC method (e.g., `authClient.Login()`).
4. The Auth server processes the request (with database/redis dependencies).
5. The Auth server responds with a gRPC message.
6. The Gateway translates this gRPC response back to HTTP/JSON formats suitable for the client.

```mermaid
sequenceDiagram
    participant Client as Web/Mobile Client
    participant Gateway as API Gateway (HTTP)
    participant Auth as Auth Server (gRPC)
    participant DB as Database / Redis
    
    Client->>Gateway: HTTP REST Request (e.g., POST /api/auth/login)
    Note over Gateway: 1. Extract IP & Metadata <br/> 2. Apply Rate Limiting <br/> 3. Parse JSON Body
    Gateway->>Auth: gRPC Call (pb.LoginRequest)
    Note over Auth: 1. Validate Business Logic <br/> 2. Hash Comparisons
    Auth->>DB: Query User Data
    DB-->>Auth: Result
    Auth-->>Gateway: gRPC Response (pb.LoginResponse)
    Note over Gateway: Translate gRPC Response to JSON
    Gateway-->>Client: HTTP JSON Response
```

## 2. API Interaction Flows

### 2.1 Rate Limiting Architecture Details

Rate limiting is enforced exclusively at the **Gateway layer** using middleware (`internal/gateway/middleware/rate_limit.go`). It uses a Sliding Window algorithm. 
This protects the deeper Auth server and databases from spam, brute-force attacks, and distributed denial-of-service attempts.

```mermaid
sequenceDiagram
    participant Client as Client IP
    participant M as Rate Limit Middleware
    participant H as Route Handler
    participant A as Auth Service
    
    Client->>M: HTTP Request (e.g., /register or /check-username)
    Note over M: Extract Client IP or Email
    M->>M: Check Sliding Window Limits
    
    alt Limit Exceeded
        M-->>Client: HTTP 429 Too Many Requests <br/> {"error": "...", "retry_in": "X seconds"}
    else Limit Not Exceeded
        M->>H: Proceed to Route Handler
        H->>A: gRPC Request
        A-->>H: gRPC Response
        H-->>Client: HTTP Response
        Note over M: Check completes & Increments Rate Limit Counter
    end
```

**Key Points:**
- **Registration IP Limit**: Restricts how many accounts can be created from a single IP within a timeframe.
- **Email Request Limit**: Prevents spamming OTP codes to the same email address.
- **Username Check Limit**: Prevents malicious actors from enumerating usernames during the sign-up phase.

### 2.2 Login Flow (Including TOTP and Verification Validation)

The login process is multi-faceted. The Auth Server checks credentials, ensures the email is verified, and determines if Multi-Factor Authentication (TOTP) is mandated.

```mermaid
sequenceDiagram
    participant Client
    participant Gateway as Gateway (/api/auth/login)
    participant Auth as Auth (gRPC Login)
    
    Client->>Gateway: POST /api/auth/login {username, password}
    Gateway->>Auth: pb.LoginRequest
    
    alt Invalid Credentials
        Auth-->>Gateway: Success: false, Message: "invalid credentials"
        Gateway-->>Client: HTTP 401 Unauthorized
    else Account Not Verified
        Auth-->>Gateway: IsVerified: false, Message: "verify email first"
        Gateway-->>Client: HTTP 403 Forbidden
    else TOTP Required
        Auth-->>Gateway: RequiresTotp: true, UserId: "id"
        Gateway-->>Client: HTTP 200 OK {"requires_totp": true}
        Note over Client: Client switches to TOTP input screen
    else Success (No TOTP)
        Auth-->>Gateway: Success: true, AccessToken, RefreshToken
        Gateway-->>Client: HTTP 200 OK {"access_token": "...", "user": {...}}
    end
```

#### Secondary Login Flow: Verifying TOTP
If the login requires TOTP, a secondary request is made:

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway (/api/auth/verify-totp)
    participant A as Auth (gRPC VerifyTOTP)
    
    C->>G: POST /api/auth/verify-totp {username, totp_code}
    G->>A: pb.VerifyTOTPRequest
    alt Invalid Code
        A-->>G: Success: false
        G-->>C: HTTP 400 Bad Request
    else Valid Code
        A-->>G: Success: true, AccessToken, RefreshToken
        G-->>C: HTTP 200 OK Token Payload
    end
```

### 2.3 User Registration Flow

The registration endpoint showcases how the gateway implements IP and email limits alongside communicating with the Auth server.

```mermaid
sequenceDiagram
    participant Client
    participant GW as Gateway Middleware & Handler
    participant Auth as Auth gRPC Server
    
    Client->>GW: POST /api/auth/register {username, email, pass}
    
    Note over GW: 1. Apply RegistrationRateLimit
    alt IP Rate Limit Failed
        GW-->>Client: HTTP 429 Too Many Requests
    else IP Block Permitted
        GW->>Auth: gRPC Register(Username, Email, Password)
        
        alt Username/Email Taken
            Auth-->>GW: Success: false
            GW-->>Client: HTTP 400 Bad Request
        else Created
            Auth-->>GW: Success: true, Email
            GW-->>Client: HTTP 201 Created (Prompt for Email Verification)
        end
    end
```

### 2.4 Token Refresh Flow

Access tokens have limited lifespans for security reasons. Clients use the refresh token to maintain a session.

```mermaid
sequenceDiagram
    participant Client
    participant Gateway as Gateway (/api/auth/refresh)
    participant Auth as Auth Server
    
    Note over Client: Access Token Expired
    Client->>Gateway: POST /api/auth/refresh {refresh_token}
    Gateway->>Auth: gRPC RefreshToken(RefreshToken)
    
    alt Token Valid
        Auth-->>Gateway: Success: true, New Access/Refresh Tokens
        Gateway-->>Client: HTTP 200 OK (New Tokens)
    else Token Invalid/Revoked
        Auth-->>Gateway: Success: false
        Gateway-->>Client: HTTP 401 Unauthorized (Force Re-login)
    end
```
