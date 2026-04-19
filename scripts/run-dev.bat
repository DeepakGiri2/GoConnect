@echo off
echo Starting GoConnect in development mode...
cd %~dp0..
docker-compose -f build\docker\docker-compose.dev.yml up --build
