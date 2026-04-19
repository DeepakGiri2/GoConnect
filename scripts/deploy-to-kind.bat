@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  Deploying GoConnect to Kind Cluster
echo ===============================================
echo.

cd /d %~dp0\..

REM Check if kubectl context is set to kind-goconnect
kubectl config current-context | findstr "kind-goconnect" >nul
if %errorlevel% neq 0 (
    echo Setting kubectl context to kind-goconnect...
    kubectl config use-context kind-goconnect
)

echo Step 1: Creating namespace...
kubectl apply -f deployments\k8s\namespace.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to create namespace
    exit /b 1
)

echo.
echo Step 2: Creating secrets and configmap...
kubectl apply -f deployments\k8s\secrets.yaml
kubectl apply -f deployments\k8s\configmap.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to create secrets/configmap
    exit /b 1
)

echo.
echo Step 3: Deploying Traefik Load Balancer...
kubectl apply -f deployments\k8s\traefik.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to deploy Traefik
    exit /b 1
)

echo.
echo Waiting for Traefik to be ready...
timeout /t 5 >nul
kubectl wait --for=condition=ready pod -l app=traefik -n traefik --timeout=300s

echo.
echo Step 4: Deploying PostgreSQL...
kubectl apply -f deployments\k8s\postgres.yaml
kubectl apply -f deployments\k8s\postgres-kind.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to deploy PostgreSQL
    exit /b 1
)

echo.
echo Step 5: Deploying Redis...
kubectl apply -f deployments\k8s\redis.yaml
kubectl apply -f deployments\k8s\redis-kind.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to deploy Redis
    exit /b 1
)

echo.
echo Step 6: Waiting for database services to be ready...
timeout /t 10 >nul
kubectl wait --for=condition=ready pod -l app=postgres -n goconnect --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n goconnect --timeout=300s

echo.
echo Step 7: Running database migrations...
echo Copying migration file to postgres pod...
for /f "tokens=*" %%i in ('kubectl get pods -n goconnect -l app^=postgres -o jsonpath^="{.items[0].metadata.name}"') do set POSTGRES_POD=%%i
kubectl cp pkg\db\migrations\001_initial_schema.sql goconnect/%POSTGRES_POD%:/tmp/migration.sql
kubectl exec -n goconnect %POSTGRES_POD% -- psql -U postgres -d goconnect -f /tmp/migration.sql

echo.
echo Step 8: Deploying Auth Service...
kubectl apply -f deployments\k8s\auth-service.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to deploy Auth Service
    exit /b 1
)

echo.
echo Step 9: Deploying Gateway...
kubectl apply -f deployments\k8s\gateway.yaml
kubectl apply -f deployments\k8s\gateway-kind.yaml
if %errorlevel% neq 0 (
    echo ERROR: Failed to deploy Gateway
    exit /b 1
)

echo.
echo Step 10: Waiting for services to be ready...
timeout /t 5 >nul
kubectl wait --for=condition=ready pod -l app=auth-service -n goconnect --timeout=300s
kubectl wait --for=condition=ready pod -l app=gateway -n goconnect --timeout=300s

echo.
echo ===============================================
echo  Deployment Complete!
echo ===============================================
echo.
echo Checking pod status...
kubectl get pods -n goconnect
echo.
echo Checking services...
kubectl get services -n goconnect
echo.
echo ===============================================
echo  Access Information:
echo ===============================================
echo.
echo Traefik LB:   http://localhost:8080
echo Traefik Dash: http://localhost:30888
echo PostgreSQL:   localhost:5432
echo Redis:        localhost:6379
echo.
echo To view logs:
echo   kubectl logs -f deployment/gateway -n goconnect
echo   kubectl logs -f deployment/auth-service -n goconnect
echo.
echo To scale services:
echo   kubectl scale deployment gateway --replicas=5 -n goconnect
echo.
pause
