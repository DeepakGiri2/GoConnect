@echo off
REM Remove Mock OAuth configuration from Windows hosts file
echo ============================================
echo Remove GoConnect Mock OAuth
echo ============================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator!
    echo Right-click and select "Run as administrator"
    pause
    exit /b 1
)

echo [1/2] Removing mock OAuth entries from hosts file...
findstr /V /C:"# GoConnect Mock OAuth" /C:"accounts.google.com" /C:"www.facebook.com" /C:"github.com" /C:"# End GoConnect Mock OAuth" C:\Windows\System32\drivers\etc\hosts > C:\Windows\System32\drivers\etc\hosts.tmp
move /Y C:\Windows\System32\drivers\etc\hosts.tmp C:\Windows\System32\drivers\etc\hosts >nul
echo Mock OAuth entries removed.
echo.

echo [2/2] Flushing DNS cache...
ipconfig /flushdns >nul
echo DNS cache flushed.
echo.

echo ============================================
echo Cleanup Complete!
echo ============================================
echo.
echo Real OAuth providers should now work normally.
echo.
pause
