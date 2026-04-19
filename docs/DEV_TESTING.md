# Development Testing Guide

> ⚠️ **WARNING: DEVELOPMENT ONLY** - This configuration is for local testing and should **NEVER** be used in production!

## Overview

This guide covers the complete development testing setup for GoConnect, including:
- Mock OAuth services (Google, Facebook, GitHub)
- Unified web-based API testing interface
- Static file server for test pages

## 🚀 Quick Start

### 1. Setup Mock OAuth (One-time)

```powershell
# Run as Administrator
.\scripts\setup-mock-oauth.bat
```

This configures your hosts file to redirect OAuth providers to localhost.

### 2. Configure Environment

```powershell
# Copy development configuration
copy .env.local .env
```

### 3. Start Services

```powershell
# Terminal 1: Start mock OAuth server
cd build\docker
docker compose -f docker-compose.dev.yml up mock-oauth -d

# Terminal 2: Start GoConnect services
docker compose -f docker-compose.dev.yml up postgres redis auth-service gateway -d

# Terminal 3 (Optional): Start dev test server
.\scripts\start-dev-server.bat
```

### 4. Open Test Interface

**Option A: Via Dev Server (Recommended)**
```
http://localhost:3000/index.html
```

**Option B: Direct File**
```powershell
start temp\index.html
```

## 📁 File Structure

```
GoConnect/
├── temp/
│   └── index.html              # Unified test interface (DEV ONLY)
├── cmd/
│   └── dev-server/
│       └── main.go             # Static file server (DEV ONLY)
├── scripts/
│   ├── setup-mock-oauth.bat    # Mock OAuth setup (DEV ONLY)
│   ├── remove-mock-oauth.bat   # Cleanup script (DEV ONLY)
│   ├── start-dev-server.bat    # Dev server launcher (DEV ONLY)
│   └── test-mock-oauth.bat     # Server verification (DEV ONLY)
├── .env.local                  # Dev environment template (DEV ONLY)
└── docs/
    ├── DEV_TESTING.md          # This file (DEV ONLY)
    └── MOCK_OAUTH_SETUP.md     # Mock OAuth details (DEV ONLY)
```

## 🧪 Test Interface Features

### API Testing Tab
- **Register**: Create new user accounts
- **Login**: Authenticate with username/password
- **Refresh Token**: Get new access tokens
- **Check Username**: Verify username availability
- **Forgot Password**: Initiate password reset
- **Verify OTP**: Validate OTP codes
- **Reset Password**: Complete password reset
- **Get Current User**: Fetch authenticated user data
- **Health Check**: Verify service status

### Mock OAuth Tab
- **Google OAuth**: Test Google login flow with mock user
- **Facebook OAuth**: Test Facebook login flow with mock user
- **GitHub OAuth**: Test GitHub login flow with mock user
- **Server Status**: Check mock OAuth server health

## 🎭 Mock User Accounts

The mock OAuth server provides these predefined users:

| Provider | Email | Username | ID |
|----------|-------|----------|-----|
| Google | mockuser@gmail.com | Mock Google User | 108123456789012345678 |
| Facebook | mockuser@facebook.com | Mock Facebook User | 1234567890123456 |
| GitHub | mockuser@github.com | mockgithubuser | 12345678 |

## 🔧 Configuration Details

### Environment Variables (.env.local)

```bash
# Development Test Server
DEV_SERVER_PORT=3000

# Mock OAuth Credentials
GOOGLE_CLIENT_ID=mock_google_client_id
GOOGLE_CLIENT_SECRET=mock_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback/google

FACEBOOK_CLIENT_ID=mock_facebook_client_id
FACEBOOK_CLIENT_SECRET=mock_facebook_client_secret
FACEBOOK_REDIRECT_URL=http://localhost:8080/api/auth/callback/facebook

GITHUB_CLIENT_ID=mock_github_client_id
GITHUB_CLIENT_SECRET=mock_github_client_secret
GITHUB_REDIRECT_URL=http://localhost:8080/api/auth/callback/github
```

### Ports Used

| Service | Port | Purpose |
|---------|------|---------|
| Dev Test Server | 3000 | Serves temp/index.html |
| Gateway | 8080 | Main API gateway |
| Auth Service | 50051 | gRPC auth service |
| Mock Google | 9000 | Mock Google OAuth |
| Mock Facebook | 9001 | Mock Facebook OAuth |
| Mock GitHub | 9002 | Mock GitHub OAuth |

