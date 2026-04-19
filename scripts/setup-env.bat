@echo off
REM GoConnect Environment Setup Script
REM This script helps you set up your .env file

echo ============================================
echo GoConnect Environment Setup
echo ============================================
echo.

if exist .env (
    echo WARNING: .env file already exists!
    set /p OVERWRITE="Do you want to overwrite it? (y/n): "
    if /i not "%OVERWRITE%"=="y" (
        echo Setup cancelled.
        exit /b 0
    )
)

echo Creating .env file from template...
copy .env.example .env

echo.
echo ============================================
echo Configuration Sections:
echo ============================================
echo.
echo 1. Database Configuration
echo 2. Redis Configuration
echo 3. JWT Configuration
echo 4. OTP Configuration
echo 5. OAuth Configuration (Google, GitHub, Facebook)
echo 6. Email Configuration
echo.

set /p CONFIG_DB="Do you want to configure Database settings? (y/n): "
if /i "%CONFIG_DB%"=="y" (
    echo.
    echo --- Database Configuration ---
    set /p DB_HOST="Database Host [localhost]: "
    set /p DB_PORT="Database Port [5432]: "
    set /p DB_NAME="Database Name [goconnect]: "
    set /p DB_USER="Database User [postgres]: "
    set /p DB_PASS="Database Password: "
    
    if not "%DB_HOST%"=="" powershell -Command "(gc .env) -replace 'DATABASE_HOST=.*', 'DATABASE_HOST=%DB_HOST%' | Out-File -encoding ASCII .env"
    if not "%DB_PORT%"=="" powershell -Command "(gc .env) -replace 'DATABASE_PORT=.*', 'DATABASE_PORT=%DB_PORT%' | Out-File -encoding ASCII .env"
    if not "%DB_NAME%"=="" powershell -Command "(gc .env) -replace 'DATABASE_NAME=.*', 'DATABASE_NAME=%DB_NAME%' | Out-File -encoding ASCII .env"
    if not "%DB_USER%"=="" powershell -Command "(gc .env) -replace 'DATABASE_USER=.*', 'DATABASE_USER=%DB_USER%' | Out-File -encoding ASCII .env"
    if not "%DB_PASS%"=="" powershell -Command "(gc .env) -replace 'DATABASE_PASSWORD=.*', 'DATABASE_PASSWORD=%DB_PASS%' | Out-File -encoding ASCII .env"
)

echo.
set /p CONFIG_REDIS="Do you want to configure Redis settings? (y/n): "
if /i "%CONFIG_REDIS%"=="y" (
    echo.
    echo --- Redis Configuration ---
    set /p REDIS_HOST="Redis Host [localhost]: "
    set /p REDIS_PORT="Redis Port [6379]: "
    set /p REDIS_PASS="Redis Password (leave empty if none): "
    
    if not "%REDIS_HOST%"=="" powershell -Command "(gc .env) -replace 'REDIS_HOST=.*', 'REDIS_HOST=%REDIS_HOST%' | Out-File -encoding ASCII .env"
    if not "%REDIS_PORT%"=="" powershell -Command "(gc .env) -replace 'REDIS_PORT=.*', 'REDIS_PORT=%REDIS_PORT%' | Out-File -encoding ASCII .env"
    if not "%REDIS_PASS%"=="" powershell -Command "(gc .env) -replace 'REDIS_PASSWORD=.*', 'REDIS_PASSWORD=%REDIS_PASS%' | Out-File -encoding ASCII .env"
)

echo.
echo --- Security Configuration ---
echo Generating secure secrets...

REM Generate JWT Secret
for /f "delims=" %%i in ('powershell -Command "[Convert]::ToBase64String((1..64 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))"') do set JWT_SECRET=%%i
powershell -Command "(gc .env) -replace 'JWT_SECRET=.*', 'JWT_SECRET=%JWT_SECRET%' | Out-File -encoding ASCII .env"
echo JWT Secret generated and saved.

REM Generate OTP Secret
for /f "delims=" %%i in ('powershell -Command "[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))"') do set OTP_SECRET=%%i
powershell -Command "(gc .env) -replace 'OTP_SECRET=.*', 'OTP_SECRET=%OTP_SECRET%' | Out-File -encoding ASCII .env"
echo OTP Secret generated and saved.

