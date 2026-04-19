@echo off
REM Configuration validation script for Windows
REM Checks if all required environment variables are set

echo.
echo ================================
echo  GoConnect Config Validator
echo ================================
echo.

set "ENV_FILE=.env"
set "ERRORS=0"

REM Check if .env file exists
if not exist "%ENV_FILE%" (
    echo [ERROR] .env file not found!
    echo Please copy .env.example to .env and configure it.
    exit /b 1
)

echo Checking required configuration variables...
echo.

REM Required variables
call :check_var "DATABASE_HOST"
call :check_var "DATABASE_PORT"
call :check_var "DATABASE_NAME"
call :check_var "DATABASE_USER"
call :check_var "DATABASE_PASSWORD"
call :check_var "REDIS_HOST"
call :check_var "REDIS_PORT"
call :check_var "JWT_SECRET"
call :check_var "JWT_ACCESS_EXPIRY"
call :check_var "JWT_REFRESH_EXPIRY"
call :check_var "OTP_SECRET"
call :check_var "OTP_EXPIRY"
call :check_var "TOTP_ENCRYPTION_KEY"
call :check_var "AUTH_SERVICE_HOST"
call :check_var "AUTH_SERVICE_PORT"
call :check_var "SMTP_HOST"
call :check_var "SMTP_PORT"
call :check_var "SMTP_USERNAME"
call :check_var "SMTP_PASSWORD"
call :check_var "FROM_EMAIL"
call :check_var "FROM_NAME"

echo.
if %ERRORS% EQU 0 (
    echo [SUCCESS] All required variables are set!
    echo.
    echo Optional variables to review:
    call :check_optional "GOOGLE_CLIENT_ID" "Google OAuth"
    call :check_optional "FACEBOOK_CLIENT_ID" "Facebook OAuth"
    call :check_optional "GITHUB_CLIENT_ID" "GitHub OAuth"
    echo.
    echo Configuration validation passed!
) else (
    echo [FAILED] %ERRORS% required variable(s) missing or empty!
    echo Please update your .env file.
    exit /b 1
)

exit /b 0

:check_var
set "VAR_NAME=%~1"
findstr /B /C:"%VAR_NAME%=" "%ENV_FILE%" >nul 2>&1
if errorlevel 1 (
    echo [MISSING] %VAR_NAME%
    set /a ERRORS+=1
) else (
    for /f "tokens=2 delims==" %%a in ('findstr /B /C:"%VAR_NAME%=" "%ENV_FILE%"') do (
        set "VAR_VALUE=%%a"
    )
    if defined VAR_VALUE (
        if "!VAR_VALUE!"=="" (
            echo [EMPTY] %VAR_NAME%
            set /a ERRORS+=1
        ) else (
            echo [OK] %VAR_NAME%
        )
    ) else (
        echo [EMPTY] %VAR_NAME%
        set /a ERRORS+=1
    )
)
goto :eof

:check_optional
set "VAR_NAME=%~1"
set "SERVICE=%~2"
findstr /B /C:"%VAR_NAME%=" "%ENV_FILE%" >nul 2>&1
if errorlevel 1 (
    echo   - %SERVICE%: Not configured
) else (
    for /f "tokens=2 delims==" %%a in ('findstr /B /C:"%VAR_NAME%=" "%ENV_FILE%"') do (
        set "VAR_VALUE=%%a"
    )
    if defined VAR_VALUE (
        if not "!VAR_VALUE!"=="" (
            echo   - %SERVICE%: Configured
        ) else (
            echo   - %SERVICE%: Not configured
        )
    ) else (
        echo   - %SERVICE%: Not configured
    )
)
goto :eof
