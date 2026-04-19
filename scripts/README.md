# GoConnect Scripts

This directory contains automation scripts for managing the GoConnect project.

## Kind (Kubernetes in Docker) Scripts

### Installation

**`install-kind.bat`** - Install Kind CLI tool
- Downloads Kind v0.25.0 to `%USERPROFILE%\bin`
- Provides instructions to add to PATH
- Run this first if Kind is not installed

### Cluster Management

**`kind-quickstart.bat`** ⭐ **RECOMMENDED** - Complete setup in one command
- Creates Kind cluster
- Builds Docker images
- Deploys all services
- **Use this for first-time setup!**

**`setup-kind-cluster.bat`** - Create Kind cluster only
- Creates cluster named "goconnect"
- Configures port mappings for services
- Waits for cluster to be ready

**`delete-kind-cluster.bat`** - Delete Kind cluster
- Removes entire cluster and all data
- Prompts for confirmation

**`kind-status.bat`** - Check cluster and service status
- Shows cluster info
- Lists pods, services, deployments
- Useful for debugging

### Build & Deploy

**`build-images.bat`** - Build Docker images
- Builds `goconnect-auth:latest`
- Builds `goconnect-gateway:latest`
- Loads images into Kind cluster

**`deploy-to-kind.bat`** - Deploy services to Kind
- Creates namespace and secrets
- Deploys PostgreSQL and Redis
- Runs database migrations
- Deploys Auth Service and Gateway
- Waits for all services to be ready

## Database Scripts

**`init-db.bat`** - Initialize database
- Sets up PostgreSQL database
- Runs migrations

**`generate-encryption-key.bat`** - Generate encryption key
- Creates secure encryption key for data protection

## Typical Workflows

### First Time Setup

```powershell
# 1. Install Kind (if not already installed)
scripts\install-kind.bat

# 2. Add %USERPROFILE%\bin to PATH and restart terminal

# 3. Run complete setup
scripts\kind-quickstart.bat
```

### Daily Development

```powershell
# Start cluster (if not running)
scripts\setup-kind-cluster.bat

# Check status
scripts\kind-status.bat

# Make code changes, then rebuild
scripts\build-images.bat

# Restart deployments
kubectl rollout restart deployment/gateway -n goconnect
kubectl rollout restart deployment/auth-service -n goconnect
```

### Reset Everything

```powershell
# Delete and recreate
scripts\delete-kind-cluster.bat
scripts\kind-quickstart.bat
```

## Quick Reference

| Task | Command |
|------|---------|
| First setup | `scripts\kind-quickstart.bat` |
| Check status | `scripts\kind-status.bat` |
| View logs | `kubectl logs -f deployment/gateway -n goconnect` |
| Rebuild after changes | `scripts\build-images.bat` |
| Reset cluster | `scripts\delete-kind-cluster.bat` |

## See Also

- `docs\KIND_SETUP.md` - Detailed Kind setup guide
- `docs\DEPLOYMENT.md` - General deployment documentation
- `deployments\k8s\` - Kubernetes manifests
