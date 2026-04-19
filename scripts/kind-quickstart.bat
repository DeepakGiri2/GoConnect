@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  GoConnect - Kind Quickstart
echo  Complete setup in one command
echo ===============================================
echo.
echo This will:
echo 1. Create a Kind cluster named 'goconnect'
echo 2. Build Docker images
echo 3. Deploy all services
echo.
echo Estimated time: 5-10 minutes
echo.
set /p confirm="Continue? (yes/no): "

if /i not "%confirm%"=="yes" (
    echo Operation cancelled.
    exit /b 0
)

cd /d %~dp0

echo.
echo ===============================================
echo Step 1/3: Creating Kind Cluster
echo ===============================================
call setup-kind-cluster.bat
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Failed to create cluster
    exit /b 1
)

echo.
echo ===============================================
echo Step 2/3: Building Docker Images
echo ===============================================
call build-images.bat
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Failed to build images
    exit /b 1
)

echo.
echo ===============================================
echo Step 3/3: Deploying to Kind
echo ===============================================
call deploy-to-kind.bat
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Failed to deploy
    exit /b 1
)

echo.
echo ===============================================
echo  SUCCESS! GoConnect is now running!
echo ===============================================
echo.
echo Access your services at:
echo   API Gateway:  http://localhost:8080
echo   PostgreSQL:   localhost:5432
echo   Redis:        localhost:6379
echo.
echo Quick commands:
echo   Status:       scripts\kind-status.bat
echo   Logs:         kubectl logs -f deployment/gateway -n goconnect
echo   Delete:       scripts\delete-kind-cluster.bat
echo.
echo Test the API:
echo   curl http://localhost:8080/health
echo.
pause
