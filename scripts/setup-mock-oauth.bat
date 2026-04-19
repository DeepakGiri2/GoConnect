@echo off
REM Setup script for Mock OAuth on Windows
echo ============================================
echo GoConnect Mock OAuth Setup
echo ============================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator!
    echo Right-click and select "Run as administrator"
    pause
    exit /b 1
)

echo [1/3] Backing up hosts file...
copy C:\Windows\System32\drivers\etc\hosts C:\Windows\System32\drivers\etc\hosts.backup.%date:~-4,4%%date:~-10,2%%date:~-7,2%
echo Backup created: hosts.backup.%date:~-4,4%%date:~-10,2%%date:~-7,2%
echo.

echo [2/3] Adding mock OAuth entries to hosts file...

REM Check if entries already exist
findstr /C:"# GoConnect Mock OAuth" C:\Windows\System32\drivers\etc\hosts >nul
if %errorLevel% equ 0 (
    echo Mock OAuth entries already exist in hosts file.
    echo To reset, run: remove-mock-oauth.bat
) else (
    echo # GoConnect Mock OAuth - Local Development >> C:\Windows\System32\drivers\etc\hosts
    echo 127.0.0.1 accounts.google.com >> C:\Windows\System32\drivers\etc\hosts
    echo 127.0.0.1 www.facebook.com >> C:\Windows\System32\drivers\etc\hosts
    echo 127.0.0.1 github.com >> C:\Windows\System32\drivers\etc\hosts
    echo # End GoConnect Mock OAuth >> C:\Windows\System32\drivers\etc\hosts
    echo Mock OAuth entries added successfully!
)
echo.

echo [3/3] Flushing DNS cache...
ipconfig /flushdns >nul
echo DNS cache flushed.
echo.

echo ============================================
echo Setup Complete!
echo ============================================
echo.
echo IMPORTANT NOTES:
echo - Real OAuth providers (Google, Facebook, GitHub) will NOT work while this is active
echo - To restore normal OAuth, run: remove-mock-oauth.bat
echo - Mock OAuth server must be running on ports 9000, 9001, 9002
echo.
echo Next steps:
echo 1. Copy .env.local to .env
echo 2. Start the mock OAuth server: docker compose -f build/docker/docker-compose.dev.yml up mock-oauth
echo 3. Start your GoConnect services
echo.
pause
