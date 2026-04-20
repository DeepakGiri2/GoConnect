@echo off
echo ===============================================
echo  Deleting Kind Cluster: goconnect
echo ===============================================
echo.
echo This will delete the entire Kind cluster and all data.
echo.
set /p confirm="Are you sure you want to continue? (yes/no): "

if /i not "%confirm%"=="yes" (
    echo Operation cancelled.
    exit /b 0
)

echo.
echo Deleting cluster...
kind delete cluster --name goconnect

if %errorlevel% equ 0 (
    echo.
    echo Cluster deleted successfully!
) else (
    echo.
    echo ERROR: Failed to delete cluster
    exit /b 1
)

echo.
pause
