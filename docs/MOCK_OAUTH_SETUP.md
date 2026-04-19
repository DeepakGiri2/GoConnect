# Mock OAuth Setup for Local Development

This guide explains how to set up mock OAuth services (Google, Facebook, GitHub) for local development **without any code changes**.

## 🎯 Overview

The mock OAuth solution allows you to test OAuth login flows locally without:
- Creating OAuth apps on Google/Facebook/GitHub
- Using real OAuth credentials
- Making external API calls during development

## 📋 Prerequisites

- Docker installed and running
- Administrator/root access (for hosts file modification)
- Ports 9000, 9001, 9002 available

## 🚀 Quick Start (Windows)

### Option 1: Using Setup Script (Recommended)

1. **Run the setup script as Administrator**:
   ```powershell
   # Right-click and select "Run as administrator"
   .\scripts\setup-mock-oauth.bat
   ```

2. **Copy the local environment file**:
   ```powershell
   copy .env.local .env
   ```

3. **Start the mock OAuth server**:
   ```powershell
   cd build\docker
   docker compose -f docker-compose.dev.yml up mock-oauth -d
   ```

4. **Start your GoConnect services**:
   ```powershell
   # Start all services
   docker compose -f docker-compose.dev.yml up
   
   # OR run locally
   go run cmd\auth\main.go
   go run cmd\gateway\main.go
   ```

### Option 2: Manual Setup

1. **Edit hosts file** (requires Administrator):
   - Open: `C:\Windows\System32\drivers\etc\hosts`
   - Add these lines:
     ```
     # GoConnect Mock OAuth - Local Development
     127.0.0.1 accounts.google.com
     127.0.0.1 www.facebook.com
     127.0.0.1 github.com
     # End GoConnect Mock OAuth
     ```

2. **Flush DNS cache**:
   ```powershell
   ipconfig /flushdns
   ```

3. **Follow steps 2-4 from Option 1**

## 🔧 How It Works

### Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   Browser   │────────▶│  Mock OAuth      │◀───────▶│  GoConnect  │
│             │         │  Server          │         │  Services   │
│             │         │  (Port 9000-9002)│         │             │
└─────────────┘         └──────────────────┘         └─────────────┘
                               │
                        ┌──────┴──────┐
                        │ Hosts File  │
                        │ Redirection │
                        └─────────────┘
```

### Component Breakdown

1. **Hosts File Redirection**: 
   - Redirects OAuth provider domains to `127.0.0.1`
   - No code changes needed

2. **Mock OAuth Server**:
   - **Google Mock** (Port 9000): Simulates Google OAuth
   - **Facebook Mock** (Port 9001): Simulates Facebook OAuth  
   - **GitHub Mock** (Port 9002): Simulates GitHub OAuth

3. **Mock User Data**:
   ```javascript
   // Google
   {
     id: "108123456789012345678",
     email: "mockuser@gmail.com",
     name: "Mock Google User"
   }
   
   // Facebook
   {
     id: "1234567890123456",
     email: "mockuser@facebook.com",
     name: "Mock Facebook User"
   }
   
   // GitHub
   {
     id: 12345678,
     login: "mockgithubuser",
     email: "mockuser@github.com"
   }
   ```

## 🧪 Testing OAuth Flow

### 1. Start the Services

```powershell
# Terminal 1: Start mock OAuth server
cd build\docker
docker compose -f docker-compose.dev.yml up mock-oauth

# Terminal 2: Start GoConnect services
go run cmd\gateway\main.go

# Terminal 3: Start auth service
go run cmd\auth\main.go
```

### 2. Test Google Login

```bash
# Initiate OAuth flow
curl http://localhost:8080/api/auth/google

