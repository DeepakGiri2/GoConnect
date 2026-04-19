# Kind (Kubernetes in Docker) Setup Guide

This guide will help you set up GoConnect on a local Kind cluster for development and testing.

## Prerequisites

- ✅ Docker Desktop installed and running
- ✅ kubectl installed
- ⚠️ Kind CLI tool (install using the steps below)

## Quick Start

### 1. Install Kind

Run the installation script:

```powershell
scripts\install-kind.bat
```

This will:
- Download Kind v0.25.0 to `%USERPROFILE%\bin\kind.exe`
- Display instructions to add it to your PATH

**After installation:**
1. Add `%USERPROFILE%\bin` to your PATH environment variable
2. Restart your terminal/PowerShell
3. Verify with: `kind version`

### 2. Create Kind Cluster

```powershell
scripts\setup-kind-cluster.bat
```

This creates a cluster named `goconnect` with:
- Port mapping for API Gateway (8080)
- Port mapping for PostgreSQL (5432)
- Port mapping for Redis (6379)

### 3. Build Docker Images

```powershell
scripts\build-images.bat
```

This builds and loads:
- `goconnect-auth:latest` - Authentication service
- `goconnect-gateway:latest` - API Gateway

### 4. Deploy to Kind

```powershell
scripts\deploy-to-kind.bat
```

This deploys:
- Namespace: `goconnect`
- PostgreSQL database
- Redis cache
- Auth Service (gRPC)
- API Gateway (HTTP)

## Access Your Services

After deployment, your services are available at:

- **API Gateway**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

### Test the API

```powershell
curl http://localhost:8080/health
```

## Useful Commands

### View Cluster Status

```powershell
scripts\kind-status.bat
```

### View Logs

```bash
# Gateway logs
kubectl logs -f deployment/gateway -n goconnect

# Auth service logs
kubectl logs -f deployment/auth-service -n goconnect

# All pods
kubectl logs -f -l app=gateway -n goconnect
```

### Scale Services

```bash
# Scale gateway to 5 replicas
kubectl scale deployment gateway --replicas=5 -n goconnect

# Scale auth service
kubectl scale deployment auth-service --replicas=5 -n goconnect
```

### Execute Commands in Pods

```bash
# Access PostgreSQL
kubectl exec -it deployment/postgres -n goconnect -- psql -U postgres -d goconnect

# Access Redis
kubectl exec -it deployment/redis -n goconnect -- redis-cli
```

### Port Forwarding (Alternative Access)

```bash
# Forward gateway
kubectl port-forward service/gateway-service 8080:80 -n goconnect

# Forward postgres
kubectl port-forward service/postgres-service 5432:5432 -n goconnect
```

### Apply Configuration Changes

```bash
# Update secrets
kubectl apply -f deployments\k8s\secrets.yaml

# Update configmap
kubectl apply -f deployments\k8s\configmap.yaml

# Restart deployments to pick up changes
kubectl rollout restart deployment/gateway -n goconnect
kubectl rollout restart deployment/auth-service -n goconnect
```

## Configuration

### Update Secrets

Edit `deployments\k8s\secrets.yaml` and update:

```yaml
stringData:
  database-password: "your_secure_password"
  jwt-secret: "your_jwt_secret_key"
  otp-secret: "your_otp_secret_key"
  google-client-id: "your_google_oauth_id"
  google-client-secret: "your_google_oauth_secret"
  # ... etc
```

Then apply:
```powershell
kubectl apply -f deployments\k8s\secrets.yaml
kubectl rollout restart deployment/gateway deployment/auth-service -n goconnect
```

### Update ConfigMap

Edit `deployments\k8s\configmap.yaml` and apply:

```powershell
kubectl apply -f deployments\k8s\configmap.yaml
kubectl rollout restart deployment/gateway deployment/auth-service -n goconnect
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n goconnect

# Describe pod for events
kubectl describe pod <pod-name> -n goconnect

# Check logs
kubectl logs <pod-name> -n goconnect
```

### Database Connection Issues

```bash
# Check if postgres is ready
kubectl get pods -l app=postgres -n goconnect

# Test connection
kubectl exec -it deployment/postgres -n goconnect -- psql -U postgres -c "SELECT 1"
```

### Image Pull Issues

```bash
# Verify images are loaded in Kind
docker exec -it goconnect-control-plane crictl images | findstr goconnect

# Rebuild and reload
scripts\build-images.bat
```

### Reset Everything

```powershell
# Delete cluster
scripts\delete-kind-cluster.bat

# Recreate from scratch
scripts\setup-kind-cluster.bat
scripts\build-images.bat
scripts\deploy-to-kind.bat
```

## Kind Cluster Configuration

The cluster is configured with port mappings in `deployments\k8s\kind-config.yaml`:

```yaml
extraPortMappings:
  - containerPort: 30080  # Maps to Gateway NodePort
    hostPort: 8080
  - containerPort: 30432  # Maps to PostgreSQL NodePort
    hostPort: 5432
  - containerPort: 30379  # Maps to Redis NodePort
    hostPort: 6379
```

## Development Workflow

1. **Make code changes** to your services
2. **Rebuild images**: `scripts\build-images.bat`
3. **Restart deployments**:
   ```bash
   kubectl rollout restart deployment/gateway -n goconnect
   kubectl rollout restart deployment/auth-service -n goconnect
   ```
4. **Watch rollout**: `kubectl rollout status deployment/gateway -n goconnect`

## Cleanup

### Delete the Cluster

```powershell
scripts\delete-kind-cluster.bat
```

This removes the entire cluster and all data.

### Delete Only Deployments

```bash
# Delete all resources in namespace
kubectl delete namespace goconnect

# Recreate namespace
kubectl apply -f deployments\k8s\namespace.yaml
```

## Advanced

### Manual Database Migration

```bash
# Get postgres pod name
kubectl get pods -n goconnect -l app=postgres

# Copy migration file
kubectl cp pkg\db\migrations\001_initial_schema.sql goconnect/<postgres-pod>:/tmp/

# Execute migration
kubectl exec -n goconnect <postgres-pod> -- psql -U postgres -d goconnect -f /tmp/001_initial_schema.sql
```

### Install Ingress Controller (Optional)

```bash
# Install nginx ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Wait for it to be ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

# Apply your ingress
kubectl apply -f deployments\k8s\ingress.yaml
```

## Next Steps

- Configure OAuth providers (see DEPLOYMENT.md)
- Set up monitoring and logging
- Configure proper secrets for production-like testing
- Test API endpoints
- Run integration tests

## Resources

- [Kind Documentation](https://kind.sigs.k8s.io/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
