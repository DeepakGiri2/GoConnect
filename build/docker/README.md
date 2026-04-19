# Docker Build Files

> ⚠️ **dev-frontend and mock-oauth are DEV ONLY** - Never deploy to production!

This directory contains Docker configuration files for GoConnect.

## Files

- **Dockerfile.auth** - Multi-stage build for Auth Service
- **Dockerfile.gateway** - Multi-stage build for Gateway Service
- **docker-compose.dev.yml** - Development environment with all services
- **docker-compose.prod.yml** - Production environment with scaling & nginx
- **nginx.conf** - Nginx reverse proxy configuration for production

## Usage

### From Project Root

```bash
# Development
docker-compose -f build/docker/docker-compose.dev.yml up --build

# Production
docker-compose -f build/docker/docker-compose.prod.yml up -d --build
```

### Using Scripts

```bash
# Development
.\scripts\run-dev.bat

# Stop
.\scripts\stop-dev.bat
```

## Docker Context

Both Dockerfiles use the **project root** as the build context:
```yaml
build:
  context: ../..              # Project root (GoConnect/)
  dockerfile: build/docker/Dockerfile.auth
```

This allows access to:
- `/cmd` - Application entry points
- `/internal` - Service implementations
- `/pkg` - Shared libraries
- `/api` - API definitions
- `go.mod` & `go.sum` - Dependencies

## Port Mappings

### Development
- PostgreSQL: 5432
- Redis: 6379
- Auth Service: 50051 (gRPC)
- Gateway: 8080 (HTTP)

### Production
- Nginx: 80 (HTTP), 443 (HTTPS)
- Services are on internal network only

## Environment Variables

See `.env.example` in project root for all available variables.

Development uses defaults in `docker-compose.dev.yml`.
Production requires a `.env` file.

## Database Migrations

Migrations are automatically applied on container startup:
```yaml
volumes:
  - ../../pkg/db/migrations:/docker-entrypoint-initdb.d
```

## Health Checks

All services include health checks:
- PostgreSQL: `pg_isready`
- Redis: `redis-cli ping`
- Services wait for dependencies to be healthy before starting

## Scaling (Production Only)

Auth and Gateway services are configured to run 3 replicas:
```yaml
deploy:
  replicas: 3
```

## Build Optimization

The project includes a `.dockerignore` file to exclude:
- Documentation files
- Test files
- Build artifacts
- IDE configurations
- Git files

This significantly reduces build time and image size.
