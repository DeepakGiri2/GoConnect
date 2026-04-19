@echo off
echo Stopping GoConnect...
cd %~dp0..
docker-compose -f build\docker\docker-compose.dev.yml down
