@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  Starting Debug Session
echo ===============================================
echo.

set /p SERVICE="Which service to debug? (gateway/auth): "

if /i "%SERVICE%"=="gateway" (
    set DEPLOYMENT=gateway
    set DEBUG_DEPLOYMENT=gateway-debug
    set DEBUG_PORT=30040
    set APP_PORT=30081
) else if /i "%SERVICE%"=="auth" (
    set DEPLOYMENT=auth-service
    set DEBUG_DEPLOYMENT=auth-service-debug
    set DEBUG_PORT=30041
    set APP_PORT=30051
) else (
    echo ERROR: Invalid service. Use 'gateway' or 'auth'
    exit /b 1
)

echo.
echo Step 1: Scaling production %DEPLOYMENT% to 0...
kubectl scale deployment %DEPLOYMENT% --replicas=0 -n goconnect

echo.
echo Step 2: Ensuring debug deployment has only 1 replica...
kubectl scale deployment %DEBUG_DEPLOYMENT% --replicas=1 -n goconnect 2>nul
if %errorlevel% neq 0 (
    echo Debug deployment doesn't exist yet. Deploying...
    cd /d %~dp0\..\..
    if /i "%SERVICE%"=="gateway" (
        kubectl apply -f deployments/k8s/gateway-debug.yaml
    ) else (
        kubectl apply -f deployments/k8s/auth-service-debug.yaml
    )
)

echo.
echo Step 3: Waiting for debug pod to be ready...
kubectl wait --for=condition=ready pod -l app=%DEBUG_DEPLOYMENT% -n goconnect --timeout=120s

echo.
echo Step 4: Getting debug pod name...
for /f "tokens=*" %%i in ('kubectl get pods -n goconnect -l app^=%DEBUG_DEPLOYMENT% -o jsonpath^="{.items[0].metadata.name}"') do set POD_NAME=%%i
echo Debug Pod: %POD_NAME%

echo.
echo ===============================================
echo  Debug Session Ready!
echo ===============================================
echo.
echo Service:      %SERVICE%
echo Debug Pod:    %POD_NAME%
echo Debug Port:   localhost:%DEBUG_PORT%
echo App Port:     localhost:%APP_PORT%
echo.
echo GoLand Setup:
echo 1. Open Run/Debug Configurations
echo 2. Select "Debug %SERVICE% Remote K8s"
echo 3. Click Debug (Shift+F9)
echo 4. Wait for "Connected to localhost:%DEBUG_PORT%"
echo 5. Set breakpoints and make requests!
echo.
echo Test Request:
if /i "%SERVICE%"=="gateway" (
    echo   curl http://localhost:%APP_PORT%/health
) else (
    echo   Use gateway to trigger auth service calls
)
echo.
echo To view logs:
echo   kubectl logs -f %POD_NAME% -n goconnect
echo.
echo When done debugging, run: .\scripts\k8s\stop-debug-session.bat
echo.
pause
