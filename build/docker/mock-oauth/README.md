# Mock OAuth Server

A lightweight Node.js-based mock OAuth2 server for local development.

## Overview

This mock server simulates OAuth2 authorization flows for:
- **Google OAuth** (Port 9000)
- **Facebook OAuth** (Port 9001)  
- **GitHub OAuth** (Port 9002)

## Mock Users

### Google User
```json
{
  "id": "108123456789012345678",
  "email": "mockuser@gmail.com",
  "name": "Mock Google User",
  "picture": "https://via.placeholder.com/150"
}
```

### Facebook User
```json
{
  "id": "1234567890123456",
  "email": "mockuser@facebook.com",
  "name": "Mock Facebook User"
}
```

### GitHub User
```json
{
  "id": 12345678,
  "login": "mockgithubuser",
  "email": "mockuser@github.com",
  "name": "Mock GitHub User"
}
```

## Usage

### Docker (Recommended)

```bash
# From build/docker directory
docker compose -f docker-compose.dev.yml up mock-oauth
```

### Standalone

```bash
# Install dependencies
npm install

# Start server
npm start
```

## Endpoints

### Google (localhost:9000)

- `GET /o/oauth2/v2/auth` - Authorization endpoint
- `POST /token` - Token exchange endpoint
- `GET /oauth2/v2/userinfo` - User info endpoint

### Facebook (localhost:9001)

- `GET /v12.0/dialog/oauth` - Authorization endpoint
- `GET /v12.0/oauth/access_token` - Token exchange endpoint
- `GET /me` - User info endpoint

### GitHub (localhost:9002)

- `GET /login/oauth/authorize` - Authorization endpoint
- `POST /login/oauth/access_token` - Token exchange endpoint
- `GET /user` - User info endpoint

## Testing

### Manual Test

```bash
# Google OAuth flow
curl "http://localhost:9000/o/oauth2/v2/auth?redirect_uri=http://localhost:8080/callback&state=test123"

# Exchange code for token
curl -X POST http://localhost:9000/token \
  -d "code=mock_google_auth_code_123&grant_type=authorization_code"

# Get user info
curl -H "Authorization: Bearer mock_google_access_token_123" \
  http://localhost:9000/oauth2/v2/userinfo
```

## Customization

Edit `server.js` to customize:
- Mock user data
- Response formats
- Additional OAuth providers
- Custom business logic

## Security Notice

⚠️ **FOR DEVELOPMENT ONLY**
- No authentication/authorization
- No data validation
- No security features
- DO NOT use in production

## Logs

The server logs all incoming requests:

```
[Google] Authorization request received
[Google] Redirecting to: http://localhost:8080/api/auth/callback/google?code=...
[Google] Token exchange request received
[Google] User info request received
```

## Troubleshooting

**Port already in use:**
```bash
# Find process using port
netstat -ano | findstr :9000

# Kill process (Windows)
taskkill /PID <PID> /F
```

**Server not responding:**
```bash
# Check logs
docker logs goconnect-mock-oauth

# Restart
docker restart goconnect-mock-oauth
```

## License

MIT
