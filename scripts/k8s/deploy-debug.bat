@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  Deploying Debug Pods to Kubernetes
echo ===============================================
echo.

cd /d %~dp0\..\..
echo Deploying from: %CD%
echo.

REM Check if kubectl context is set to kind-goconnect
kubectl config current-context | findstr "kind-goconnect" >nul
if %errorlevel% neq 0 (
    echo Setting kubectl context to kind-goconnect...
    kubectl config use-context kind-goconnect
)

echo Step 1: Deploying Gateway Debug Pod...
kubectl apply -f deployments\k8s\gateway-debug.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to deploy gateway debug pod
    exit /b 1
)

echo.
echo Step 2: Deploying Auth Service Debug Pod...
kubectl apply -f deployments\k8s\auth-service-debug.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to deploy auth service debug pod
    exit /b 1
)

echo.
echo Step 3: Waiting for debug pods to be ready...
timeout /t 5 >nul
kubectl wait --for=condition=ready pod -l app=gateway-debug -n goconnect --timeout=120s
kubectl wait --for=condition=ready pod -l app=auth-service-debug -n goconnect --timeout=120s

echo.
echo ===============================================
echo  Debug Pods Deployed Successfully!
echo ===============================================
echo.

echo Debug Pod Status:
kubectl get pods -n goconnect -l app=gateway-debug
kubectl get pods -n goconnect -l app=auth-service-debug

echo.
echo ===============================================
echo  Debug Connection Information:
echo ===============================================
echo.
echo Gateway Debug:
echo   - Application: http://localhost:30081
echo   - Debug Port:  localhost:30040
echo   - VS Code:     "Debug Gateway (Remote K8s)"
echo.
echo Auth Service Debug:
echo   - GRPC Port:   localhost:30051
echo   - Debug Port:  localhost:30041
echo   - VS Code:     "Debug Auth Service (Remote K8s)"
echo.
echo To connect:
echo 1. Set breakpoints in VS Code
echo 2. Press F5 and select the debug configuration
echo 3. Make requests to trigger breakpoints
echo.
echo To view logs:
echo   kubectl logs -f deployment/gateway-debug -n goconnect
echo   kubectl logs -f deployment/auth-service-debug -n goconnect
echo.
echo See deployments\k8s\DEBUG-SETUP.md for full guide
echo.
pause
