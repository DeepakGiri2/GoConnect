@echo off
echo ===============================================
echo  Stopping All Kubernetes Processes
echo ===============================================
echo.

echo [1/4] Killing all port-forward processes...
taskkill /F /IM kubectl.exe >nul 2>&1
if %errorlevel% EQU 0 (
    echo   ✓ Port-forwards stopped
) else (
    echo   ℹ No port-forwards running
)

timeout /t 2 >nul

echo.
echo [2/4] Stopping Kind cluster...
kind delete cluster --name goconnect
if %errorlevel% EQU 0 (
    echo   ✓ Kind cluster deleted
) else (
    echo   ✗ Failed to delete cluster
)

timeout /t 2 >nul

echo.
echo [3/4] Cleaning up Docker containers...
docker ps -a --filter "name=goconnect" -q | ForEach-Object { docker rm -f $_ } 2>nul
echo   ✓ Docker containers cleaned

timeout /t 1 >nul

echo.
echo [4/4] Cleaning up Docker networks...
docker network prune -f >nul 2>&1
echo   ✓ Docker networks cleaned

echo.
echo ===============================================
echo  All Kubernetes Processes Stopped!
echo ===============================================
echo.
echo What was stopped:
echo   ✓ All kubectl port-forwards
echo   ✓ Kind cluster "goconnect"
echo   ✓ All Docker containers
echo   ✓ Unused Docker networks
echo.
echo To restart the cluster, run:
echo   .\scripts\k8s\setup-kind-cluster.bat
echo   .\scripts\k8s\build-images.bat
echo   .\scripts\k8s\deploy-to-kind.bat
echo.
pause
