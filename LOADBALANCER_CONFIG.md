# Traefik Load Balancer Configuration

## ✅ Current Load Balancing Algorithm: **Least Connections**

### **What Changed:**

Your Traefik load balancer now uses the **Least Connections** algorithm instead of the default Round Robin.

### **How It Works:**

```
Request → Traefik Load Balancer
          ↓
    Checks active connections on:
    - Gateway Pod 1 (worker2): 5 active connections
    - Gateway Pod 2 (worker3): 3 active connections ← Routes here!
          ↓
    Routes to pod with FEWER active connections
```

### **Algorithms Available:**

| Algorithm | Description | Use Case |
|-----------|-------------|----------|
| **leastconn** ✅ | Routes to instance with fewest active connections | **Current** - Best for long-running requests |
| `roundrobin` | Distributes evenly across all instances | Default - Simple distribution |
| `wrr` | Weighted round robin based on server weights | When servers have different capacities |

### **Configuration Location:**

`@c:\Dev\Proj\GoConnect\deployments\k8s\traefik.yaml:189-203`

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: goconnect-ingressroute
  namespace: goconnect
spec:
  entryPoints:
    - web
  routes:
  - match: PathPrefix(`/`)
    kind: Rule
    services:
    - name: gateway-service
      port: 80
      strategy: leastconn  # ← Least Connections algorithm
```

### **Verify It's Working:**

```powershell
# Check IngressRoute
kubectl get ingressroute -n goconnect

# Describe to see strategy
kubectl describe ingressroute goconnect-ingressroute -n goconnect

# Check Traefik logs
kubectl logs -n traefik deployment/traefik --tail=50
```

### **Test Load Distribution:**

```powershell
# Send multiple requests
for ($i=1; $i -le 10; $i++) {
    Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
    Write-Host "Request $i sent"
}

# Watch which pods handle requests
kubectl logs -f deployment/gateway -n goconnect
```

### **View in Traefik Dashboard:**

1. Open: http://localhost:8888/dashboard/
2. Go to **HTTP → Services**
3. Click on **gateway-service**
4. You'll see the load balancer configuration with `leastconn` strategy

### **Benefits of Least Connections:**

✅ **Smart routing** - Routes to less busy instances  
✅ **Better for long requests** - Avoids overloading single instance  
✅ **Handles variable request times** - Adapts to different workload patterns  
✅ **Improved performance** - More even resource utilization  

### **How to Change Algorithm:**

To switch to a different algorithm, edit the IngressRoute:

```yaml
services:
- name: gateway-service
  port: 80
  strategy: roundrobin  # or wrr (weighted round robin)
```

Then apply:
```powershell
kubectl apply -f deployments\k8s\traefik.yaml
```

---

**Your load balancer is now optimized to route to the least busy Gateway instance!** 🎯
