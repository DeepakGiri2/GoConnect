@echo off
echo ===============================================
echo  Port-Forwarding Debug Ports
echo ===============================================
echo.
echo This will forward debug ports to localhost.
echo Keep this window open while debugging!
echo.
echo Press Ctrl+C to stop port-forwarding
echo.

set /p SERVICE="Which service? (gateway/auth/both): "

if /i "%SERVICE%"=="gateway" (
    call :forward_gateway
) else if /i "%SERVICE%"=="auth" (
    call :forward_auth
) else if /i "%SERVICE%"=="both" (
    start "Gateway Debug Port Forward" cmd /c "kubectl port-forward -n goconnect deployment/gateway-debug 40000:40000"
    timeout /t 2 >nul
    start "Auth Debug Port Forward" cmd /c "kubectl port-forward -n goconnect deployment/auth-service-debug 40001:40000"
    echo.
    echo ✅ Port forwarding started in separate windows
    echo    Gateway:  localhost:40000
    echo    Auth:     localhost:40001
    echo.
    echo Update GoLand configurations:
    echo    Gateway port: 40000
    echo    Auth port:    40001
    pause
    exit /b 0
) else (
    echo ERROR: Invalid choice
    exit /b 1
)
exit /b 0

:forward_gateway
echo Starting port-forward for Gateway debug...
echo.
echo ✅ Connect GoLand to localhost:40000
echo.
kubectl port-forward -n goconnect deployment/gateway-debug 40000:40000
exit /b 0

:forward_auth
echo Starting port-forward for Auth Service debug...
echo.
echo ✅ Connect GoLand to localhost:40000
echo.
kubectl port-forward -n goconnect deployment/auth-service-debug 40000:40000
exit /b 0
