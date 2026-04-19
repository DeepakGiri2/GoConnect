# gRPC Communication Setup

This document explains how the Auth Service and Gateway communicate via gRPC.

## Architecture

```
┌─────────────────┐                    ┌──────────────────┐
│  API Gateway    │  ←── gRPC ──→      │  Auth Service    │
│  (HTTP/REST)    │                    │  (gRPC Server)   │
│  Port: 8080     │                    │  Port: 50051     │
└─────────────────┘                    └──────────────────┘
```

## gRPC Service Definition

Location: `shared/proto/auth.proto`

### Service Methods
```protobuf
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc GenerateOTP(GenerateOTPRequest) returns (GenerateOTPResponse);
  rpc VerifyOTP(VerifyOTPRequest) returns (VerifyOTPResponse);
  rpc ResetPassword(ResetPasswordRequest) returns (ResetPasswordResponse);
  rpc OAuthLogin(OAuthLoginRequest) returns (OAuthLoginResponse);
  rpc CheckUsernameAvailability(CheckUsernameRequest) returns (CheckUsernameResponse);
}
```

## Auth Service (gRPC Server)

**Implementation:** `services/auth/grpc/server.go`

### Server Setup
```go
// In services/auth/cmd/main.go
server := grpc.NewServer()
pb.RegisterAuthServiceServer(server, grpcServer)
listener.Listen("tcp", ":50051")
server.Serve(listener)
```

### Key Features
- Listens on port 50051
- Implements all RPC methods defined in protobuf
- Returns structured responses
- Handles errors gracefully

## Gateway (gRPC Client)

**Implementation:** `services/gateway/cmd/main.go`

### Client Setup
```go
// Connect to Auth Service
authServiceAddr := "localhost:50051"
conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
authClient := pb.NewAuthServiceClient(conn)
```

### Usage Example
```go
// In services/gateway/handlers/auth_handler.go
resp, err := h.authClient.Register(context.Background(), &pb.RegisterRequest{
    Username: req.Username,
    Email:    req.Email,
    Password: req.Password,
})
```

## Communication Flow

### Example: User Registration

1. **Client → Gateway (HTTP)**
   ```
   POST /api/auth/register
   {
     "username": "johndoe",
     "email": "john@example.com",
     "password": "SecurePass123"
   }
   ```

2. **Gateway → Auth Service (gRPC)**
   ```protobuf
   RegisterRequest {
     username: "johndoe"
     email: "john@example.com"
     password: "SecurePass123"
   }
   ```

3. **Auth Service → Gateway (gRPC)**
   ```protobuf
   RegisterResponse {
     success: true
     message: "registration successful"
     user_id: "550e8400-e29b-41d4-a716-446655440000"
     username: "johndoe"
     email: "john@example.com"
     access_token: "eyJhbGci..."
     refresh_token: "eyJhbGci..."
   }
   ```

4. **Gateway → Client (HTTP)**
   ```json
   {
     "success": true,
     "message": "registration successful",
     "user": {
       "id": "550e8400-e29b-41d4-a716-446655440000",
       "username": "johndoe",
       "email": "john@example.com"
     },
     "access_token": "eyJhbGci...",
     "refresh_token": "eyJhbGci..."
   }
   ```

## Configuration

### Development Environment

**Auth Service:**
```env
AUTH_SERVICE_HOST=localhost
AUTH_SERVICE_PORT=50051
```

**Gateway:**
```env
AUTH_SERVICE_HOST=localhost
AUTH_SERVICE_PORT=50051
```

### Docker Compose

**Auth Service:**
```yaml
auth-service:
  ports:
    - "50051:50051"
  environment:
    AUTH_SERVICE_PORT: 50051
```

**Gateway:**
```yaml
gateway:
  environment:
    AUTH_SERVICE_HOST: auth-service
    AUTH_SERVICE_PORT: 50051
  depends_on:
    - auth-service
```

### Kubernetes

**Auth Service:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
spec:
  selector:
    app: auth-service
  ports:
  - port: 50051
    targetPort: 50051
  type: ClusterIP
```

**Gateway:**
```yaml
env:
- name: AUTH_SERVICE_HOST
  value: "auth-service"
- name: AUTH_SERVICE_PORT
  value: "50051"
```

## Generate Protobuf Code

### Prerequisites
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generate
```bash
# Windows
.\scripts\generate-proto.bat

# Or manually
protoc --go_out=. --go_opt=paths=source_relative ^
       --go-grpc_out=. --go-grpc_opt=paths=source_relative ^
       shared/proto/auth.proto
```

## Testing gRPC Communication

### 1. Start Auth Service
```bash
cd services\auth
go run cmd\main.go
```

### 2. Start Gateway
```bash
cd services\gateway
go run cmd\main.go
```

### 3. Test Endpoint
```bash
curl -X POST http://localhost:8080/api/auth/register ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"test\",\"email\":\"test@example.com\",\"password\":\"Test1234\"}"
```

## Using grpcurl for Testing

### Install grpcurl
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### List Services
```bash
grpcurl -plaintext localhost:50051 list
```

### Call Method
```bash
grpcurl -plaintext -d '{"username":"test","email":"test@example.com","password":"Test1234"}' ^
  localhost:50051 auth.AuthService/Register
```

## Error Handling

### Connection Errors
```go
conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal("Failed to connect to Auth Service:", err)
}
defer conn.Close()
```

### RPC Errors
```go
resp, err := h.authClient.Login(ctx, &pb.LoginRequest{...})
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    return
}

if !resp.Success {
    c.JSON(http.StatusUnauthorized, gin.H{"error": resp.Message})
    return
}
```

## Best Practices

1. **Connection Pooling**: gRPC connections are long-lived and reused
2. **Context Timeouts**: Always use context with timeout for RPC calls
3. **Error Handling**: Check both connection errors and response errors
4. **Service Discovery**: Use environment variables for service addresses
5. **Health Checks**: Implement gRPC health checking protocol
6. **Monitoring**: Log all RPC calls for debugging

## Performance Optimization

### Connection Reuse
```go
// Create once, reuse many times
authClient := pb.NewAuthServiceClient(conn)
```

### Context with Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := h.authClient.Login(ctx, &pb.LoginRequest{...})
```

### Load Balancing
For multiple Auth Service instances, use gRPC load balancing:
```go
conn, err := grpc.Dial(
    "dns:///auth-service:50051",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

## Troubleshooting

### "Connection refused"
- Verify Auth Service is running
- Check port 50051 is not blocked
- Verify service host/port configuration

### "Unimplemented method"
- Regenerate protobuf code
- Verify server implements all methods
- Check protobuf versions match

### "Context deadline exceeded"
- Increase timeout duration
- Check network latency
- Verify service is responsive

## Security

### TLS/SSL (Production)
```go
// Server
creds, _ := credentials.NewServerTLSFromFile(certFile, keyFile)
server := grpc.NewServer(grpc.Creds(creds))

// Client
creds, _ := credentials.NewClientTLSFromFile(certFile, "")
conn, _ := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
```

### Authentication
Implement token-based authentication using gRPC metadata:
```go
md := metadata.Pairs("authorization", "Bearer "+token)
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

## Additional Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers](https://protobuf.dev/)
- [gRPC-Go Examples](https://github.com/grpc/grpc-go/tree/master/examples)
