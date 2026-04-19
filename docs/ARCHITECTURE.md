# GoConnect Architecture

## Overview

GoConnect is a scalable microservices backend for a messaging application, built with Go and designed for high availability and horizontal scaling.

## System Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ NGINX Load      │
│ Balancer        │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│   API Gateway           │
│   - JWT Validation      │
│   - Rate Limiting       │
│   - Request Routing     │
│   - CORS                │
└──────────┬──────────────┘
           │ gRPC
           ▼
    ┌──────────────┐
    │ Auth Service │
    │ - Register   │
    │ - Login      │
    │ - OAuth      │
    │ - OTP        │
    └──────┬───────┘
           │
    ┌──────┴───────┐
    │              │
    ▼              ▼
┌──────────┐  ┌─────────┐
│PostgreSQL│  │  Redis  │
└──────────┘  └─────────┘
```

## Components

### 1. API Gateway
**Technology:** Go + Gin Framework  
**Port:** 8080  
**Responsibilities:**
- HTTP/REST API endpoints
- JWT token validation
- Request routing to microservices via gRPC
- Rate limiting (100 req/min per IP)
- CORS handling
- Request/response logging
- OAuth callback handling

**Key Features:**
- Stateless design
- Horizontally scalable
- Health check endpoint
- Middleware-based architecture

### 2. Auth Service
**Technology:** Go + gRPC  
**Port:** 50051  
**Responsibilities:**
- User registration and authentication
- Password management (hashing with bcrypt)
- JWT token generation and validation
- OAuth integration (Google, Facebook, GitHub)
- OTP generation and verification (TOTP-based)
- Bloom filter for username checks

**Key Features:**
- gRPC server for inter-service communication
- Stateless OTP (no DB storage)
- Bloom filter for fast username availability
- Support for multiple authentication methods

### 3. Load Balancer
**Technology:** NGINX  
**Responsibilities:**
- Distribute traffic across gateway instances
- SSL/TLS termination
- Health checks
- Connection pooling

**Load Balancing Strategy:**
- Round-robin by default
- Least connections for better distribution
- Automatic failover

### 4. Database (PostgreSQL)
**Version:** 15+  
**Responsibilities:**
- User data storage
- OAuth account mapping
- Refresh token management

**Tables:**
- `users`: User profiles and credentials
- `oauth_accounts`: OAuth provider mappings
- `refresh_tokens`: JWT refresh tokens

**Optimization:**
- Indexes on username, email
- Connection pooling (max 25 connections)
- Automatic timestamp updates

### 5. Cache (Redis)
**Version:** 7+  
**Responsibilities:**
- Bloom filter storage for usernames
- Session caching (future)
- Rate limiting counters

## Data Flow

### Registration Flow
```
Client → Gateway → Auth Service → PostgreSQL
                 → Redis (Bloom Filter)
       ← Gateway ← Auth Service (JWT tokens)
```

### Login Flow
```
Client → Gateway → Auth Service → PostgreSQL (verify credentials)
       ← Gateway ← Auth Service (JWT tokens)
```

### OAuth Flow
```
Client → Gateway → OAuth Provider (Google/Facebook/GitHub)
Provider → Gateway (callback with code)
Gateway → Auth Service → OAuth Provider (exchange code)
       → PostgreSQL (create/update user)
       ← JWT tokens
```

### OTP Flow
```
Client → Gateway → Auth Service (generate OTP using TOTP)
       ← Gateway ← OTP (sent via email in production)

Client → Gateway → Auth Service (verify OTP)
       → Auth Service (validate using TOTP)
       ← Gateway ← Success/Failure
