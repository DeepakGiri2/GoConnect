@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  Building Debug Docker Images for GoConnect
echo ===============================================
echo.

cd /d %~dp0\..\..
echo Building from: %CD%
echo.

echo Step 1: Building Auth Service DEBUG image...
docker build -t goconnect-auth:debug -f build\docker\Dockerfile.auth.debug .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build auth service debug image
    exit /b 1
)

echo.
echo Step 2: Building Gateway DEBUG image...
docker build -t goconnect-gateway:debug -f build\docker\Dockerfile.gateway.debug .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build gateway debug image
    exit /b 1
)

echo.
echo Step 3: Loading debug images into Kind cluster...
kind load docker-image goconnect-auth:debug --name goconnect
if %errorlevel% neq 0 (
    echo ERROR: Failed to load auth debug image into Kind
    exit /b 1
)

kind load docker-image goconnect-gateway:debug --name goconnect
if %errorlevel% neq 0 (
    echo ERROR: Failed to load gateway debug image into Kind
    exit /b 1
)

echo.
echo ===============================================
echo  Debug Images Built and Loaded Successfully!
echo ===============================================
echo.
echo Images:
echo - goconnect-auth:debug
echo - goconnect-gateway:debug
echo.
echo Next steps:
echo 1. Deploy debug pods: kubectl apply -f deployments\k8s\gateway-debug.yaml
echo 2. Deploy debug pods: kubectl apply -f deployments\k8s\auth-service-debug.yaml
echo 3. Connect VS Code debugger to localhost:30040 (gateway) or localhost:30041 (auth)
echo.
echo See deployments\k8s\DEBUG-SETUP.md for full guide
echo.
pause
