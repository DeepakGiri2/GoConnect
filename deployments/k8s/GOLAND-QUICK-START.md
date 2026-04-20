# GoLand Debugging - Quick Start

## 🎯 The Problem: 2 Pods = Load Balancing

When you have 2 gateway pods running:

```
Your Request → Traefik Ingress → Load Balancer
                                      ↓
                        ┌─────────────┴──────────────┐
                        ↓                            ↓
                 Gateway Pod 1              Gateway Pod 2
                (Debugger attached)         (No debugger)
                   50% chance                 50% chance
```

**Problem:** You don't know which pod gets the request!
- If Pod 1 gets it → ✅ Breakpoint hits
- If Pod 2 gets it → ❌ Request completes, no debugging

## ✅ Solution: Scale to 1 Pod for Debugging

```
Your Request → Traefik Ingress
                    ↓
              Gateway Pod 1
           (Debugger attached)
              100% guaranteed!
```

## 🚀 Super Quick Start (3 Commands)

```powershell
# 1. Start debug session (scales production to 0, debug to 1)
.\scripts\k8s\start-debug-session.bat
# Choose: gateway

# 2. In GoLand: Run → Debug → "Debug Gateway Remote K8s"

# 3. Make request
curl http://localhost:30081/health

# ✅ Breakpoint will hit every time!
```

## 📋 GoLand Setup (One-Time)

### Option 1: Use Pre-configured (Automatic)

The `.idea/runConfigurations/` folder already has debug configs!

1. Open GoLand
2. Go to **Run → Debug** dropdown
3. You'll see:
   - `Debug Gateway Remote K8s (Pod 1)`
   - `Debug Auth Service Remote K8s`

### Option 2: Create Manually

1. **Run** → **Edit Configurations...**
2. Click **+** → **Go Remote**
3. Fill in:
   - **Name:** `Debug Gateway Remote K8s`
   - **Host:** `localhost`
   - **Port:** `30040` (gateway) or `30041` (auth)
4. Click **OK**

## 🎮 Debugging Workflow

### Step 1: Build Debug Images (One-Time)

```powershell
.\scripts\k8s\build-debug-images.bat
```

This builds images with:
- Delve debugger included
- No optimizations (`-gcflags="all=-N -l"`)
- Debug port exposed (40000)

### Step 2: Start Debug Session

```powershell
.\scripts\k8s\start-debug-session.bat
```

This will:
1. ✅ Scale production pods to **0**
2. ✅ Deploy **1 debug pod**
3. ✅ Ensure predictable routing
4. ✅ Show connection info

Output:
```
Debug Session Ready!

Service:      gateway
Debug Pod:    gateway-debug-xxxxx-aaaaa
Debug Port:   localhost:30040
App Port:     localhost:30081

GoLand Setup:
1. Select "Debug gateway Remote K8s"
2. Click Debug (Shift+F9)
3. Set breakpoints and make requests!
```

### Step 3: Debug in GoLand

1. **Set breakpoints** in your code (click left gutter)
2. **Run** → **Debug** → Select `Debug Gateway Remote K8s`
3. **Wait** for console message: `Connected to localhost:30040`
4. **Make request:**
   ```powershell
   curl http://localhost:30081/health
   ```
5. **Breakpoint hits!** 🎉

### Step 4: Stop Debug Session

```powershell
.\scripts\k8s\stop-debug-session.bat
```

This will:
1. Scale debug pod to **0**
2. Restore production pods to **2**
3. Resume normal operation

## 🔍 How to Know Which Pod Gets Request (Advanced)

### Method 1: Watch Logs with Pod Names

```powershell
# Terminal 1: Watch all gateway logs
kubectl logs -f -l app=gateway -n goconnect --all-containers=true --prefix=true

# Make request
curl http://localhost:8080/health

# Output shows which pod:
# [pod/gateway-xxxxx-aaaaa] Handling request /health
```

### Method 2: Add Pod Name to Response

Add this to your handler:

