# GoLand Remote Debugging in Kubernetes

## 🎯 The Multi-Pod Debugging Challenge

When you have **2+ replicas** of a service, requests are load-balanced across pods. You don't know which pod will handle your request, making debugging difficult.

## ✅ Solutions

### Solution 1: **Scale Down to 1 Replica (Recommended for Debugging)**

This is the **simplest and most reliable** approach for active debugging:

```powershell
# Scale production deployment to 0
kubectl scale deployment gateway --replicas=0 -n goconnect

# Deploy single debug pod
kubectl apply -f deployments\k8s\gateway-debug.yaml

# Verify only 1 debug pod is running
kubectl get pods -n goconnect -l app=gateway-debug
```

**Pros:**
- ✅ All requests go to your debug pod
- ✅ Guaranteed to hit your breakpoints
- ✅ Simple setup

**Cons:**
- ⚠️ Not production-like (single instance)

---

### Solution 2: **Port-Forward to Specific Pod**

Debug a **specific pod** by name, bypassing load balancing:

```powershell
# 1. Get pod names
kubectl get pods -n goconnect -l app=gateway-debug

# Output example:
# NAME                            READY   STATUS
# gateway-debug-xxxxx-aaaaa      1/1     Running
# gateway-debug-xxxxx-bbbbb      1/1     Running

# 2. Port-forward to SPECIFIC pod (Pod 1)
kubectl port-forward -n goconnect gateway-debug-xxxxx-aaaaa 40000:40000

# 3. In another terminal, forward application port from SAME pod
kubectl port-forward -n goconnect gateway-debug-xxxxx-aaaaa 8081:8080

# 4. Connect GoLand debugger to localhost:40000
# 5. Send requests to http://localhost:8081 (goes to specific pod!)
```

**Pros:**
- ✅ Can debug specific pod
- ✅ Multiple pods can run
- ✅ You control which pod receives traffic

**Cons:**
- ⚠️ Need separate port for each pod
- ⚠️ Manual port-forward setup

---

### Solution 3: **Add Sticky Session / Pod Selector**

Modify your test requests to always hit the same pod:

#### A. Add Pod Name to Logs

First, identify which pod handles each request by adding pod name to logs:

```go
// In your handler (cmd/gateway/main.go or middleware)
podName := os.Getenv("HOSTNAME") // Kubernetes sets this to pod name
log.Printf("Request handled by pod: %s", podName)
```

Then watch logs to see which pod got the request:

```powershell
# Terminal 1: Watch pod 1 logs
kubectl logs -f gateway-debug-xxxxx-aaaaa -n goconnect

# Terminal 2: Watch pod 2 logs
kubectl logs -f gateway-debug-xxxxx-bbbbb -n goconnect

# Terminal 3: Make request
curl http://localhost:8080/health

# Look at which terminal shows the log - that's your pod!
```

#### B. Use Session Affinity (Sticky Sessions)

Modify the debug service to use session affinity:

```yaml
# In gateway-debug.yaml service section
apiVersion: v1
kind: Service
metadata:
  name: gateway-debug-service
  namespace: goconnect
spec:
  sessionAffinity: ClientIP  # <-- Add this
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 3600
  # ... rest of service config
```

**Pros:**
- ✅ Same client goes to same pod
- ✅ More production-like testing

**Cons:**
- ⚠️ Still need to identify initial pod
- ⚠️ Different IPs go to different pods

---

### Solution 4: **Debug Multiple Pods Simultaneously**

Set up debugging for both pods at once using different ports:

```powershell
# Terminal 1: Port-forward Pod 1 debug port
kubectl port-forward -n goconnect gateway-debug-xxxxx-aaaaa 40000:40000

# Terminal 2: Port-forward Pod 2 debug port
kubectl port-forward -n goconnect gateway-debug-xxxxx-bbbbb 40001:40000

# Connect GoLand debugger to localhost:40000 for Pod 1
# Connect another GoLand instance (or use multiple configs) to localhost:40001 for Pod 2
```

**Pros:**
- ✅ Debug any pod that gets the request
- ✅ Both pods are debuggable

