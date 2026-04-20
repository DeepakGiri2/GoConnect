# Kubernetes Scripts

All Kubernetes cluster management scripts for GoConnect.

## 📁 Scripts in This Folder

### **Cluster Setup**
- **`install-kind.bat`** - Download and install Kind for Windows
- **`setup-kind-cluster.bat`** - Create the 6-node Kind cluster
- **`delete-kind-cluster.bat`** - Delete the Kind cluster

### **Application Deployment**
- **`build-images.bat`** - Build Docker images for auth-service and gateway
- **`deploy-to-kind.bat`** - Deploy the application to Kind cluster
- **`deploy-to-kind-prod.bat`** - Deploy production configuration

### **Port Forwarding (Optional)**
- **`start-port-forwards.bat`** - Optional: Alternative port forwarding (not needed with NodePort setup)
- **`port-forward-services.bat`** - Alternative port forwarding script

### **Monitoring**
- **`open-monitoring.bat`** - Open Prometheus and Grafana dashboards
- **`open-dashboards.bat`** - Open Traefik and other dashboards

### **Debugging**
- **`build-debug-images.bat`** - Build debug Docker images with Delve
- **`deploy-debug.bat`** - Deploy debug pods with remote debugging enabled
- **`show-cluster-info.bat`** - Show all cluster info, pods, services, and quick commands

## 🚀 Quick Start

### **First Time Setup:**
```powershell
# 1. Install Kind (if not already installed)
.\scripts\k8s\install-kind.bat

# 2. Create the cluster
.\scripts\k8s\setup-kind-cluster.bat

# 3. Build Docker images
.\scripts\k8s\build-images.bat

# 4. Deploy the application
.\scripts\k8s\deploy-to-kind.bat

# 5. Access services (all exposed via NodePort, no port-forwarding needed!)
# Gateway: http://localhost:8080
# Grafana: http://localhost:3000
# Prometheus: http://localhost:9090
```

### **Daily Use:**
```powershell
# All services are directly accessible via NodePort
# Gateway:    http://localhost:8080
# Grafana:    http://localhost:3000 (admin/admin123)
# Prometheus: http://localhost:9090
# Traefik:    http://localhost:8888
```

### **Shutdown:**
```powershell
# Stop everything
.\scripts\stop-all-k8s.bat
```

## 🔧 Individual Commands

### **Rebuild and Redeploy:**
```powershell
.\scripts\k8s\build-images.bat
kubectl rollout restart deployment/gateway -n goconnect
kubectl rollout restart deployment/auth-service -n goconnect
```

### **View Logs:**
```powershell
kubectl logs -f deployment/gateway -n goconnect
kubectl logs -f deployment/auth-service -n goconnect
```

### **Check Status:**
```powershell
kubectl get pods -n goconnect
kubectl get all -n goconnect
```

## 📊 Access URLs

| Service | URL | Port | Type |
|---------|-----|------|------|
| Gateway API | http://localhost:8080 | 8080 | via Traefik Ingress |
| Traefik Dashboard | http://localhost:8888 | 8888 | NodePort |
| Grafana | http://localhost:3000 | 3000 | NodePort (admin/admin123) |
| Prometheus | http://localhost:9090 | 9090 | NodePort |
| PostgreSQL | localhost:5432 | 5432 | NodePort |
| Redis | localhost:6379 | 6379 | NodePort |

## 🐛 Debugging

### **The Multi-Pod Challenge:**
With 2+ pods, requests are load-balanced. You don't know which pod gets the request!

**Solution:** Scale to **1 debug pod** for predictable debugging.

### **Quick Start (GoLand or VS Code):**
```powershell
# 1. Build debug images (one-time)
.\scripts\k8s\build-debug-images.bat

# 2. Start debug session (scales to 1 pod)
.\scripts\k8s\start-debug-session.bat

# 3. Connect debugger:
#    GoLand: Run → Debug → "Debug Gateway Remote K8s"
#    VS Code: F5 → Select "Debug Gateway (Remote K8s)"

# 4. Make request → Breakpoint guaranteed to hit!
curl http://localhost:30081/health

# 5. When done, restore production
.\scripts\k8s\stop-debug-session.bat
```

**Debug Ports:**
- Gateway: `localhost:30040`
- Auth Service: `localhost:30041`

**Guides:**
- **GoLand:** `deployments\k8s\GOLAND-QUICK-START.md`
- **VS Code:** `deployments\k8s\DEBUG-SETUP.md`
- **Full Guide:** `deployments\k8s\GOLAND-DEBUG-GUIDE.md`

### **Kubectl Cheat Sheet:**
See `deployments\k8s\KUBECTL-CHEATSHEET.md` for all kubectl commands:
- View logs
- Get pod/container names
- Execute commands in pods
- Debug services
- And much more!

### **Cluster Information:**
```powershell
# Show all pods, services, and useful commands
.\scripts\k8s\show-cluster-info.bat
```

## �� Notes

- Gateway is accessed via Traefik Ingress Controller on port 8080
- Traefik handles all incoming HTTP traffic and routes to services
- All scripts assume you're in the repository root
- Use `stop-all-k8s.bat` to cleanly shutdown everything
