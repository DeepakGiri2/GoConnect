@echo off
echo ========================================
echo GoConnect Setup Script
echo ========================================
echo.

echo Step 1: Checking Go installation...
go version
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed. Please install Go 1.23 or later.
    exit /b 1
)
echo.

echo Step 2: Installing dependencies...
cd %~dp0..
go mod download
if %errorlevel% neq 0 (
    echo ERROR: Failed to download dependencies.
    exit /b 1
)
echo.

echo Step 3: Setting up environment file...
if not exist .env (
    copy .env.example .env
    echo Created .env file. Please update it with your configuration.
) else (
    echo .env file already exists.
)
echo.

echo Step 4: Checking Docker installation...
docker --version
if %errorlevel% neq 0 (
    echo WARNING: Docker is not installed. You won't be able to use Docker Compose.
) else (
    echo Docker is installed.
)
echo.

echo Step 5: Checking protoc installation...
protoc --version
if %errorlevel% neq 0 (
    echo WARNING: protoc is not installed. You need it to generate gRPC code.
    echo Please install from: https://github.com/protocolbuffers/protobuf/releases
) else (
    echo protoc is installed.
    echo.
    echo Generating protobuf files...
    call scripts\generate-proto.bat
)
echo.

echo ========================================
echo Setup Complete!
echo ========================================
echo.
echo Next steps:
echo 1. Update .env file with your configuration
echo 2. Set up OAuth credentials (Google, Facebook, GitHub)
echo 3. Run: docker-compose -f docker\docker-compose.dev.yml up
echo.
echo For more information, see README.md
echo.
