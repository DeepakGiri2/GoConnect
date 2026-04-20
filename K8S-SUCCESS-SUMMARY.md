# ✅ Kubernetes Cluster - Successfully Running!

## 🎉 Deployment Status

Your GoConnect application is now running in Kubernetes with:
- ✅ 6-node Kind cluster
- ✅ Traefik Ingress Controller
- ✅ Prometheus + Grafana monitoring
- ✅ PostgreSQL database
- ✅ Redis cache
- ✅ Gateway service (2 replicas)
- ✅ Auth service (2 replicas)

## 🌐 Access URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| **Gateway API** | http://localhost:8080 | - |
| **Traefik Dashboard** | http://localhost:8888 | - |
| **Grafana** | http://localhost:3000 | admin/admin123 |
| **Prometheus** | http://localhost:9090 | - |
| **PostgreSQL** | localhost:5432 | postgres/postgres |
| **Redis** | localhost:6379 | - |

## 📋 Current Pods

Run to see all pods:
```powershell
kubectl get pods -n goconnect
```

Expected output:
```
NAME                            READY   STATUS    RESTARTS   AGE
auth-service-74b99c8975-xxxxx   1/1     Running   0          Xm
auth-service-74b99c8975-yyyyy   1/1     Running   0          Xm
gateway-5474757447-xxxxx        1/1     Running   0          Xm
gateway-5474757447-yyyyy        1/1     Running   0          Xm
postgres-0                      1/1     Running   0          Xm
redis-0                         1/1     Running   0          Xm
```

## 🔍 Quick Commands Cheat Sheet

### View Logs
```powershell
# Gateway logs (live)
kubectl logs -f deployment/gateway -n goconnect

# Auth service logs (live)
kubectl logs -f deployment/auth-service -n goconnect

# PostgreSQL logs
kubectl logs postgres-0 -n goconnect

# Redis logs
kubectl logs redis-0 -n goconnect
```

### Get Pod Names
```powershell
# All pods in goconnect namespace
kubectl get pods -n goconnect

# Gateway pods only
kubectl get pods -n goconnect -l app=gateway

# Auth service pods only
kubectl get pods -n goconnect -l app=auth-service
```

### Execute Commands
```powershell
# PostgreSQL shell
kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect

# Redis CLI
kubectl exec -it redis-0 -n goconnect -- redis-cli

# Shell into gateway pod
kubectl exec -it deployment/gateway -n goconnect -- /bin/sh
```

### Check Status
```powershell
# All services
kubectl get svc -n goconnect

# Ingress status
kubectl get ingress -n goconnect

# Events (troubleshooting)
kubectl get events -n goconnect --sort-by=.metadata.creationTimestamp

# Resource usage
kubectl top pods -n goconnect
```

## 🐛 Debugging Setup

### Remote Debugging with VS Code

1. **Build debug images:**
   ```powershell
   .\scripts\k8s\build-debug-images.bat
   ```

2. **Deploy debug pods:**
   ```powershell
   .\scripts\k8s\deploy-debug.bat
   ```

3. **Connect debugger:**
   - Set breakpoints in your code
   - Press **F5** in VS Code
   - Select "Debug Gateway (Remote K8s)" or "Debug Auth Service (Remote K8s)"

**Debug Ports:**
- Gateway: `localhost:30040`
- Auth Service: `localhost:30041`

### Complete Debugging Guide
📖 See `deployments\k8s\DEBUG-SETUP.md` for full documentation

### Kubectl Reference
📖 See `deployments\k8s\KUBECTL-CHEATSHEET.md` for all commands

## 🛠️ Common Operations

### Restart Services
```powershell
# Rolling restart (zero downtime)
kubectl rollout restart deployment/gateway -n goconnect
kubectl rollout restart deployment/auth-service -n goconnect
```

