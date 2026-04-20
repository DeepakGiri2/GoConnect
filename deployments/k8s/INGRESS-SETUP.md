# Traefik Ingress Controller Setup

## Architecture

```
Internet/Localhost:8080
        ↓
   Traefik (NodePort 30080 → 80)
        ↓
   Ingress Controller
        ↓
   Gateway Service (ClusterIP:80)
        ↓
   Gateway Pods (8080)

Monitoring Stack (Direct NodePort Access):
- Grafana:    NodePort 30300 → localhost:3000
- Prometheus: NodePort 30090 → localhost:9090
- Traefik:    NodePort 30888 → localhost:8888
```

## Changes Made

### Fixed Port Conflicts
- **Removed**: `gateway-kind.yaml` (was causing NodePort 30080 conflict with Traefik)
- **Changed**: `gateway-service` from LoadBalancer → ClusterIP
- **Updated**: `ingress.yaml` from nginx → traefik

### Configuration
- **Ingress Controller**: Traefik v2.10
- **IngressClass**: traefik
- **Entry Point**: web (port 80)
- **Gateway Access**: http://localhost:8080 (via Traefik)

## How It Works

### Application Traffic
1. **Kind Cluster** exposes Traefik port 30080 to host port 8080
2. **Traefik** receives all traffic on port 80 (mapped to host 8080)
3. **Ingress Resource** routes `/` to `gateway-service`
4. **Gateway Service** (ClusterIP) forwards to gateway pods on port 8080

### Monitoring Stack
- **Grafana** and **Prometheus** use **NodePort** for direct access
- No port-forwarding required - all services accessible immediately after deployment
- Monitoring namespace keeps observability stack isolated

## No Port Conflicts or Warnings

All services use appropriate types:
- **Traefik**: NodePort 30080 → 8080 (ingress traffic)
- **Gateway**: ClusterIP (internal only, accessed via Ingress)
- **Grafana**: NodePort 30300 → 3000 (monitoring UI)
- **Prometheus**: NodePort 30090 → 9090 (metrics)
- **Postgres**: NodePort 30432 → 5432 (database)
- **Redis**: NodePort 30379 → 6379 (cache)

## Testing

```powershell
# Check all deployments
kubectl get pods -n goconnect
kubectl get pods -n traefik
kubectl get pods -n monitoring

# Check ingress status
kubectl get ingress -n goconnect

# Test gateway
curl http://localhost:8080/health

# Access dashboards
start http://localhost:8080        # Gateway API
start http://localhost:8888        # Traefik Dashboard
start http://localhost:3000        # Grafana (admin/admin123)
start http://localhost:9090        # Prometheus
```

## Complete Deployment

Run the deployment script - everything will be set up automatically:

```powershell
.\scripts\k8s\deploy-to-kind.bat
```

All services will be immediately accessible via NodePort - **no port-forwarding needed!**
