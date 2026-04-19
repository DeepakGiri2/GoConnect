@echo off
echo Installing Kind for Windows...

set KIND_VERSION=v0.25.0
set INSTALL_DIR=%USERPROFILE%\bin

if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

echo Downloading Kind %KIND_VERSION%...
curl -Lo "%INSTALL_DIR%\kind.exe" "https://kind.sigs.k8s.io/dl/%KIND_VERSION%/kind-windows-amd64"

echo.
echo Kind installed to: %INSTALL_DIR%\kind.exe
echo.
echo Please add %INSTALL_DIR% to your PATH:
echo 1. Press Win+X and select "System"
echo 2. Click "Advanced system settings"
echo 3. Click "Environment Variables"
echo 4. Under "User variables", select "Path" and click "Edit"
echo 5. Click "New" and add: %INSTALL_DIR%
echo 6. Click OK on all dialogs
echo 7. Restart your terminal/PowerShell
echo.
echo After adding to PATH, verify with: kind version
pause