echo.
set /p CONFIG_OAUTH="Do you want to configure OAuth providers? (y/n): "
if /i "%CONFIG_OAUTH%"=="y" (
    echo.
    echo --- OAuth Configuration ---
    echo.
    echo Google OAuth:
    set /p GOOGLE_ID="Google Client ID (leave empty to skip): "
    set /p GOOGLE_SECRET="Google Client Secret: "
    
    if not "%GOOGLE_ID%"=="" powershell -Command "(gc .env) -replace 'GOOGLE_CLIENT_ID=.*', 'GOOGLE_CLIENT_ID=%GOOGLE_ID%' | Out-File -encoding ASCII .env"
    if not "%GOOGLE_SECRET%"=="" powershell -Command "(gc .env) -replace 'GOOGLE_CLIENT_SECRET=.*', 'GOOGLE_CLIENT_SECRET=%GOOGLE_SECRET%' | Out-File -encoding ASCII .env"
    
    echo.
    echo GitHub OAuth:
    set /p GITHUB_ID="GitHub Client ID (leave empty to skip): "
    set /p GITHUB_SECRET="GitHub Client Secret: "
    
    if not "%GITHUB_ID%"=="" powershell -Command "(gc .env) -replace 'GITHUB_CLIENT_ID=.*', 'GITHUB_CLIENT_ID=%GITHUB_ID%' | Out-File -encoding ASCII .env"
    if not "%GITHUB_SECRET%"=="" powershell -Command "(gc .env) -replace 'GITHUB_CLIENT_SECRET=.*', 'GITHUB_CLIENT_SECRET=%GITHUB_SECRET%' | Out-File -encoding ASCII .env"
    
    echo.
    echo Facebook OAuth:
    set /p FACEBOOK_ID="Facebook Client ID (leave empty to skip): "
    set /p FACEBOOK_SECRET="Facebook Client Secret: "
    
    if not "%FACEBOOK_ID%"=="" powershell -Command "(gc .env) -replace 'FACEBOOK_CLIENT_ID=.*', 'FACEBOOK_CLIENT_ID=%FACEBOOK_ID%' | Out-File -encoding ASCII .env"
    if not "%FACEBOOK_SECRET%"=="" powershell -Command "(gc .env) -replace 'FACEBOOK_CLIENT_SECRET=.*', 'FACEBOOK_CLIENT_SECRET=%FACEBOOK_SECRET%' | Out-File -encoding ASCII .env"
)

echo.
set /p CONFIG_EMAIL="Do you want to configure Email settings? (y/n): "
if /i "%CONFIG_EMAIL%"=="y" (
    echo.
    echo --- Email Configuration ---
    echo.
    echo Common SMTP Servers:
    echo 1. Gmail: smtp.gmail.com:587
    echo 2. SendGrid: smtp.sendgrid.net:587
    echo 3. AWS SES: email-smtp.us-east-1.amazonaws.com:587
    echo.
    set /p SMTP_HOST="SMTP Host [smtp.gmail.com]: "
    set /p SMTP_PORT="SMTP Port [587]: "
    set /p SMTP_USER="SMTP Username (email): "
    set /p SMTP_PASS="SMTP Password (App Password for Gmail): "
    set /p FROM_EMAIL="From Email [noreply@goconnect.com]: "
    set /p FROM_NAME="From Name [GoConnect]: "
    
    if not "%SMTP_HOST%"=="" powershell -Command "(gc .env) -replace 'SMTP_HOST=.*', 'SMTP_HOST=%SMTP_HOST%' | Out-File -encoding ASCII .env"
    if not "%SMTP_PORT%"=="" powershell -Command "(gc .env) -replace 'SMTP_PORT=.*', 'SMTP_PORT=%SMTP_PORT%' | Out-File -encoding ASCII .env"
    if not "%SMTP_USER%"=="" powershell -Command "(gc .env) -replace 'SMTP_USERNAME=.*', 'SMTP_USERNAME=%SMTP_USER%' | Out-File -encoding ASCII .env"
    if not "%SMTP_PASS%"=="" powershell -Command "(gc .env) -replace 'SMTP_PASSWORD=.*', 'SMTP_PASSWORD=%SMTP_PASS%' | Out-File -encoding ASCII .env"
    if not "%FROM_EMAIL%"=="" powershell -Command "(gc .env) -replace 'FROM_EMAIL=.*', 'FROM_EMAIL=%FROM_EMAIL%' | Out-File -encoding ASCII .env"
    if not "%FROM_NAME%"=="" powershell -Command "(gc .env) -replace 'FROM_NAME=.*', 'FROM_NAME=%FROM_NAME%' | Out-File -encoding ASCII .env"
)

echo.
echo ============================================
echo Setup Complete!
echo ============================================
echo.
echo Your .env file has been created and configured.
echo.
echo Next steps:
echo 1. Review and verify your .env file
echo 2. Start PostgreSQL and Redis services
echo 3. Run database migrations: psql -U postgres -d goconnect -f pkg/db/migrations/001_initial_schema.sql
echo 4. Start the services: 
echo    - Auth Service: cd cmd/auth ^&^& go run main.go
echo    - Gateway: cd cmd/gateway ^&^& go run main.go
echo.
echo For detailed setup instructions, see: docs/OAUTH_JWT_OTP_SETUP.md
echo.

pause
