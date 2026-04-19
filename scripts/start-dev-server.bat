@echo off
REM Start Development Test Server
REM ⚠️ DEVELOPMENT ONLY - DO NOT USE IN PRODUCTION ⚠️

echo ============================================
echo Starting GoConnect Dev Test Server
echo ============================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %errorLevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Set development port
set DEV_SERVER_PORT=3000

echo Starting server on http://localhost:%DEV_SERVER_PORT%
echo.

REM Run the dev server
go run cmd\dev-server\main.go

pause