**Cons:**
- ⚠️ Need to manage multiple debugger connections
- ⚠️ More complex setup

---

## 🚀 Recommended Debugging Workflow

### Step 1: Build and Deploy Debug Image

```powershell
# Build debug images with Delve
.\scripts\k8s\build-debug-images.bat
```

### Step 2: Deploy Single Debug Pod

```powershell
# Option A: Deploy debug alongside production
kubectl apply -f deployments\k8s\gateway-debug.yaml

# Option B: Replace production with debug (recommended)
kubectl scale deployment gateway --replicas=0 -n goconnect
kubectl apply -f deployments\k8s\gateway-debug.yaml
```

### Step 3: Verify Single Pod Running

```powershell
kubectl get pods -n goconnect -l app=gateway-debug

# Should see only 1 pod:
# NAME                            READY   STATUS
# gateway-debug-xxxxx-aaaaa      1/1     Running
```

### Step 4: Connect GoLand Debugger

1. **Open GoLand**
2. **Set breakpoints** in your code
3. **Click Run** → **Edit Configurations**
4. **Add New Configuration** → **Go Remote**
   - Name: `Debug Gateway Remote K8s`
   - Host: `localhost`
   - Port: `30040` (for gateway) or `30041` (for auth-service)
5. **Click Debug** (Shift+F9)
6. **Wait for "Connected"** message

### Step 5: Make Request

```powershell
# Request will hit your debug pod
curl http://localhost:8080/health

# Or use your debug service directly
curl http://localhost:30081/health
```

### Step 6: Debug!

- ✅ Breakpoint will hit
- ✅ Step through code
- ✅ Inspect variables
- ✅ Evaluate expressions

---

## 📋 GoLand Configuration (Manual Setup)

### Create Remote Debug Configuration

1. **Run** → **Edit Configurations...**
2. **Click +** → **Go Remote**
3. **Configure:**
   - **Name:** `Debug Gateway Remote K8s`
   - **Host:** `localhost`
   - **Port:** `30040` (gateway) or `30041` (auth-service)
   - **On disconnect:** Ask
4. **Click OK**

### Alternative: Import Configurations

GoLand configurations are already created in:
- `.idea/runConfigurations/Debug_Gateway_Remote_K8s_Pod_1.xml`
- `.idea/runConfigurations/Debug_Auth_Service_Remote_K8s.xml`

Just open the project and they'll appear in the Run/Debug dropdown!

---

## 🔍 How to Identify Which Pod Got the Request

### Method 1: Add Pod Name to Response Header

```go
// In your HTTP handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
    podName := os.Getenv("HOSTNAME")
    w.Header().Set("X-Pod-Name", podName)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "pod": podName,
    })
}
```

Then check response:
```powershell
curl -v http://localhost:8080/health

# Look for:
# < X-Pod-Name: gateway-debug-xxxxx-aaaaa
```

### Method 2: Watch Logs in Real-Time

```powershell
# Terminal 1: Pod 1 logs
kubectl logs -f gateway-debug-xxxxx-aaaaa -n goconnect

# Terminal 2: Pod 2 logs
kubectl logs -f gateway-debug-xxxxx-bbbbb -n goconnect

# Make request - see which terminal shows activity
curl http://localhost:8080/health
```

### Method 3: Use kubectl stern (Multi-Pod Log Viewer)

```powershell
# Install stern (if not already)
choco install stern

# View logs from ALL gateway pods with pod name prefix
stern gateway-debug -n goconnect

# Output shows which pod handled request:
# gateway-debug-xxxxx-aaaaa | Request received
```

---

## 🎯 Quick Reference

### Debugging Single Pod (Recommended)

```powershell
# 1. Scale down production
kubectl scale deployment gateway --replicas=0 -n goconnect

# 2. Ensure only 1 debug pod
kubectl scale deployment gateway-debug --replicas=1 -n goconnect

# 3. Verify
kubectl get pods -n goconnect -l app=gateway-debug

# 4. Debug in GoLand (localhost:30040)

# 5. Make request
curl http://localhost:8080/health
```

