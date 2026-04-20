# Prometheus Queries for GoConnect 📊

## 🎯 Essential Queries for Your Cluster

### **1. Check All Pods Status**
```promql
kube_pod_status_phase{namespace="goconnect"}
```
**What it shows:** Current phase of each pod (Running=1, Pending=0, Failed=3)

---

### **2. Count Running Pods**
```promql
count(kube_pod_status_phase{namespace="goconnect",phase="Running"})
```
**Expected:** Should show 6 (2 auth, 2 gateway, 1 postgres, 1 redis)

---

### **3. Pod Restart Count**
```promql
kube_pod_container_status_restarts_total{namespace="goconnect"}
```
**What to look for:** High restart counts indicate pods crashing

---

### **4. Auth Service Health**
```promql
up{namespace="goconnect",app="auth-service"}
```
**Expected:** 1 for each healthy pod, 0 for down pods

---

### **5. Gateway Service Health**
```promql
up{namespace="goconnect",app="gateway"}
```
**Expected:** 1 for each healthy pod

---

### **6. Database Pod Status**
```promql
kube_pod_status_phase{namespace="goconnect",pod=~"postgres.*"}
```
**Expected:** 1 (Running)

---

### **7. Redis Pod Status**
```promql
kube_pod_status_phase{namespace="goconnect",pod=~"redis.*"}
```
**Expected:** 1 (Running)

---

### **8. All Services Discovery**
```promql
up{job="kubernetes-services"}
```
**What it shows:** Which services Prometheus is successfully scraping

---

### **9. Node Status**
```promql
kube_node_status_condition{condition="Ready",status="true"}
```
**Expected:** 6 (1 control-plane + 5 workers)

---

### **10. Traefik Load Balancer Status**
```promql
up{namespace="traefik"}
```
**Expected:** 1 (healthy)

---

## 📈 Advanced Queries

### **Container Memory Usage**
```promql
container_memory_usage_bytes{namespace="goconnect",pod=~"auth-service.*"}
```

### **Container CPU Usage**
```promql
rate(container_cpu_usage_seconds_total{namespace="goconnect"}[5m])
```

### **Pod Age**
```promql
time() - kube_pod_created{namespace="goconnect"}
```

### **Containers Not Ready**
```promql
kube_pod_container_status_ready{namespace="goconnect"} == 0
```

### **Pod Distribution Across Nodes**
```promql
count(kube_pod_info{namespace="goconnect"}) by (node)
```

---

## 🎨 Grafana Panel Examples

### **Panel 1: Pod Status Gauge**
- **Query:** `count(kube_pod_status_phase{namespace="goconnect",phase="Running"})`
- **Visualization:** Stat
- **Title:** "Running Pods"
- **Thresholds:** <4=red, 4-5=yellow, 6=green

### **Panel 2: Auth Service Health**
- **Query:** `up{namespace="goconnect",app="auth-service"}`
- **Visualization:** Time series
- **Title:** "Auth Service Uptime"
- **Legend:** `{{pod}}`

### **Panel 3: Pod Restarts**
- **Query:** `sum(kube_pod_container_status_restarts_total{namespace="goconnect"}) by (pod)`
- **Visualization:** Bar chart
- **Title:** "Pod Restart Count"

### **Panel 4: Service Distribution**
- **Query:** `count(kube_pod_info{namespace="goconnect"}) by (node)`
- **Visualization:** Pie chart
- **Title:** "Pods per Node"

---

## 🔍 Debugging Specific Issues

### **Find Which Auth Pod is Down**
```promql
kube_pod_status_phase{namespace="goconnect",pod=~"auth-service.*",phase!="Running"}
```

### **Check Database Connectivity Issues**
```promql
up{service="postgres-service"} == 0
```

### **Find Pods with High Restarts**
```promql
kube_pod_container_status_restarts_total{namespace="goconnect"} > 5
```

### **Memory Pressure Detection**
```promql
container_memory_usage_bytes{namespace="goconnect"} / container_spec_memory_limit_bytes{namespace="goconnect"} > 0.8
```

---

## 🎯 Quick Dashboard Setup in Grafana

1. **Login to Grafana:** http://localhost:3000 (admin/admin123)

2. **Import Kubernetes Dashboard:**
   - Click **+** → **Import**
   - Enter ID: **15661** (Kubernetes Cluster Monitoring)
   - Select **Prometheus** datasource
   - Click **Import**

3. **Create Custom Dashboard:**
   - Click **+** → **Create Dashboard**
   - Add Panel → Select queries above
   - Arrange and save

4. **Set Up Alerts (Optional):**
   - Edit panel → Alert tab
   - Set condition: `when last() of query(A) is below 1`
   - Configure notification channel

---

## 📊 What You Should See Now

**Healthy Cluster Metrics:**
- ✅ 6 running pods in `goconnect` namespace
- ✅ 0 pod restarts
- ✅ All services showing `up{...} = 1`
- ✅ 2 auth-service endpoints
- ✅ 2 gateway endpoints
- ✅ 1 postgres endpoint
- ✅ 1 redis endpoint
- ✅ 1 traefik endpoint

---

## 🚀 Try It Now!

1. Open **Prometheus:** http://localhost:9090
2. Click **Graph** tab
3. Paste this query:
   ```promql
   count(kube_pod_status_phase{namespace="goconnect",phase="Running"})
   ```
4. Click **Execute**
5. You should see **6** (all your pods!)

Then try the other queries to explore your cluster! 🎉
