@echo off
echo Initializing database...
cd %~dp0..

set /p POSTGRES_PASSWORD=Enter PostgreSQL password: 

psql -h localhost -U postgres -d goconnect -f shared\db\migrations\001_initial_schema.sql

if %errorlevel% neq 0 (
    echo ERROR: Failed to initialize database.
    exit /b 1
)

echo Database initialized successfully!
