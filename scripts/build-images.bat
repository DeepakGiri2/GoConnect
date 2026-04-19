@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  Building Docker Images for GoConnect
echo ===============================================
echo.

cd /d %~dp0\..

echo Step 1: Building Auth Service image...
docker build -t goconnect-auth:latest -f build\docker\Dockerfile.auth .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build auth service image
    exit /b 1
)

echo.
echo Step 2: Building Gateway image...
docker build -t goconnect-gateway:latest -f build\docker\Dockerfile.gateway .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build gateway image
    exit /b 1
)

echo.
echo Step 3: Loading images into Kind cluster...
kind load docker-image goconnect-auth:latest --name goconnect
if %errorlevel% neq 0 (
    echo ERROR: Failed to load auth image into Kind
    exit /b 1
)

kind load docker-image goconnect-gateway:latest --name goconnect
if %errorlevel% neq 0 (
    echo ERROR: Failed to load gateway image into Kind
    exit /b 1
)

echo.
echo ===============================================
echo  Images Built and Loaded Successfully!
echo ===============================================
echo.
echo Images:
echo - goconnect-auth:latest
echo - goconnect-gateway:latest
echo.
echo Next step:
echo - Deploy to cluster: scripts\deploy-to-kind.bat
echo.
pause
