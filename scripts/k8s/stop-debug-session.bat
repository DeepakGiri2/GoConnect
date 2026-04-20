@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  Stopping Debug Session
echo ===============================================
echo.

cd /d %~dp0\..\..
echo Working from: %CD%
echo.

set /p SERVICE="Which service to restore? (gateway/auth/both): "

if /i "%SERVICE%"=="gateway" (
    call :restore_service gateway gateway-debug
) else if /i "%SERVICE%"=="auth" (
    call :restore_service auth-service auth-service-debug
) else if /i "%SERVICE%"=="both" (
    call :restore_service gateway gateway-debug
    call :restore_service auth-service auth-service-debug
) else (
    echo ERROR: Invalid service. Use 'gateway', 'auth', or 'both'
    exit /b 1
)

echo.
echo ===============================================
echo  Debug Session Stopped
echo ===============================================
echo.
echo Services restored to production mode
echo.
pause
exit /b 0

:restore_service
setlocal
set PROD=%~1
set DEBUG=%~2

echo.
echo Restoring %PROD%...
echo - Scaling debug deployment to 0...
kubectl scale deployment %DEBUG% --replicas=0 -n goconnect 2>nul

echo - Scaling production deployment to 2...
kubectl scale deployment %PROD% --replicas=2 -n goconnect

echo - Waiting for production pods to be ready...
kubectl wait --for=condition=ready pod -l app=%PROD% -n goconnect --timeout=120s 2>nul

echo - Done!
endlocal
exit /b 0