```go
import "os"

func healthHandler(w http.ResponseWriter, r *http.Request) {
    podName := os.Getenv("HOSTNAME")
    
    w.Header().Set("X-Pod-Name", podName)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "pod": podName,
    })
}
```

Then check response:
```powershell
curl -v http://localhost:8080/health | findstr "X-Pod-Name"
# X-Pod-Name: gateway-5474757447-wj5q8
```

### Method 3: Port-Forward to Specific Pod

```powershell
# Get pod name
$POD = kubectl get pods -n goconnect -l app=gateway -o jsonpath='{.items[0].metadata.name}'

# Forward debug port from specific pod
kubectl port-forward -n goconnect $POD 40000:40000

# Forward app port from SAME pod
kubectl port-forward -n goconnect $POD 8082:8080

# Now requests to localhost:8082 ALWAYS go to that specific pod!
```

## 🎯 Best Practices

### ✅ DO

- ✅ Use **1 replica** when actively debugging
- ✅ Use `start-debug-session.bat` to automate setup
- ✅ Scale back to 2+ after debugging
- ✅ Add pod names to logs for visibility
- ✅ Test with production replicas after debugging

### ❌ DON'T

- ❌ Debug with 2+ replicas (unpredictable)
- ❌ Leave debug pods running in production
- ❌ Forget to scale back after debugging
- ❌ Debug multiple services simultaneously (confusing)

## 🐛 Troubleshooting

### "Failed to connect" in GoLand

```powershell
# Check debug pod is running
kubectl get pods -n goconnect -l app=gateway-debug

# Check Delve is listening
kubectl logs -l app=gateway-debug -n goconnect | findstr "API server listening"

# Test port
Test-NetConnection -ComputerName localhost -Port 30040
```

### Breakpoints Not Hitting

```powershell
# Ensure only 1 pod
kubectl get pods -n goconnect -l app=gateway-debug
# Should show ONLY 1 pod

# If multiple pods, scale to 1
kubectl scale deployment gateway-debug --replicas=1 -n goconnect
```

### Wrong Pod Getting Requests

```powershell
# Use debug service directly (bypasses ingress)
curl http://localhost:30081/health

# Or use port-forward to specific pod
kubectl port-forward -n goconnect <pod-name> 8082:8080
curl http://localhost:8082/health
```

## 📊 Debug Ports Reference

| Service | Debug Port | App Port | GoLand Config |
|---------|-----------|----------|---------------|
| Gateway | 30040 | 30081 | Debug Gateway Remote K8s |
| Auth Service | 30041 | 30051 | Debug Auth Service Remote K8s |

## 💡 Pro Tips

1. **Use debug service ports** (30081, 30051) to bypass load balancing
2. **Check pod count** before debugging: `kubectl get pods -n goconnect`
3. **Watch logs** while debugging: `kubectl logs -f -l app=gateway-debug -n goconnect`
4. **Auto-restart** on code changes: rebuild image → reload into Kind → restart pod
5. **Multiple breakpoints**: Set them all before starting debugger

## 🎓 Complete Example

```powershell
# 1. Build debug images (one-time)
.\scripts\k8s\build-debug-images.bat

# 2. Start debug session
.\scripts\k8s\start-debug-session.bat
# Choose: gateway

# 3. In GoLand:
#    - Set breakpoint in cmd/gateway/main.go
#    - Run → Debug → "Debug Gateway Remote K8s"
#    - Wait for "Connected"

# 4. Make request
curl http://localhost:30081/health

# 5. Breakpoint hits! Debug!
#    - Inspect variables
#    - Step through code
#    - Evaluate expressions

# 6. When done
.\scripts\k8s\stop-debug-session.bat
# Choose: gateway
```

## 🚀 Summary

**Problem:** 2 pods = 50% chance of hitting breakpoint

**Solution:** 1 debug pod = 100% guaranteed debugging

**Workflow:**
1. Build debug images
2. Start debug session (1 pod)
3. Debug in GoLand
4. Stop debug session (restore 2 pods)

**That's it!** 🎉

---

For full details, see `GOLAND-DEBUG-GUIDE.md`
