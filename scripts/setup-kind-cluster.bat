@echo off
setlocal enabledelayedexpansion

echo ===============================================
echo  GoConnect - Kind Cluster Setup
echo ===============================================
echo.

REM Check if kind is installed
where kind >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Kind is not installed or not in PATH
    echo Please run install-kind.bat first
    exit /b 1
)

REM Check if Docker is running
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker is not running
    echo Please start Docker Desktop and try again
    exit /b 1
)

echo Step 1: Checking for existing cluster...
kind get clusters 2>nul | findstr /C:"goconnect" >nul
if %errorlevel% equ 0 (
    echo Found existing goconnect cluster. Deleting...
    kind delete cluster --name goconnect
    timeout /t 3 >nul
)

echo.
echo Step 2: Creating Kind cluster with configuration...
kind create cluster --config=deployments\k8s\kind-config.yaml --wait 300s
if %errorlevel% neq 0 (
    echo ERROR: Failed to create Kind cluster
    exit /b 1
)

echo.
echo Step 3: Waiting for cluster to be ready...
kubectl wait --for=condition=Ready nodes --all --timeout=300s
if %errorlevel% neq 0 (
    echo ERROR: Cluster nodes not ready
    exit /b 1
)

echo.
echo Step 4: Verifying cluster...
kubectl cluster-info
kubectl get nodes

echo.
echo ===============================================
echo  Kind Cluster Created Successfully!
echo ===============================================
echo.
echo Cluster Name: goconnect
echo Context: kind-goconnect
echo.
echo Next steps:
echo 1. Build Docker images: scripts\build-images.bat
echo 2. Deploy to cluster: scripts\deploy-to-kind.bat
echo.
pause
