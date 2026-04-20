# GoConnect Monitoring Stack 📊

## ✅ Monitoring System Deployed!

Your Kubernetes cluster now has a complete monitoring stack with Prometheus and Grafana.

## 🎯 Access Dashboards

### **Prometheus - Metrics Collection**
```
http://localhost:9090
```
**Features:**
- Real-time metrics scraping
- PromQL query language
- Alert management
- Service discovery

### **Grafana - Visualization**
```
http://localhost:3000
```
**Default Credentials:**
- Username: `admin`
- Password: `admin123`

**Features:**
- Beautiful dashboards
- Custom visualizations  
- Alerting
- Multi-datasource support

### **Traefik Dashboard - Load Balancer**
```
http://localhost:8888/dashboard/
```

## 📈 What You Can Monitor

### **1. Pod Health & Status**
```promql
# Running pods
up{job="kubernetes-pods"}

# Pod restarts
kube_pod_container_status_restarts_total
```

### **2. Database Connections**
```promql
# PostgreSQL connections (if metrics exposed)
pg_stat_database_numbackends

# Check pod status
kube_pod_status_phase{namespace="goconnect"}
```

### **3. Service Availability**
```promql
# Auth service endpoints
up{job="kubernetes-services",service="auth-service"}

# Gateway service endpoints
up{job="kubernetes-services",service="gateway-service"}
```

### **4. Resource Usage**
```promql
# Container CPU usage
container_cpu_usage_seconds_total{namespace="goconnect"}

# Container memory usage
container_memory_usage_bytes{namespace="goconnect"}
```

## 🔍 Prometheus Queries to Try

### Check if auth-service pods are up:
```promql
up{kubernetes_namespace="goconnect",app="auth-service"}
```

### Count running pods:
```promql
count(kube_pod_status_phase{namespace="goconnect",phase="Running"})
```

### Pod restart count:
```promql
sum(kube_pod_container_status_restarts_total{namespace="goconnect"}) by (pod)
```

## 📊 Create Grafana Dashboards

### **Step 1: Import Pre-built Dashboard**

1. Go to http://localhost:3000
2. Login (admin/admin123)
3. Click **Dashboards** → **Import**
4. Enter Dashboard ID:
   - **15661** - Kubernetes Cluster Monitoring
   - **13770** - Kubernetes Pods
   - **12125** - Traefik Official

5. Select **Prometheus** as datasource
6. Click **Import**

### **Step 2: Create Custom Dashboard**

1. Click **+** → **Dashboard** → **Add new panel**
2. Enter query:
   ```promql
   up{namespace="goconnect"}
   ```
3. Select visualization type
4. Save dashboard

## 🎨 Useful Grafana Panels

### **Auth Service Health Panel**
```promql
up{kubernetes_namespace="goconnect",app="auth-service"}
```
**Type:** Stat  
**Thresholds:** 0 = red, 1 = green

### **Gateway Request Rate Panel**
```promql
rate(http_requests_total{service="gateway"}[5m])
```
**Type:** Graph  
**Unit:** requests/sec

### **Pod Status Panel**
```promql
count(kube_pod_status_phase{namespace="goconnect"}) by (phase)
```
**Type:** Pie Chart

### **Database Connections Panel**
```promql
count(up{job="kubernetes-services",service="postgres-service"})
```
**Type:** Stat

## 🔔 Set Up Alerts (Future)

### **Pod Down Alert**
```yaml
- alert: PodDown
  expr: up{namespace="goconnect"} == 0
  for: 5m
  annotations:
    summary: "Pod {{ $labels.pod }} is down"
```

### **High Memory Usage**
```yaml
- alert: HighMemoryUsage
  expr: container_memory_usage_bytes > 400000000
  for: 5m
  annotations:
    summary: "High memory usage on {{ $labels.pod }}"
```

## 🛠️ Useful Commands

### **View Prometheus Targets**
```powershell
# Check what Prometheus is scraping
curl http://localhost:9090/api/v1/targets | ConvertFrom-Json
```

### **Query Metrics via CLI**
```powershell
# Get current pod status
curl "http://localhost:9090/api/v1/query?query=up{namespace='goconnect'}" | ConvertFrom-Json
```

### **Restart Monitoring Pods**
```powershell
kubectl rollout restart deployment/prometheus -n monitoring
kubectl rollout restart deployment/grafana -n monitoring
```

### **View Logs**
```powershell
# Prometheus logs
kubectl logs -f deployment/prometheus -n monitoring

# Grafana logs
kubectl logs -f deployment/grafana -n monitoring
```

## 📁 Configuration Files

- **Prometheus Config:** `@c:\Dev\Proj\GoConnect\deployments\k8s\prometheus.yaml`
- **Grafana Config:** `@c:\Dev\Proj\GoConnect\deployments\k8s\grafana.yaml`
- **Namespace:** `@c:\Dev\Proj\GoConnect\deployments\k8s\monitoring-namespace.yaml`

## 🚀 Quick Start Script

Run this to open all dashboards:
```powershell
.\scripts\open-monitoring.bat
```

## 📊 Next Steps

1. **Open Prometheus** → Check "Status" → "Targets" to see what's being monitored
2. **Open Grafana** → Import Kubernetes dashboard (ID: 15661)
3. **Create custom panels** for your auth and gateway services
4. **Set up alerts** when metrics go outside normal ranges
5. **Monitor your services** and identify issues in real-time

## 🔧 Troubleshooting

### Prometheus not scraping?
```powershell
# Check Prometheus logs
kubectl logs deployment/prometheus -n monitoring

# Check service discovery
curl http://localhost:9090/api/v1/targets
```

### Grafana can't connect to Prometheus?
```powershell
# Test connectivity from Grafana pod
kubectl exec -n monitoring deployment/grafana -- wget -O- http://prometheus:9090/api/v1/status/config
```

### No metrics showing?
Make sure your services expose metrics on the right port with annotations:
```yaml
annotations:
  prometheus.io/scrape: 'true'
  prometheus.io/port: '8080'
  prometheus.io/path: '/metrics'
```

---

**Your monitoring stack is ready! Start exploring your cluster metrics!** 🎉
