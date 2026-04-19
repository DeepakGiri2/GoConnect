@echo off
REM Quick test script for Mock OAuth server
echo ============================================
echo Mock OAuth Server Test
echo ============================================
echo.

echo Testing Google Mock Server (Port 9000)...
curl -s http://localhost:9000/o/oauth2/v2/auth?redirect_uri=http://localhost:8080/callback^&state=test123
echo.
echo.

echo Testing Facebook Mock Server (Port 9001)...
curl -s http://localhost:9001/v12.0/dialog/oauth?redirect_uri=http://localhost:8080/callback^&state=test123
echo.
echo.

echo Testing GitHub Mock Server (Port 9002)...
curl -s http://localhost:9002/login/oauth/authorize?redirect_uri=http://localhost:8080/callback^&state=test123
echo.
echo.

echo ============================================
echo Test Complete
echo ============================================
echo.
echo If you see redirect URLs above, the servers are working!
echo If you see errors, check:
echo - Mock OAuth server is running
echo - Ports 9000-9002 are available
echo - Docker container is healthy
echo.
pause