### Scale Services
```powershell
# Scale gateway to 5 replicas
kubectl scale deployment gateway --replicas=5 -n goconnect

# Scale auth service to 3 replicas
kubectl scale deployment auth-service --replicas=3 -n goconnect
```

### Update Configuration
```powershell
# Apply updated deployment
kubectl apply -f deployments\k8s\gateway.yaml

# Check rollout status
kubectl rollout status deployment/gateway -n goconnect
```

### Database Operations
```powershell
# Run SQL query
kubectl exec postgres-0 -n goconnect -- psql -U postgres -d goconnect -c "SELECT COUNT(*) FROM users;"

# Backup database
kubectl exec postgres-0 -n goconnect -- pg_dump -U postgres goconnect > backup.sql

# Copy file from pod
kubectl cp goconnect/postgres-0:/backup.sql ./backup.sql
```

## 🔄 Start/Stop Cluster

### Stop Everything
```powershell
.\scripts\stop-all-k8s.bat
```

### Start Cluster (If Stopped)
```powershell
# Cluster already exists, just ensure Docker is running
kubectl config use-context kind-goconnect

# Check status
kubectl get pods -A
```

### Delete and Recreate Cluster
```powershell
# Delete cluster
.\scripts\k8s\delete-kind-cluster.bat

# Recreate
.\scripts\k8s\setup-kind-cluster.bat
.\scripts\k8s\build-images.bat
.\scripts\k8s\deploy-to-kind.bat
```

## 📊 Monitoring

### Grafana Dashboards
1. Open http://localhost:3000
2. Login: admin/admin123
3. Add Prometheus datasource (already configured)
4. Import dashboards for Go applications

### Prometheus Metrics
1. Open http://localhost:9090
2. Query metrics:
   - `rate(http_requests_total[5m])`
   - `go_goroutines`
   - `process_cpu_seconds_total`

### Traefik Dashboard
1. Open http://localhost:8888
2. View routing rules and backends

## 🚨 Troubleshooting

### Pod Not Starting
```powershell
# Check pod details
kubectl describe pod <pod-name> -n goconnect

# Check logs
kubectl logs <pod-name> -n goconnect

# Check previous logs (if crashed)
kubectl logs <pod-name> -n goconnect --previous
```

### Service Not Reachable
```powershell
# Check service endpoints
kubectl get endpoints -n goconnect

# Test from within cluster
kubectl run -it --rm debug --image=busybox -n goconnect -- wget -O- http://gateway-service
```

### Database Connection Issues
```powershell
# Check postgres is ready
kubectl exec postgres-0 -n goconnect -- pg_isready

# Test connection
kubectl exec postgres-0 -n goconnect -- psql -U postgres -c "\l"
```

## 📚 Documentation Files

| File | Description |
|------|-------------|
| `scripts/k8s/README.md` | Main K8s scripts documentation |
| `deployments/k8s/KUBECTL-CHEATSHEET.md` | Complete kubectl command reference |
| `deployments/k8s/DEBUG-SETUP.md` | Remote debugging guide |
| `deployments/k8s/INGRESS-SETUP.md` | Ingress controller architecture |

## 🎯 Next Steps

1. **Test the API:**
   ```powershell
   curl http://localhost:8080/health
   ```

2. **View logs to verify everything works:**
   ```powershell
   kubectl logs -f deployment/gateway -n goconnect
   ```

3. **Set up debugging** (optional):
   ```powershell
   .\scripts\k8s\build-debug-images.bat
   .\scripts\k8s\deploy-debug.bat
   ```

4. **Explore monitoring:**
   - Grafana: http://localhost:3000
   - Prometheus: http://localhost:9090

## 🎓 Learning Resources

- Run `.\scripts\k8s\show-cluster-info.bat` to see all cluster information
- Check pod logs regularly: `kubectl logs -f deployment/gateway -n goconnect`
- Monitor events: `kubectl get events -n goconnect -w`

---

**Cluster Status:** ✅ **RUNNING**

All services are deployed and ready!
