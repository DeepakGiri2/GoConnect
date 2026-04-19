# Deployment Guide

## Prerequisites

- Docker & Docker Compose
- Kubernetes cluster (for production)
- kubectl configured
- PostgreSQL client (for manual setup)

## Local Development

### Using Docker Compose

1. **Setup environment**
```bash
cp .env.example .env
# Edit .env with your configuration
```

2. **Start services**
```bash
docker-compose -f docker\docker-compose.dev.yml up --build
```

3. **Access services**
- API Gateway: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

4. **Stop services**
```bash
docker-compose -f docker\docker-compose.dev.yml down
```

### Running Services Locally

1. **Start PostgreSQL and Redis**
```bash
docker-compose -f docker\docker-compose.dev.yml up postgres redis
```

2. **Run database migrations**
```bash
psql -h localhost -U postgres -d goconnect -f shared\db\migrations\001_initial_schema.sql
```

3. **Start Auth Service**
```bash
cd services\auth
go run cmd\main.go
```

4. **Start Gateway** (in new terminal)
```bash
cd services\gateway
go run cmd\main.go
```

## Production Deployment

### Docker Compose (Small-Medium Scale)

1. **Update environment variables**
```bash
cp .env.example .env
# Update with production values
```

2. **Deploy**
```bash
docker-compose -f docker\docker-compose.prod.yml up -d
```

3. **Scale services**
```bash
docker-compose -f docker\docker-compose.prod.yml up -d --scale gateway=3 --scale auth-service=3
```

4. **Monitor logs**
```bash
docker-compose -f docker\docker-compose.prod.yml logs -f
```

### Kubernetes (Large Scale)

#### 1. Build and Push Images

```bash
# Build images
docker build -t your-registry/goconnect-auth:latest -f docker/Dockerfile.auth .
docker build -t your-registry/goconnect-gateway:latest -f docker/Dockerfile.gateway .

# Push to registry
docker push your-registry/goconnect-auth:latest
docker push your-registry/goconnect-gateway:latest
```

#### 2. Update Kubernetes Manifests

Edit `k8s/secrets.yaml` and `k8s/configmap.yaml` with your configuration.

Update image names in `k8s/auth-service.yaml` and `k8s/gateway.yaml`.

#### 3. Deploy to Kubernetes

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create secrets and configmap
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/configmap.yaml

# Deploy database and cache
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml

# Wait for databases to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n goconnect --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n goconnect --timeout=300s

# Deploy services
kubectl apply -f k8s/auth-service.yaml
kubectl apply -f k8s/gateway.yaml

# Setup ingress (optional)
kubectl apply -f k8s/ingress.yaml
```

#### 4. Verify Deployment

```bash
# Check pods
kubectl get pods -n goconnect

# Check services
kubectl get services -n goconnect

# Check logs
kubectl logs -f deployment/gateway -n goconnect
kubectl logs -f deployment/auth-service -n goconnect
```

#### 5. Access Services

```bash
# Port forward for testing
kubectl port-forward service/gateway-service 8080:80 -n goconnect

# Or get LoadBalancer IP
kubectl get service gateway-service -n goconnect
```

## Database Migrations

### Initial Setup

```bash
psql -h localhost -U postgres -d goconnect -f shared/db/migrations/001_initial_schema.sql
```

### In Kubernetes

```bash
# Copy migration file to postgres pod
kubectl cp shared/db/migrations/001_initial_schema.sql goconnect/postgres-0:/tmp/

# Execute migration
kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect -f /tmp/001_initial_schema.sql
```

## OAuth Configuration

### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create new project
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add authorized redirect URI: `https://your-domain.com/api/auth/callback/google`
6. Update secrets with Client ID and Secret

### Facebook OAuth

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create new app
3. Add Facebook Login product
4. Add redirect URI: `https://your-domain.com/api/auth/callback/facebook`
5. Update secrets with App ID and Secret

### GitHub OAuth

1. Go to [GitHub Settings → Developer Settings](https://github.com/settings/developers)
2. Create new OAuth App
3. Add callback URL: `https://your-domain.com/api/auth/callback/github`
4. Update secrets with Client ID and Secret

## SSL/TLS Configuration

### Using Let's Encrypt with Kubernetes

1. **Install cert-manager**
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

2. **Create ClusterIssuer**
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

3. **Apply ingress** (already configured in `k8s/ingress.yaml`)

## Monitoring and Logging

### View Logs

**Docker Compose:**
```bash
docker-compose -f docker/docker-compose.prod.yml logs -f gateway
docker-compose -f docker/docker-compose.prod.yml logs -f auth-service
```

**Kubernetes:**
```bash
kubectl logs -f deployment/gateway -n goconnect
kubectl logs -f deployment/auth-service -n goconnect
```

### Health Checks

```bash
# Gateway health
curl http://localhost:8080/health

# In Kubernetes
kubectl exec -it deployment/gateway -n goconnect -- wget -qO- http://localhost:8080/health
```

## Scaling

### Docker Compose

```bash
docker-compose -f docker/docker-compose.prod.yml up -d --scale gateway=5 --scale auth-service=5
```

### Kubernetes

```bash
# Manual scaling
kubectl scale deployment gateway --replicas=5 -n goconnect
kubectl scale deployment auth-service --replicas=5 -n goconnect

# Auto-scaling (HPA already configured)
kubectl get hpa -n goconnect
```

## Backup and Recovery

### Database Backup

```bash
# Docker
docker exec goconnect-postgres pg_dump -U postgres goconnect > backup.sql

# Kubernetes
kubectl exec postgres-0 -n goconnect -- pg_dump -U postgres goconnect > backup.sql
```

### Database Restore

```bash
# Docker
docker exec -i goconnect-postgres psql -U postgres goconnect < backup.sql

# Kubernetes
kubectl exec -i postgres-0 -n goconnect -- psql -U postgres goconnect < backup.sql
```

## Troubleshooting

### Service Won't Start

1. Check logs for errors
2. Verify environment variables
3. Ensure database is accessible
4. Check port availability

### Database Connection Issues

1. Verify database is running
2. Check connection credentials
3. Verify network connectivity
4. Check firewall rules

### OAuth Not Working

1. Verify OAuth credentials
2. Check redirect URLs match exactly
3. Ensure HTTPS in production
4. Verify OAuth app is enabled

### High Memory/CPU Usage

1. Check for memory leaks in logs
2. Scale up resources
3. Enable HPA for auto-scaling
4. Optimize database queries

## Security Checklist

- [ ] Change default passwords
- [ ] Update JWT secret
- [ ] Update OTP secret
- [ ] Configure OAuth credentials
- [ ] Enable HTTPS/TLS
- [ ] Set up firewall rules
- [ ] Enable rate limiting
- [ ] Regular security updates
- [ ] Monitor logs for suspicious activity
- [ ] Set up backup strategy

## Performance Optimization

1. **Database**
   - Enable connection pooling
   - Add indexes for frequently queried fields
   - Regular VACUUM and ANALYZE

2. **Redis**
   - Configure persistence
   - Set max memory policy
   - Enable AOF for durability

3. **Services**
   - Enable gRPC connection pooling
   - Optimize JWT token validation
   - Cache frequently accessed data

4. **Load Balancing**
   - Use sticky sessions if needed
   - Configure health checks
   - Set appropriate timeouts