# You'll receive an auth URL - open it in browser
# The mock server will automatically redirect with an auth code
```

### 3. Expected Flow

1. User clicks "Login with Google"
2. Browser → `accounts.google.com` (redirected to localhost:9000 via hosts)
3. Mock server → Generates mock auth code
4. Redirects back to your app with code
5. Your app exchanges code for mock access token
6. Your app fetches mock user info
7. User logged in with mock data

## 📝 Mock OAuth Endpoints

### Google (Port 9000)
- **Authorization**: `/o/oauth2/v2/auth`
- **Token Exchange**: `/token`
- **User Info**: `/oauth2/v2/userinfo`

### Facebook (Port 9001)
- **Authorization**: `/v12.0/dialog/oauth`
- **Token Exchange**: `/v12.0/oauth/access_token`
- **User Info**: `/me`

### GitHub (Port 9002)
- **Authorization**: `/login/oauth/authorize`
- **Token Exchange**: `/login/oauth/access_token`
- **User Info**: `/user`

## 🛠️ Customizing Mock Users

Edit `build/docker/mock-oauth/server.js`:

```javascript
const mockUsers = {
  google: {
    id: 'your-custom-id',
    email: 'custom@gmail.com',
    name: 'Custom User'
  },
  // ... modify as needed
};
```

Restart the mock server after changes.

## 🔄 Switching Between Mock and Real OAuth

### To Use Mock OAuth
```powershell
.\scripts\setup-mock-oauth.bat
copy .env.local .env
docker compose -f build\docker\docker-compose.dev.yml up mock-oauth -d
```

### To Use Real OAuth
```powershell
.\scripts\remove-mock-oauth.bat
# Update .env with real OAuth credentials
# Restart services
```

## 🐛 Troubleshooting

### Issue: OAuth redirects to real Google/Facebook/GitHub

**Solution**: 
- Verify hosts file has the correct entries
- Flush DNS: `ipconfig /flushdns`
- Restart browser
- Check mock OAuth server is running

### Issue: Port already in use

**Solution**:
```powershell
# Check what's using the port
netstat -ano | findstr :9000

# Kill the process (replace PID)
taskkill /PID <PID> /F
```

### Issue: Access denied when editing hosts file

**Solution**:
- Run script as Administrator
- Disable antivirus temporarily
- Check file permissions on hosts file

### Issue: Mock OAuth server not responding

**Solution**:
```powershell
# Check if container is running
docker ps | findstr mock-oauth

# View logs
docker logs goconnect-mock-oauth

# Restart container
docker restart goconnect-mock-oauth
```

## 📊 Viewing Logs

### Mock OAuth Server Logs
```powershell
docker logs -f goconnect-mock-oauth
```

You'll see:
```
Mock Google OAuth server running on port 9000
Mock Facebook OAuth server running on port 9001
Mock GitHub OAuth server running on port 9002
[Google] Authorization request received
[Google] Redirecting to: http://localhost:8080/api/auth/callback/google?code=...
```

## ⚠️ Important Notes

1. **Internet Access**: Real OAuth providers won't work while mock is active
2. **Browser Cache**: Clear browser cache if switching between mock/real
3. **Production**: NEVER use this in production
4. **Security**: Mock server has NO security - local development only
5. **Database**: Mock users will be created in your local database

## 🧹 Cleanup

### Remove Mock OAuth Configuration
```powershell
# Run as Administrator
.\scripts\remove-mock-oauth.bat

# Stop mock server
docker compose -f build\docker\docker-compose.dev.yml stop mock-oauth
```

### Complete Reset
```powershell
# Remove all containers
docker compose -f build\docker\docker-compose.dev.yml down -v

# Remove mock OAuth entries
.\scripts\remove-mock-oauth.bat

# Restore original .env
copy .env.example .env
```

## 📚 Additional Resources

- [OAuth 2.0 Specification](https://oauth.net/2/)
- [Docker Documentation](https://docs.docker.com/)
- [GoConnect Architecture](./ARCHITECTURE.md)

## 🤝 Contributing

Found an issue or want to improve the mock OAuth setup? 
- Submit issues on GitHub
- Create pull requests with improvements
- Update documentation

---

**Last Updated**: 2026-04-18  
**Maintainer**: GoConnect Team
