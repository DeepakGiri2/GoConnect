#!/bin/bash

# Generate a secure 32-character encryption key for TOTP secrets

echo "Generating secure AES-256 encryption key..."
echo ""

# Generate random 32 characters
KEY=$(openssl rand -base64 24 | tr -dc 'A-Za-z0-9' | head -c 32)

echo -e "\033[0;32mTOTP_ENCRYPTION_KEY=$KEY\033[0m"
echo ""
echo -e "\033[0;33mCopy this to your .env file\033[0m"
echo -e "\033[0;31mIMPORTANT: Store this key securely and never commit to version control!\033[0m"
echo ""
