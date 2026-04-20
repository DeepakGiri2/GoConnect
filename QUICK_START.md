# GoConnect Kubernetes Cluster - Quick Start Guide 🚀

## ✅ Your Cluster is Running!

### **📊 Access Dashboards (NEW PORTS!)**

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3001 | admin / admin123 |
| **Prometheus** | http://localhost:9091 | No login |
| **Traefik LB** | http://localhost:8888/dashboard/ | No login |
| **Gateway API** | http://localhost:8080 | - |

> **Note:** Port 3000 is reserved for your frontend, so Grafana uses **3001**

### **🎯 Quick Start Steps:**

**1. Start Port Forwards:**
```powershell
.\scripts\start-port-forwards.bat
```

**2. Open Monitoring Dashboards:**
```powershell
.\scripts\open-monitoring.bat
```

**3. Import Kubernetes Dashboard in Grafana:**
- Open http://localhost:3001
- Login: admin / admin123
- Click **+ → Import**
- Enter Dashboard ID: **15661**
- Select **Prometheus** datasource
- Click **Import**

## 📈 Test Prometheus Queries

Open http://localhost:9091 and try these:

### **Count Running Pods:**
```promql
kube_pod_status_phase{namespace="goconnect",phase="Running"}
```

### **Check Auth Service:**
```promql
up{namespace="goconnect",app="auth-service"}
```

### **Pod Restarts:**
```promql
kube_pod_container_status_restarts_total{namespace="goconnect"}
```

## 🏗️ Cluster Architecture

```
6-Node Kubernetes Cluster
├── Control Plane (goconnect-control-plane)
│   └── Ports: 8080, 8443, 5432, 6379, 8888, 9091, 3001
│
└── Workers:
    ├── worker1 (Traefik) - Load Balancer
    ├── worker2 (Services) - 1x Gateway + 1x Auth
    ├── worker3 (Services) - 1x Gateway + 1x Auth  
    ├── worker4 (PostgreSQL) - Database
    └── worker5 (Redis) - Cache
```

## 📦 What's Running

**Application (goconnect namespace):**
- ✅ Auth Service (2 replicas)
- ✅ Gateway (2 replicas)
- ✅ PostgreSQL (1 replica)
- ✅ Redis (1 replica)

**Load Balancer (traefik namespace):**
- ✅ Traefik (1 replica) - Least Connections algorithm

**Monitoring (monitoring namespace):**
- ✅ Prometheus (1 replica) - Metrics collection
- ✅ Grafana (1 replica) - Dashboards

## 🔧 Useful Commands

### **View All Pods:**
```powershell
kubectl get pods -n goconnect
kubectl get pods -n monitoring  
kubectl get pods -n traefik
```

### **Check Services:**
```powershell
kubectl get svc -n goconnect
kubectl get svc -n monitoring
```

### **View Logs:**
```powershell
kubectl logs -f deployment/gateway -n goconnect
kubectl logs -f deployment/auth-service -n goconnect
kubectl logs -f deployment/prometheus -n monitoring
kubectl logs -f deployment/grafana -n monitoring
```

### **Restart Deployments:**
```powershell
kubectl rollout restart deployment/gateway -n goconnect
kubectl rollout restart deployment/auth-service -n goconnect
kubectl rollout restart deployment/prometheus -n monitoring
kubectl rollout restart deployment/grafana -n monitoring
```

### **Database Access:**
```powershell
# Via port-forward (already included in start-port-forwards.bat)
psql -h localhost -p 5432 -U postgres -d goconnect
# Password: postgres

# Or directly in pod
kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect
```

## 🎨 Available Grafana Dashboards

**Pre-built Kubernetes Dashboards (Import by ID):**
- **15661** - Kubernetes Cluster Monitoring (Recommended)
- **13770** - Kubernetes Pods
- **12125** - Traefik Official Dashboard
- **1860** - Node Exporter Full

**How to Import:**
1. Go to Grafana → **+ → Import**
2. Enter dashboard ID
3. Select **Prometheus** as datasource
4. Click **Import**

## 📊 Database Schema

**Tables in `goconnect` database:**
- `users` - Verified active users
- `unverified_users` - Pending email verification
- `oauth_accounts` - OAuth provider links
- `refresh_tokens` - JWT refresh tokens

**Check tables:**
```powershell
kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect -c "\dt"
```

## 🚨 Troubleshooting

### **Port-forwards keep dying:**
```powershell
# Restart all
.\scripts\start-port-forwards.bat
```

### **Dashboards are empty:**
```powershell
# Wait 30-60 seconds for Prometheus to scrape metrics
# Check Prometheus targets: http://localhost:9091/targets
```

### **Can't access services:**
```powershell
# Check if pods are running
kubectl get pods -n goconnect

# Check port-forwards are active
Get-Process kubectl
```

### **Database connection issues:**
```powershell
# Check if postgres is running
kubectl get pods -n goconnect -l app=postgres

# Check database credentials
kubectl get secret goconnect-secrets -n goconnect -o yaml
```

## 📝 Configuration Files

- **Cluster:** `deployments/k8s/kind-config.yaml`
- **Database:** `deployments/k8s/postgres.yaml`
- **Auth Service:** `deployments/k8s/auth-service.yaml`
- **Gateway:** `deployments/k8s/gateway.yaml`
- **Traefik:** `deployments/k8s/traefik.yaml`
- **Prometheus:** `deployments/k8s/prometheus.yaml`
- **Grafana:** `deployments/k8s/grafana.yaml`
- **Secrets:** `deployments/k8s/secrets.yaml`
- **ConfigMap:** `deployments/k8s/configmap.yaml`

## 🎯 Next Steps

1. ✅ Import Grafana dashboard (ID: 15661)
2. ✅ Check Prometheus targets: http://localhost:9091/targets
3. ✅ Test API: http://localhost:8080/health
4. ✅ View Traefik dashboard: http://localhost:8888/dashboard/
5. ⬜ Add custom metrics to your services
6. ⬜ Set up alerts in Grafana
7. ⬜ Deploy your frontend to port 3000

## 📚 Documentation

- **Full Access Guide:** `CLUSTER_ACCESS.md`
- **Monitoring Guide:** `MONITORING_GUIDE.md`
- **Prometheus Queries:** `PROMETHEUS_QUERIES.md`
- **Load Balancer Config:** `LOADBALANCER_CONFIG.md`

---

**Your cluster is fully operational! Start by opening Grafana and importing dashboard 15661!** 🎉
