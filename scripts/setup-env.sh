#!/bin/bash

# GoConnect Environment Setup Script
# This script helps you set up your .env file

echo "============================================"
echo "GoConnect Environment Setup"
echo "============================================"
echo ""

if [ -f .env ]; then
    read -p "WARNING: .env file already exists! Do you want to overwrite it? (y/n): " OVERWRITE
    if [ "$OVERWRITE" != "y" ]; then
        echo "Setup cancelled."
        exit 0
    fi
fi

echo "Creating .env file from template..."
cp .env.example .env

echo ""
echo "============================================"
echo "Configuration Sections:"
echo "============================================"
echo ""
echo "1. Database Configuration"
echo "2. Redis Configuration"
echo "3. JWT Configuration"
echo "4. OTP Configuration"
echo "5. OAuth Configuration (Google, GitHub, Facebook)"
echo "6. Email Configuration"
echo ""

read -p "Do you want to configure Database settings? (y/n): " CONFIG_DB
if [ "$CONFIG_DB" = "y" ]; then
    echo ""
    echo "--- Database Configuration ---"
    read -p "Database Host [localhost]: " DB_HOST
    read -p "Database Port [5432]: " DB_PORT
    read -p "Database Name [goconnect]: " DB_NAME
    read -p "Database User [postgres]: " DB_USER
    read -sp "Database Password: " DB_PASS
    echo ""
    
    [ ! -z "$DB_HOST" ] && sed -i "s/DATABASE_HOST=.*/DATABASE_HOST=$DB_HOST/" .env
    [ ! -z "$DB_PORT" ] && sed -i "s/DATABASE_PORT=.*/DATABASE_PORT=$DB_PORT/" .env
    [ ! -z "$DB_NAME" ] && sed -i "s/DATABASE_NAME=.*/DATABASE_NAME=$DB_NAME/" .env
    [ ! -z "$DB_USER" ] && sed -i "s/DATABASE_USER=.*/DATABASE_USER=$DB_USER/" .env
    [ ! -z "$DB_PASS" ] && sed -i "s/DATABASE_PASSWORD=.*/DATABASE_PASSWORD=$DB_PASS/" .env
fi

echo ""
read -p "Do you want to configure Redis settings? (y/n): " CONFIG_REDIS
if [ "$CONFIG_REDIS" = "y" ]; then
    echo ""
    echo "--- Redis Configuration ---"
    read -p "Redis Host [localhost]: " REDIS_HOST
    read -p "Redis Port [6379]: " REDIS_PORT
    read -sp "Redis Password (leave empty if none): " REDIS_PASS
    echo ""
    
    [ ! -z "$REDIS_HOST" ] && sed -i "s/REDIS_HOST=.*/REDIS_HOST=$REDIS_HOST/" .env
    [ ! -z "$REDIS_PORT" ] && sed -i "s/REDIS_PORT=.*/REDIS_PORT=$REDIS_PORT/" .env
    [ ! -z "$REDIS_PASS" ] && sed -i "s/REDIS_PASSWORD=.*/REDIS_PASSWORD=$REDIS_PASS/" .env
fi

echo ""
echo "--- Security Configuration ---"
echo "Generating secure secrets..."

# Generate JWT Secret
JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')
sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
echo "JWT Secret generated and saved."

# Generate OTP Secret
OTP_SECRET=$(openssl rand -base64 32 | tr -d '\n')
sed -i "s|OTP_SECRET=.*|OTP_SECRET=$OTP_SECRET|" .env
echo "OTP Secret generated and saved."

