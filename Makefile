.PHONY: help setup proto build test clean docker-dev docker-prod k8s-deploy dev-test dev-server mock-oauth

help:
	@echo "GoConnect - Available Commands"
	@echo "=============================="
	@echo "setup         - Install dependencies and setup environment"
	@echo "proto         - Generate protobuf code"
	@echo "build         - Build all services"
	@echo "test          - Run tests"
	@echo "clean         - Clean build artifacts"
	@echo "docker-dev    - Start development environment with Docker Compose"
	@echo "docker-stop   - Stop Docker Compose"
	@echo "docker-prod   - Start production environment with Docker Compose"
	@echo "k8s-deploy    - Deploy to Kubernetes"
	@echo "k8s-delete    - Delete from Kubernetes"
	@echo ""
	@echo "Development Testing (DEV ONLY - DO NOT USE IN PRODUCTION)"
	@echo "=========================================================="
	@echo "dev-test      - Start dev test server (serves temp/index.html)"
	@echo "mock-oauth    - Start mock OAuth server"
	@echo "dev-all       - Start all dev services (mock-oauth + gateway + auth)"

setup:
	@echo "Setting up GoConnect..."
	go mod download
	@if not exist .env copy .env.example .env
	@echo "Setup complete!"

proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/shared/proto/auth.proto
	@echo "Protobuf generation complete!"

build:
	@echo "Building services..."
	go build -o bin/auth-service.exe ./cmd/auth/main.go
	go build -o bin/gateway.exe ./cmd/gateway/main.go
	@echo "Build complete!"

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	if exist bin rmdir /s /q bin
	go clean

docker-dev:
	@echo "Starting development environment..."
	docker-compose -f build/docker/docker-compose.dev.yml up --build

docker-stop:
	@echo "Stopping Docker Compose..."
	docker-compose -f build/docker/docker-compose.dev.yml down

docker-prod:
	@echo "Starting production environment..."
	docker-compose -f build/docker/docker-compose.prod.yml up -d --build

k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deployments/k8s/namespace.yaml
	kubectl apply -f deployments/k8s/secrets.yaml
	kubectl apply -f deployments/k8s/configmap.yaml
	kubectl apply -f deployments/k8s/postgres.yaml
	kubectl apply -f deployments/k8s/redis.yaml
	kubectl apply -f deployments/k8s/auth-service.yaml
	kubectl apply -f deployments/k8s/gateway.yaml
	kubectl apply -f deployments/k8s/ingress.yaml
	@echo "Deployment complete!"

k8s-delete:
	@echo "Deleting from Kubernetes..."
	kubectl delete namespace goconnect
	@echo "Deletion complete!"

# ⚠️ DEVELOPMENT ONLY TARGETS - DO NOT USE IN PRODUCTION ⚠️
dev-test:
	@echo "=========================================="
	@echo "⚠️  DEV ONLY - NOT FOR PRODUCTION ⚠️"
	@echo "=========================================="
	@echo "Starting development test server..."
	@echo "Test interface: http://localhost:3000/index.html"
	go run cmd/dev-server/main.go

mock-oauth:
	@echo "=========================================="
	@echo "⚠️  DEV ONLY - NOT FOR PRODUCTION ⚠️"
	@echo "=========================================="
	@echo "Starting mock OAuth server..."
	docker compose -f build/docker/docker-compose.dev.yml up mock-oauth -d
	@echo "Mock servers running on ports 9000-9002"

dev-all:
	@echo "=========================================="
	@echo "⚠️  DEV ONLY - NOT FOR PRODUCTION ⚠️"
	@echo "=========================================="
	@echo "Starting all development services..."
	docker compose -f build/docker/docker-compose.dev.yml up -d
	@echo ""
	@echo "Services started!"
	@echo "- Gateway: http://localhost:8080"
	@echo "- Mock OAuth: ports 9000-9002"
	@echo ""
	@echo "Start test server with: make dev-test"
	@echo "or open: http://localhost:3000/index.html"
