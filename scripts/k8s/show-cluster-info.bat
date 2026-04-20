@echo off
echo ===============================================
echo  GoConnect Cluster Information
echo ===============================================
echo.

echo === NAMESPACES ===
kubectl get namespaces
echo.

echo === ALL PODS ===
kubectl get pods -A
echo.

echo === GOCONNECT PODS (with container names) ===
for /f "tokens=*" %%p in ('kubectl get pods -n goconnect --no-headers -o custom-columns^=":metadata.name"') do (
    echo.
    echo Pod: %%p
    kubectl get pod %%p -n goconnect -o jsonpath='{range .spec.containers[*]}  Container: {.name}{"\n"}{end}'
)
echo.

echo === SERVICES ===
echo.
echo --- GoConnect Services ---
kubectl get svc -n goconnect
echo.
echo --- Traefik Services ---
kubectl get svc -n traefik
echo.
echo --- Monitoring Services ---
kubectl get svc -n monitoring
echo.

echo === INGRESS ===
kubectl get ingress -n goconnect
echo.

echo === QUICK POD NAMES ===
echo.
echo Gateway Pods:
kubectl get pods -n goconnect -l app=gateway --no-headers -o custom-columns=":metadata.name"
echo.
echo Auth Service Pods:
kubectl get pods -n goconnect -l app=auth-service --no-headers -o custom-columns=":metadata.name"
echo.
echo PostgreSQL Pod:
kubectl get pods -n goconnect -l app=postgres --no-headers -o custom-columns=":metadata.name"
echo.
echo Redis Pod:
kubectl get pods -n goconnect -l app=redis --no-headers -o custom-columns=":metadata.name"
echo.

echo === ACCESS URLS ===
echo.
echo Gateway API:  http://localhost:8080
echo Traefik Dash: http://localhost:8888
echo Grafana:      http://localhost:3000 (admin/admin123)
echo Prometheus:   http://localhost:9090
echo PostgreSQL:   localhost:5432
echo Redis:        localhost:6379
echo.

echo === USEFUL COMMANDS ===
echo.
echo View Gateway logs:
echo   kubectl logs -f deployment/gateway -n goconnect
echo.
echo View Auth Service logs:
echo   kubectl logs -f deployment/auth-service -n goconnect
echo.
echo Shell into Postgres:
echo   kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect
echo.
echo Shell into Redis:
echo   kubectl exec -it redis-0 -n goconnect -- redis-cli
echo.
echo Get pod events:
echo   kubectl get events -n goconnect --sort-by=.metadata.creationTimestamp
echo.
pause