## 📊 Testing Workflows

### Test New User Registration

1. Open test interface
2. Go to API Testing tab
3. Fill in Register form
4. Click "Register"
5. Check response for access/refresh tokens
6. Tokens auto-fill in header fields

### Test OAuth Login Flow

1. Open test interface
2. Go to Mock OAuth tab
3. Verify all mock servers are online
4. Click "Test Google Login"
5. Automatically redirected through mock flow
6. User created/authenticated in database

### Test Password Reset

1. Use "Forgot Password" to send OTP
2. Check console/logs for OTP code
3. Use "Verify OTP" to validate code
4. Use "Reset Password" to set new password
5. Login with new credentials

## 🛠️ Troubleshooting

### Test Interface Not Loading

```powershell
# Check if dev server is running
netstat -ano | findstr :3000

# Restart dev server
.\scripts\start-dev-server.bat
```

### API Requests Failing

```powershell
# Check if services are running
docker ps

# Check service logs
docker logs goconnect-gateway
docker logs goconnect-auth

# Verify health
curl http://localhost:8080/health
```

### Mock OAuth Not Working

```powershell
# Verify hosts file
type C:\Windows\System32\drivers\etc\hosts | findstr "GoConnect"

# Check mock server
docker logs goconnect-mock-oauth

# Test mock server directly
.\scripts\test-mock-oauth.bat
```

### CORS Errors

The dev server includes CORS headers. If you're opening `index.html` directly from file system, you might encounter CORS issues. Use the dev server instead:

```powershell
.\scripts\start-dev-server.bat
# Then open http://localhost:3000/index.html
```

## 🧹 Cleanup

### Remove Mock OAuth Configuration

```powershell
# Run as Administrator
.\scripts\remove-mock-oauth.bat
```

### Stop All Services

```powershell
# Stop Docker services
cd build\docker
docker compose -f docker-compose.dev.yml down

# Stop dev server (Ctrl+C in terminal)
```

### Reset Environment

```powershell
# Remove development .env
del .env

# Restore from example
copy .env.example .env
```

## ⚠️ Security Warnings

### DO NOT in Production:
- ❌ Use mock OAuth credentials
- ❌ Run dev test server
- ❌ Expose temp directory
- ❌ Use .env.local configuration
- ❌ Redirect OAuth providers via hosts file
- ❌ Run mock OAuth server
- ❌ Use weak JWT/OTP secrets

### DO in Production:
- ✅ Use real OAuth credentials from providers
- ✅ Use production-grade secrets
- ✅ Enable proper CORS policies
- ✅ Use HTTPS for all endpoints
- ✅ Remove all development test files
- ✅ Follow security best practices

## 📝 Adding Custom Tests

### Modify Test Interface

Edit `temp/index.html`:

```html
<!-- Add new test card -->
<div class="endpoint-card bg-white rounded-lg shadow p-6">
    <h3 class="text-lg font-semibold text-gray-800 mb-3">🆕 New Feature</h3>
    <button onclick="testNewFeature()" 
            class="w-full bg-purple-500 hover:bg-purple-600 text-white font-medium py-2 px-4 rounded-md transition">
        Test Feature
    </button>
</div>

<script>
async function testNewFeature() {
    await makeRequest('/your-endpoint', 'POST', { data: 'value' });
}
</script>
```

### Add Mock OAuth Provider

Edit `build/docker/mock-oauth/server.js` to add new providers.

## 🤝 Development Tips

1. **Keep test page updated**: When adding new endpoints, update `temp/index.html`
2. **Use mock data**: Keep mock users consistent for testing
3. **Check logs**: Always verify server logs for detailed error messages
4. **Test flows**: Test complete user journeys, not just individual endpoints
5. **Clear tokens**: Clear tokens between test runs to avoid stale data

## 📚 Related Documentation

- [Mock OAuth Setup](MOCK_OAUTH_SETUP.md) - Detailed mock OAuth documentation
- [API Documentation](API.md) - Complete API reference
- [Architecture](ARCHITECTURE.md) - System architecture overview

---

**Last Updated**: 2026-04-18  
**Environment**: Development Only  
**Security Level**: Low (Development Testing)
