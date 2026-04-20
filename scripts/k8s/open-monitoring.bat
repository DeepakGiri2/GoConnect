@echo off
echo ===============================================
echo  GoConnect - Monitoring Dashboards
echo ===============================================
echo.
echo Opening monitoring dashboards...
echo.

echo Opening Prometheus (Metrics)...
start http://localhost:9091

timeout /t 2 >nul

echo Opening Grafana (Dashboards)...
start http://localhost:3001

echo.
echo ===============================================
echo  Access Information:
echo ===============================================
echo.
echo Prometheus:  http://localhost:9091
echo   - View metrics and targets
echo   - Query metrics with PromQL
echo.
echo Grafana:     http://localhost:3001
echo   - Username: admin
echo   - Password: admin123
echo   - Create custom dashboards
echo.
echo Traefik Dashboard: http://localhost:8888/dashboard/
echo.
echo ===============================================
echo.
pause
