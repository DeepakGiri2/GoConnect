@echo off
echo ===============================================
echo  GoConnect - Port Forward Services
echo ===============================================
echo.
echo Starting port forwarding for all services...
echo.
echo Access your services at:
echo   Gateway API:        http://localhost:8080
echo   Traefik Dashboard:  http://localhost:8888/dashboard/
echo   PostgreSQL:         localhost:5432
echo   Redis:              localhost:6379
echo.
echo Note: Traefik dashboard requires /dashboard/ path
echo.
echo Press Ctrl+C to stop all forwarding
echo.

start "Gateway" kubectl port-forward -n goconnect svc/gateway-service 8080:80
start "Traefik" kubectl port-forward -n traefik svc/traefik 8888:8080
start "PostgreSQL" kubectl port-forward -n goconnect svc/postgres-service 5432:5432
start "Redis" kubectl port-forward -n goconnect svc/redis-service 6379:6379

echo.
echo All services are now forwarded!
echo Close this window or press any key to stop...
pause >nul

taskkill /FI "WINDOWTITLE eq Gateway" /F
taskkill /FI "WINDOWTITLE eq Traefik" /F  
taskkill /FI "WINDOWTITLE eq PostgreSQL" /F
taskkill /FI "WINDOWTITLE eq Redis" /F