echo ""
read -p "Do you want to configure OAuth providers? (y/n): " CONFIG_OAUTH
if [ "$CONFIG_OAUTH" = "y" ]; then
    echo ""
    echo "--- OAuth Configuration ---"
    echo ""
    echo "Google OAuth:"
    read -p "Google Client ID (leave empty to skip): " GOOGLE_ID
    if [ ! -z "$GOOGLE_ID" ]; then
        read -p "Google Client Secret: " GOOGLE_SECRET
        sed -i "s/GOOGLE_CLIENT_ID=.*/GOOGLE_CLIENT_ID=$GOOGLE_ID/" .env
        sed -i "s/GOOGLE_CLIENT_SECRET=.*/GOOGLE_CLIENT_SECRET=$GOOGLE_SECRET/" .env
    fi
    
    echo ""
    echo "GitHub OAuth:"
    read -p "GitHub Client ID (leave empty to skip): " GITHUB_ID
    if [ ! -z "$GITHUB_ID" ]; then
        read -p "GitHub Client Secret: " GITHUB_SECRET
        sed -i "s/GITHUB_CLIENT_ID=.*/GITHUB_CLIENT_ID=$GITHUB_ID/" .env
        sed -i "s/GITHUB_CLIENT_SECRET=.*/GITHUB_CLIENT_SECRET=$GITHUB_SECRET/" .env
    fi
    
    echo ""
    echo "Facebook OAuth:"
    read -p "Facebook Client ID (leave empty to skip): " FACEBOOK_ID
    if [ ! -z "$FACEBOOK_ID" ]; then
        read -p "Facebook Client Secret: " FACEBOOK_SECRET
        sed -i "s/FACEBOOK_CLIENT_ID=.*/FACEBOOK_CLIENT_ID=$FACEBOOK_ID/" .env
        sed -i "s/FACEBOOK_CLIENT_SECRET=.*/FACEBOOK_CLIENT_SECRET=$FACEBOOK_SECRET/" .env
    fi
fi

echo ""
read -p "Do you want to configure Email settings? (y/n): " CONFIG_EMAIL
if [ "$CONFIG_EMAIL" = "y" ]; then
    echo ""
    echo "--- Email Configuration ---"
    echo ""
    echo "Common SMTP Servers:"
    echo "1. Gmail: smtp.gmail.com:587"
    echo "2. SendGrid: smtp.sendgrid.net:587"
    echo "3. AWS SES: email-smtp.us-east-1.amazonaws.com:587"
    echo ""
    read -p "SMTP Host [smtp.gmail.com]: " SMTP_HOST
    read -p "SMTP Port [587]: " SMTP_PORT
    read -p "SMTP Username (email): " SMTP_USER
    read -sp "SMTP Password (App Password for Gmail): " SMTP_PASS
    echo ""
    read -p "From Email [noreply@goconnect.com]: " FROM_EMAIL
    read -p "From Name [GoConnect]: " FROM_NAME
    
    [ ! -z "$SMTP_HOST" ] && sed -i "s/SMTP_HOST=.*/SMTP_HOST=$SMTP_HOST/" .env
    [ ! -z "$SMTP_PORT" ] && sed -i "s/SMTP_PORT=.*/SMTP_PORT=$SMTP_PORT/" .env
    [ ! -z "$SMTP_USER" ] && sed -i "s/SMTP_USERNAME=.*/SMTP_USERNAME=$SMTP_USER/" .env
    [ ! -z "$SMTP_PASS" ] && sed -i "s/SMTP_PASSWORD=.*/SMTP_PASSWORD=$SMTP_PASS/" .env
    [ ! -z "$FROM_EMAIL" ] && sed -i "s/FROM_EMAIL=.*/FROM_EMAIL=$FROM_EMAIL/" .env
    [ ! -z "$FROM_NAME" ] && sed -i "s/FROM_NAME=.*/FROM_NAME=$FROM_NAME/" .env
fi

echo ""
echo "============================================"
echo "Setup Complete!"
echo "============================================"
echo ""
echo "Your .env file has been created and configured."
echo ""
echo "Next steps:"
echo "1. Review and verify your .env file"
echo "2. Start PostgreSQL and Redis services"
echo "3. Run database migrations: psql -U postgres -d goconnect -f pkg/db/migrations/001_initial_schema.sql"
echo "4. Start the services:"
echo "   - Auth Service: cd cmd/auth && go run main.go"
echo "   - Gateway: cd cmd/gateway && go run main.go"
echo ""
echo "For detailed setup instructions, see: docs/OAUTH_JWT_OTP_SETUP.md"
echo ""
