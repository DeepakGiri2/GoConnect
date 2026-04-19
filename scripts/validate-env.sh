#!/bin/bash
# Configuration validation script for Linux/Mac
# Checks if all required environment variables are set

set -e

ENV_FILE=".env"
ERRORS=0

echo ""
echo "================================"
echo " GoConnect Config Validator"
echo "================================"
echo ""

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "[ERROR] .env file not found!"
    echo "Please copy .env.example to .env and configure it."
    exit 1
fi

echo "Checking required configuration variables..."
echo ""

# Function to check if a variable is set and not empty
check_var() {
    local var_name=$1
    local var_value=$(grep "^${var_name}=" "$ENV_FILE" 2>/dev/null | cut -d '=' -f 2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    if [ -z "$var_value" ]; then
        echo "[MISSING/EMPTY] $var_name"
        ERRORS=$((ERRORS + 1))
    else
        echo "[OK] $var_name"
    fi
}

# Function to check optional variables
check_optional() {
    local var_name=$1
    local service=$2
    local var_value=$(grep "^${var_name}=" "$ENV_FILE" 2>/dev/null | cut -d '=' -f 2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    if [ -z "$var_value" ]; then
        echo "  - $service: Not configured"
    else
        echo "  - $service: Configured"
    fi
}

# Required variables
check_var "DATABASE_HOST"
check_var "DATABASE_PORT"
check_var "DATABASE_NAME"
check_var "DATABASE_USER"
check_var "DATABASE_PASSWORD"
check_var "REDIS_HOST"
check_var "REDIS_PORT"
check_var "JWT_SECRET"
check_var "JWT_ACCESS_EXPIRY"
check_var "JWT_REFRESH_EXPIRY"
check_var "OTP_SECRET"
check_var "OTP_EXPIRY"
check_var "TOTP_ENCRYPTION_KEY"
check_var "AUTH_SERVICE_HOST"
check_var "AUTH_SERVICE_PORT"
check_var "SMTP_HOST"
check_var "SMTP_PORT"
check_var "SMTP_USERNAME"
check_var "SMTP_PASSWORD"
check_var "FROM_EMAIL"
check_var "FROM_NAME"

echo ""

if [ $ERRORS -eq 0 ]; then
    echo "[SUCCESS] All required variables are set!"
    echo ""
    echo "Optional variables to review:"
    check_optional "GOOGLE_CLIENT_ID" "Google OAuth"
    check_optional "FACEBOOK_CLIENT_ID" "Facebook OAuth"
    check_optional "GITHUB_CLIENT_ID" "GitHub OAuth"
    echo ""
    echo "Configuration validation passed!"
    exit 0
else
    echo "[FAILED] $ERRORS required variable(s) missing or empty!"
    echo "Please update your .env file."
    exit 1
fi
