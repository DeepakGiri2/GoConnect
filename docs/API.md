# GoConnect API Documentation

Base URL: `http://localhost:8080/api`

## Authentication Endpoints

### 1. Register

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePass123"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "registration successful",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Validation Rules:**
- Username: 3-50 characters, alphanumeric and underscore only
- Email: Valid email format
- Password: Minimum 8 characters, must contain uppercase, lowercase, and number

---

### 2. Login

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "SecurePass123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "login successful",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### 3. Refresh Token

**Endpoint:** `POST /auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### 4. Forgot Password

**Endpoint:** `POST /auth/forgot-password`

**Request Body:**
```json
{
  "email": "john@example.com"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "OTP sent to your email"
}
```

**Note:** In production, send OTP via email. For development, check server logs for the OTP.

---

### 5. Verify OTP

**Endpoint:** `POST /auth/verify-otp`

**Request Body:**
```json
{
  "email": "john@example.com",
  "otp": "123456"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "OTP verified successfully",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### 6. Reset Password

**Endpoint:** `POST /auth/reset-password`

**Request Body:**
```json
{
  "email": "john@example.com",
  "otp": "123456",
  "new_password": "NewSecurePass123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "password reset successful"
}
```

---

### 7. Check Username Availability

**Endpoint:** `GET /auth/check-username?username=johndoe`

**Response (200 OK):**
```json
{
  "available": false
}
```

---

## OAuth Endpoints

### 1. Initiate OAuth (Google)

**Endpoint:** `GET /auth/oauth/google`

**Response (200 OK):**
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...",
  "state": "random_state_string"
}
```

**Frontend Flow:**
1. Call this endpoint
2. Redirect user to `auth_url`
3. User authenticates with Google
4. Google redirects to callback URL with code

---

### 2. OAuth Callback (Google)

**Endpoint:** `GET /auth/callback/google?code=...&state=...`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "OAuth login successful",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Note:** Same endpoints available for Facebook (`/auth/oauth/facebook`) and GitHub (`/auth/oauth/github`)

---

## Protected Endpoints

All protected endpoints require the `Authorization` header with Bearer token.

**Header:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Get Current User

**Endpoint:** `GET /me`

**Response (200 OK):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "johndoe",
  "email": "john@example.com"
}
```

---

## Health Check

**Endpoint:** `GET /health`

**Response (200 OK):**
```json
{
  "status": "healthy"
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid username format"
}
```

### 401 Unauthorized
```json
{
  "error": "invalid or expired token"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate limit exceeded"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Authentication Flow

### Username/Password Authentication
1. Register: `POST /auth/register`
2. Login: `POST /auth/login`
3. Use access token for protected endpoints
4. Refresh when expired: `POST /auth/refresh`

### OAuth Authentication
1. Initiate: `GET /auth/oauth/{provider}`
2. User authenticates with provider
3. Callback: `GET /auth/callback/{provider}`
4. Use access token for protected endpoints

### Password Reset Flow
1. Request OTP: `POST /auth/forgot-password`
2. Verify OTP: `POST /auth/verify-otp`
3. Reset password: `POST /auth/reset-password`
4. Login with new password

---

## Rate Limiting

- 100 requests per minute per IP address
- Applies to all endpoints

---

## Token Expiry

- Access Token: 15 minutes
- Refresh Token: 7 days
- OTP: 5 minutes

---

## Testing with cURL

### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe","email":"john@example.com","password":"SecurePass123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe","password":"SecurePass123"}'
```

### Protected Endpoint
```bash
curl -X GET http://localhost:8080/api/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```
