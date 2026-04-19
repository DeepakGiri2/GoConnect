# Gateway Service - Architecture and Edge Protection Features

The **Gateway Service** (`cmd/gateway`) acts as the front door and Edge API boundary for the GoConnect application. It exposes public-facing HTTP/REST interfaces leveraging the Gin (`gin-gonic/gin`) web framework and acts primarily as an orchestrator and defensive perimeter rather than executing business logic directly.

## 1. Comprehensive Feature Breakdown

### 1.1 High-Performance Sliding Window Limiter
The crown jewel of the Gateway's internal structure is `pkg/ratelimit/sliding_window.go`, which intercepts abusive connections right at the TCP/HTTP border:
- **How it works**: Unlike a "fixed window" counter (which resets on a clock minute and can be blitzed at the 59th second), the custom Gateway limiter maps individual attempts as discrete time-series plots inside of Redis. 
- **Endpoint Specificity**: The limiter tracks isolated parameters concurrently:
  - `CheckIPRegistrationLimit`: Halts a single IP (botnet node) from registering dozens of accounts.
  - `CheckEmailRegistrationLimit`: Prevents harassment via email-spam payloads.
  - `CheckUsernameCheckLimit`: Curtails rapid iteration (enumeration attacks) hitting the `/check-username` path.
- All failed rate metrics instantly yield `HTTP 429 Too Many Requests` terminating the connection entirely before it reaches gRPC.

### 1.2 Protocol Translation (HTTP to gRPC)
Inside `internal/gateway/handlers/auth_handler.go`, the Gateway bridges two disparate communication technologies:
- **Binding & Parsing**: The Gateway handles decoding client JSON structures (`ShouldBindJSON`), validating primitive forms (e.g. `binding:"required,email"`). If validation fails, it safely constructs a 400 Bad Request error without notifying the Auth server.
- **gRPC Marshaling**: Upon a clean payload, it translates the HTTP constructs into Protobuf (`pb.LoginRequest`), establishes context streams, and dials the internal Auth server asynchronously. 
- **Graceful Re-mapping**: Interprets structured gRPC backend errors (like "requires_totp" states or "unverified flags"), flattening them into elegant, standard-compliant HTTP API payloads carrying `access_token` scopes.

### 1.3 Client IP Extraction Security
Located in `internal/gateway/middleware/ip_extractor.go`:
- Prevents basic IP spoofing. The service looks rigorously through `X-Forwarded-For` and `X-Real-IP` proxies, parsing proxy chains properly to evaluate the originating client's actual Internet address. This prevents a user from spoofing headers to evade the Sliding Window limiters.

### 1.4 API Edge Defenses 
- **CORS Mitigation**: Defines rigid Cross-Origin security profiles filtering unexpected web domains. 
- **Panic Protection**: Deploys `gin.Recovery()` guaranteeing that no matter how malformed a JSON or client request is, the Go web socket will not panic and crash the horizontal scaling pods.

---

## 2. Infrastructure Dependencies and Justifications

### A. The Auth Service (gRPC Server)
The operational backbone that computes the Gateway's requests.
- **What it does**: Connected via an insecure (internal network only) dialed gRPC socket on startup. 
- **Why it's used**: Decoupling the Gateway from the Authorization logic allows devops to scale out the Gateway pods infinitely (to handle mass HTTP socket ingestion from global load balancers) independently of the core Database-bound Auth backend.

### B. Redis Cluster
Crucial to the survivability of the Gateway.
- **What it does**: Holds the distributed counts and expiration metrics for the `SlidingWindowLimiter`.
- **Why it's used**: Consider the system scaling to 5 Gateway Pods. If rate limits were held in Go's local memory per instance, an attacker could rotate requests cleanly across the 5 instances. By extracting the state externally to Redis, the sliding windows reflect a global, multi-instance truth. A hit against Gateway Pod A affects the strict budget observed by Gateway Pod C instantly.