```

## Security Architecture

### Authentication
- **Password Storage**: Bcrypt with cost factor 12
- **JWT Tokens**: 
  - Access Token: 15 minutes expiry
  - Refresh Token: 7 days expiry, stored in DB
- **OTP**: TOTP-based, 5-minute validity, stateless

### Authorization
- JWT token validation at gateway
- User context propagated via token claims
- Role-based access control (future)

### Data Protection
- All passwords hashed before storage
- Sensitive data encrypted at rest
- TLS/SSL for data in transit
- OAuth tokens securely stored

### Rate Limiting
- 100 requests per minute per IP
- Prevents brute force attacks
- Applied at gateway level

## Scalability Design

### Horizontal Scaling
- All services are stateless
- Can scale independently
- No session affinity required
- Database connection pooling

### Load Distribution
- NGINX load balancer
- Multiple gateway instances
- Multiple auth service instances
- Kubernetes HPA for auto-scaling

### Caching Strategy
- Bloom filter for username checks
- Redis for session data
- Reduces database load

### Database Optimization
- Indexed queries
- Connection pooling
- Read replicas (future)
- Partitioning (future)

## Communication Patterns

### Inter-Service Communication
- **Protocol**: gRPC
- **Format**: Protocol Buffers
- **Benefits**: 
  - Type-safe
  - High performance
  - Bi-directional streaming support

### Client Communication
- **Protocol**: HTTP/REST
- **Format**: JSON
- **Benefits**: 
  - Wide compatibility
  - Easy to use
  - Browser-friendly

## Deployment Architecture

### Development
```
Docker Compose
├── PostgreSQL (1 instance)
├── Redis (1 instance)
├── Auth Service (1 instance)
└── Gateway (1 instance)
```

### Production (Docker Compose)
```
Docker Compose + NGINX
├── PostgreSQL (1 instance + backup)
├── Redis (1 instance)
├── Auth Service (3 replicas)
├── Gateway (3 replicas)
└── NGINX Load Balancer
```

### Production (Kubernetes)
```
Kubernetes Cluster
├── Namespace: goconnect
├── StatefulSet: PostgreSQL
├── StatefulSet: Redis
├── Deployment: Auth Service (3 replicas, HPA)
├── Deployment: Gateway (3 replicas, HPA)
├── Service: LoadBalancer (external access)
└── Ingress: SSL/TLS termination
```

## Monitoring and Observability

### Logging
- Structured logging with Zap
- Log levels: debug, info, warn, error
- Request/response logging at gateway
- Error tracking and alerting

### Health Checks
- Gateway: `/health` endpoint
- Kubernetes liveness/readiness probes
- Database connection checks
- Service dependency checks

### Metrics (Future)
- Prometheus metrics
- Request rate and latency
- Error rates
- Resource utilization

## Future Enhancements

### Phase 2 Services
1. **User Service**
   - Profile management
   - Avatar upload
   - Contact list
   - User search

2. **Message Service**
   - 1-to-1 messaging
   - Message history
   - Real-time delivery (WebSocket)
   - Message encryption

3. **Group Service**
   - Group creation/management
   - Member management
   - Group messaging
   - Admin controls

### Additional Features
- WebSocket support for real-time messaging
- Message queue (RabbitMQ/Kafka)
- File storage (S3/MinIO)
- Push notifications
- Message search (Elasticsearch)
- Analytics and reporting

## Technology Stack

### Backend
- **Language**: Go 1.23+
- **HTTP Framework**: Gin
- **RPC Framework**: gRPC
- **Database**: PostgreSQL 18+
- **Cache**: Redis 7+
- **Load Balancer**: NGINX

### Libraries
- `github.com/gin-gonic/gin` - HTTP framework
- `google.golang.org/grpc` - gRPC
- `github.com/google/uuid` - GUID generation
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/golang-jwt/jwt/v5` - JWT tokens
- `github.com/bits-and-blooms/bloom/v3` - Bloom filter
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/go-redis/redis/v8` - Redis client
- `golang.org/x/oauth2` - OAuth 2.0
- `github.com/spf13/viper` - Configuration
- `go.uber.org/zap` - Logging

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **CI/CD**: GitHub Actions (future)
- **Cloud**: AWS/GCP/Azure compatible

## Design Principles

1. **Microservices**: Loosely coupled, independently deployable
2. **Stateless**: All services are stateless for easy scaling
3. **Security First**: Authentication, authorization, encryption
4. **Observability**: Comprehensive logging and monitoring
5. **Fault Tolerance**: Health checks, retries, circuit breakers
6. **Performance**: Caching, connection pooling, optimization
7. **Maintainability**: Clean code, documentation, testing

## Best Practices

### Code Organization
- Clear separation of concerns
- Repository pattern for data access
- Service layer for business logic
- Handler layer for API endpoints

### Error Handling
- Graceful error handling
- Meaningful error messages
- Error logging and tracking
- No sensitive data in errors

### Configuration Management
- Environment-based configuration
- Secrets in secure storage
- Configuration validation
- Sensible defaults

### Testing Strategy
- Unit tests for business logic
- Integration tests for API endpoints
- End-to-end tests for workflows
- Load testing for scalability