### Debugging Specific Pod

```powershell
# 1. Get pod name
$POD = kubectl get pods -n goconnect -l app=gateway-debug -o jsonpath='{.items[0].metadata.name}'

# 2. Port-forward debug port
kubectl port-forward -n goconnect $POD 40000:40000

# 3. Port-forward app port (to same pod)
kubectl port-forward -n goconnect $POD 8081:8080

# 4. Connect GoLand to localhost:40000

# 5. Request to localhost:8081 (specific pod)
curl http://localhost:8081/health
```

### View Which Pod Handles Requests

```powershell
# Add to code:
log.Printf("Pod %s handling request %s", os.Getenv("HOSTNAME"), r.URL.Path)

# Watch logs:
kubectl logs -f deployment/gateway-debug -n goconnect --all-containers=true
```

---

## 🛠️ Troubleshooting

### "Failed to connect" in GoLand

1. **Check pod is running:**
   ```powershell
   kubectl get pods -n goconnect -l app=gateway-debug
   ```

2. **Check Delve is listening:**
   ```powershell
   kubectl logs gateway-debug-xxxxx-aaaaa -n goconnect | findstr "API server listening"
   ```

3. **Test port connectivity:**
   ```powershell
   Test-NetConnection -ComputerName localhost -Port 30040
   ```

### Breakpoints not hitting

1. **Verify only 1 pod:**
   ```powershell
   kubectl get pods -n goconnect -l app=gateway-debug --no-headers | Measure-Object -Line
   # Should show: Count = 1
   ```

2. **Check request reaches the pod:**
   ```powershell
   kubectl logs -f gateway-debug-xxxxx-aaaaa -n goconnect
   # Make request, verify log output
   ```

3. **Ensure source code matches:**
   - Local code must match deployed image
   - Rebuild if you made changes

---

## 📊 Comparison: Multiple Pods vs Single Pod

| Aspect | Multiple Pods (2+) | Single Pod (1) |
|--------|-------------------|----------------|
| **Debugging Difficulty** | Hard | Easy |
| **Which pod gets request?** | Unknown (load balanced) | Always the same |
| **Breakpoint reliability** | 50% chance | 100% guaranteed |
| **Setup complexity** | High | Low |
| **Production similarity** | High | Low |
| **Recommended for** | Load testing | Active debugging |

**Verdict:** Use **single pod for debugging**, multiple pods for testing load balancing.

---

## 💡 Pro Tips

1. **Always scale to 1 replica when actively debugging**
2. **Use pod name in logs** to track which pod handles requests
3. **Port-forward to specific pod** when you need to debug that exact instance
4. **Scale back up** after debugging: `kubectl scale deployment gateway --replicas=2 -n goconnect`
5. **Use session affinity** for sticky debugging sessions

---

## 🎓 Full Example: Complete Debug Session

```powershell
# 1. Build debug image
.\scripts\k8s\build-debug-images.bat

# 2. Scale production to 0
kubectl scale deployment gateway --replicas=0 -n goconnect

# 3. Deploy single debug pod
kubectl apply -f deployments\k8s\gateway-debug.yaml
kubectl scale deployment gateway-debug --replicas=1 -n goconnect

# 4. Wait for ready
kubectl wait --for=condition=ready pod -l app=gateway-debug -n goconnect --timeout=60s

# 5. Verify single pod
kubectl get pods -n goconnect -l app=gateway-debug
# NAME                            READY   STATUS
# gateway-debug-xxxxx-aaaaa      1/1     Running

# 6. Open GoLand, set breakpoints

# 7. Run Debug Configuration "Debug Gateway Remote K8s"

# 8. Wait for "Connected to localhost:30040"

# 9. Make request
curl http://localhost:8080/health

# 10. Breakpoint hits! Debug away!

# 11. When done, scale back
kubectl delete deployment gateway-debug -n goconnect
kubectl scale deployment gateway --replicas=2 -n goconnect
```

---

**Remember:** For debugging, **1 pod = predictable debugging**. For testing, **multiple pods = realistic load balancing**.
