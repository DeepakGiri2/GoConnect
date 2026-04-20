@echo off
echo ===============================================
echo  Starting Port Forwards for All Services
echo ===============================================
echo.

echo Killing existing port-forwards...
taskkill /F /IM kubectl.exe >nul 2>&1

timeout /t 2 >nul

echo Starting port forwards in separate windows...
echo.

start "Grafana (3001)" kubectl port-forward -n monitoring svc/grafana 3001:3000
echo   ✓ Grafana:     http://localhost:3001

timeout /t 1 >nul

start "Prometheus (9091)" kubectl port-forward -n monitoring svc/prometheus 9091:9090
echo   ✓ Prometheus:  http://localhost:9091

timeout /t 1 >nul

start "Traefik (8888)" kubectl port-forward -n traefik svc/traefik 8888:8080
echo   ✓ Traefik:     http://localhost:8888/dashboard/

timeout /t 1 >nul

start "Gateway (8080)" kubectl port-forward -n goconnect svc/gateway-service 8080:80
echo   ✓ Gateway:     http://localhost:8080

echo.
echo ===============================================
echo  All Port Forwards Started!
echo ===============================================
echo.
echo Access URLs:
echo   Grafana:     http://localhost:3001 (admin/admin123)
echo   Prometheus:  http://localhost:9091
echo   Traefik:     http://localhost:8888/dashboard/
echo   Gateway:     http://localhost:8080
echo.
echo Close individual windows to stop port-forwarding
echo.
pause
