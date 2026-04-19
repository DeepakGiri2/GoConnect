# GoConnect Kubernetes Cluster - Access Guide

## ✅ Your Services Are Running!

### **Load Balancer (Traefik) - WORKING**
```
http://localhost:8080          # API Gateway through Traefik
http://localhost:8443          # HTTPS (if configured)
```

### **Traefik Dashboard**
**Current Session (Port Forward):**
```
http://localhost:8888/dashboard/
```

**After Next Cluster Recreation:**
```
http://localhost:8888/dashboard/  # Will be directly accessible without port-forward
```

> **Note:** The dashboard port (30888) is now added to the Kind config. The next time you recreate the cluster, it will be automatically exposed.

### **Direct Services Access**
```
PostgreSQL:  localhost:5432    # Username: postgres, Password: postgres, Database: goconnect
Redis:       localhost:6379
```

### **Test Your API**
```powershell
# Health check
curl http://localhost:8080/health

# Should return: {"status":"healthy"}
```

### **View Your Cluster**

**All Pods:**
```powershell
kubectl get pods -n goconnect
kubectl get pods -n traefik
```

**Services:**
```powershell
kubectl get svc -n goconnect
kubectl get svc -n traefik
```

**Logs:**
```powershell
# Gateway logs
kubectl logs -f deployment/gateway -n goconnect

# Auth Service logs
kubectl logs -f deployment/auth-service -n goconnect

# Traefik logs
kubectl logs -f deployment/traefik -n traefik
```

### **Database Access**
```powershell
# Connect to PostgreSQL
psql -h localhost -p 5432 -U postgres -d goconnect
# Password: postgres

# Or via kubectl
kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect
```

### **Kubernetes Dashboard (Optional)**
```powershell
# Install dashboard
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml

# Access it
kubectl proxy
# Then open: http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/
```

## 📊 Your 6-Node Cluster Architecture

```
Master Node (control-plane)
├── Exposed Ports: 8080, 8443, 5432, 6379
│
Worker Nodes:
├── worker1 (Traefik)      - Load Balancer  
├── worker2 (Services)     - 1x Gateway + 1x Auth
├── worker3 (Services)     - 1x Gateway + 1x Auth
├── worker4 (PostgreSQL)   - Database
└── worker5 (Redis)        - Cache
```

## 🔧 Useful Commands

```powershell
# View all resources
kubectl get all -n goconnect

# Describe a pod
kubectl describe pod <pod-name> -n goconnect

# Execute command in pod
kubectl exec -it <pod-name> -n goconnect -- /bin/sh

# Port forward (if needed)
kubectl port-forward -n goconnect svc/gateway-service 9090:80

# Restart deployments
kubectl rollout restart deployment/gateway -n goconnect
kubectl rollout restart deployment/auth-service -n goconnect

# Delete and recreate cluster
kind delete cluster --name goconnect
kind create cluster --config=deployments/k8s/kind-config.yaml
```

## 🚀 Quick Start

1. **Access your API:** http://localhost:8080
2. **Check health:** http://localhost:8080/health
3. **View logs:** `kubectl logs -f deployment/gateway -n goconnect`
4. **Connect to DB:** `psql -h localhost -p 5432 -U postgres -d goconnect`

Your cluster is fully operational! 🎉
