@echo off
echo ===============================================
echo  Kind Cluster Status
echo ===============================================
echo.

echo Checking Kind clusters...
kind get clusters
echo.

echo Checking kubectl context...
kubectl config current-context
echo.

echo Checking cluster info...
kubectl cluster-info
echo.

echo ===============================================
echo  GoConnect Namespace Status
echo ===============================================
echo.

echo Pods:
kubectl get pods -n goconnect
echo.

echo Services:
kubectl get services -n goconnect
echo.

echo Deployments:
kubectl get deployments -n goconnect
echo.

echo ConfigMaps and Secrets:
kubectl get configmaps,secrets -n goconnect
echo.

pause
