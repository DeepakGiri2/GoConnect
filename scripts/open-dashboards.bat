@echo off
echo ===============================================
echo  Opening GoConnect Dashboards
echo ===============================================
echo.

echo Opening Traefik Dashboard...
start http://localhost:8888/dashboard/

timeout /t 2 >nul

echo Opening Gateway Health Check...
start http://localhost:8080/health

echo.
echo Dashboards opened in your browser!
echo.
echo If they don't load, make sure port-forwarding is active:
echo   kubectl port-forward -n traefik svc/traefik 8888:8080
echo   kubectl port-forward -n goconnect svc/gateway-service 8080:80
echo.
pause
