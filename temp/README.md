# ⚠️ DEVELOPMENT TEST FILES ONLY ⚠️

**DO NOT USE IN PRODUCTION**

This directory contains development testing tools.

## Files

- **`index.html`** - Unified test interface (API + OAuth testing)
- **`test-frontend.html`** - Legacy API tester (use index.html instead)
- **`test-mock-oauth.html`** - Legacy OAuth tester (use index.html instead)

## Usage

### Option 1: Via Dev Server (Recommended)
```powershell
.\scripts\start-dev-server.bat
# Open: http://localhost:3000/index.html
```

### Option 2: Direct File
```powershell
start index.html
```

## Features

✅ Test all API endpoints  
✅ Mock OAuth login flows  
✅ Auto-fill tokens  
✅ View responses  
✅ Check server status  

## See Also

- `DEV_TESTING_QUICKSTART.md` - Quick start guide
- `docs/DEV_TESTING.md` - Complete documentation
- `docs/MOCK_OAUTH_SETUP.md` - Mock OAuth details

---

**REMINDER: These files are for local development testing only!**
