@echo off
echo Generating protobuf files to proto_gen folder...
echo.

REM Check if protoc is installed
where protoc >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: protoc is not installed.
    echo Please install protoc from: https://github.com/protocolbuffers/protobuf/releases
    echo Or use: choco install protoc
    exit /b 1
)

REM Generate Go code from proto files to proto_gen folder
protoc --go_out=api/shared/proto_gen --go_opt=paths=source_relative ^
    --go-grpc_out=api/shared/proto_gen --go-grpc_opt=paths=source_relative ^
    api/shared/proto/auth.proto

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Protobuf files generated successfully to proto_gen!
    echo   - api/shared/proto_gen/shared/proto/auth.pb.go
    echo   - api/shared/proto_gen/shared/proto/auth_grpc.pb.go
    echo.
    echo Next steps:
    echo   1. Commit these files: git add api/shared/proto_gen/
    echo   2. For local dev, run: .\scripts\copy-proto-gen.bat
) else (
    echo.
    echo ERROR: Failed to generate protobuf files
    exit /b 1
)

echo.
pause
