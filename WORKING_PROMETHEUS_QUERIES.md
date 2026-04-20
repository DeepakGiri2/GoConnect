# ✅ WORKING Prometheus Queries for Your Cluster

## 🎯 These Queries WORK RIGHT NOW!

### **1. See All Your Pods Status**
```promql
kube_pod_status_phase{namespace="goconnect"}
```
**What you'll see:**
- auth-service-xxx: phase="Running", value=1
- gateway-xxx: phase="Running", value=1  
- postgres-0: phase="Running", value=1
- redis-0: phase="Running", value=1

### **2. Count Running Pods in goconnect**
```promql
count(kube_pod_status_phase{namespace="goconnect",phase="Running"})
```
**Expected:** 6

### **3. See All Namespaces**
```promql
count(kube_pod_info) by (namespace)
```
**You'll see:** goconnect, monitoring, traefik, kube-system, etc.

### **4. Pod Info (Names, Nodes, IPs)**
```promql
kube_pod_info{namespace="goconnect"}
```

### **5. Container Status**
```promql
kube_pod_container_status_ready{namespace="goconnect"}
```
**1 = ready, 0 = not ready**

### **6. Pod Restart Count**
```promql
kube_pod_container_status_restarts_total{namespace="goconnect"}
```

### **7. All Pods Across All Namespaces**
```promql
kube_pod_status_phase
```

### **8. Pods on Each Node**
```promql
count(kube_pod_info) by (node)
```

### **9. Container Info**
```promql
kube_pod_container_info{namespace="goconnect"}
```

### **10. Deployment Replicas**
```promql
kube_deployment_status_replicas{namespace="goconnect"}
```

## 📊 For Grafana Dashboards

### **Import These Working Dashboards:**

**Dashboard ID: 15661** - Kubernetes Cluster Monitoring (Highly Recommended!)
```
1. Go to Grafana: http://localhost:3001
2. Login: admin / admin123
3. Click + → Import
4. Enter: 15661
5. Select Prometheus datasource
6. Click Import
```

**Dashboard ID: 13770** - Kubernetes Pods
```
Great for seeing pod status, restarts, and resource usage
```

**Dashboard ID: 14981** - Kubernetes System Stats
```
Shows overall cluster health
```

## 🔧 Why Your App Metrics Aren't Showing

**Your Go services don't expose metrics endpoints yet!**

To add metrics to your services, you need to:

1. **Add Prometheus client to Go code:**
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Add to main()
http.Handle("/metrics", promhttp.Handler())
```

2. **Annotate your Kubernetes deployments:**
```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"  
    prometheus.io/path: "/metrics"
```

3. **Custom metrics examples:**
```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

## ✅ What's Working Now

**Prometheus IS collecting:**
- ✅ All pod statuses
- ✅ Container information
- ✅ Deployment replicas
- ✅ Pod restarts
- ✅ Node information
- ✅ Namespace data
- ✅ kube-state-metrics
- ✅ Prometheus itself

**Not yet available (need to add):**
- ❌ HTTP request rates
- ❌ Auth service metrics
- ❌ Gateway metrics
- ❌ Custom business metrics
- ❌ Database query performance
- ❌ Redis cache hits

## 🚀 Try These NOW

**Open Prometheus:** http://localhost:9091

**Paste this query:**
```promql
kube_pod_status_phase{namespace="goconnect"}
```

**Click Execute** - You'll see all your pods!

**Then try:**
```promql
count(kube_pod_status_phase{namespace="goconnect",phase="Running"})
```

**Should show: 6** ✅

## 📊 Grafana Quick Setup

1. **Import Dashboard 15661:**
   - Go to http://localhost:3001
   - Click **+ → Import**
   - Type: **15661**
   - Select **Prometheus**
   - Click **Import**

2. **You'll immediately see:**
   - Cluster CPU usage
   - Memory usage
   - Pod status
   - Network traffic
   - All your 6 nodes
   - All your pods

**This works RIGHT NOW because kube-state-metrics is running!** 🎉

---

**Try it: Open Prometheus, paste `kube_pod_status_phase{namespace="goconnect"}` and click Execute!**
