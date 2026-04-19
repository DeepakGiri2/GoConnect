@echo off
REM Generate a secure 32-character encryption key for TOTP secrets

echo Generating secure AES-256 encryption key...
echo.

REM Generate random 32 characters using PowerShell
powershell -Command "$key = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_}); Write-Host 'TOTP_ENCRYPTION_KEY='$key -ForegroundColor Green; Write-Host ''; Write-Host 'Copy this to your .env file' -ForegroundColor Yellow; Write-Host 'IMPORTANT: Store this key securely and never commit to version control!' -ForegroundColor Red"

echo.
pause
