# Remote Debugging Setup for GoConnect

## 🐛 Overview

This guide shows how to debug your Go services running in Kubernetes using Delve remote debugger.

## 📋 Prerequisites

- Kind cluster running
- Docker Desktop
- VS Code with Go extension
- Delve debugger

## 🚀 Quick Start

### 1. Build Debug Images

```powershell
# Build auth service debug image
docker build -t goconnect-auth:debug -f build\docker\Dockerfile.auth.debug .

# Build gateway debug image
docker build -t goconnect-gateway:debug -f build\docker\Dockerfile.gateway.debug .

# Load into Kind cluster
kind load docker-image goconnect-auth:debug --name goconnect
kind load docker-image goconnect-gateway:debug --name goconnect
```

### 2. Deploy Debug Pods

```powershell
# Deploy debug versions (alongside production pods)
kubectl apply -f deployments\k8s\auth-service-debug.yaml
kubectl apply -f deployments\k8s\gateway-debug.yaml

# Check debug pods are running
kubectl get pods -n goconnect -l app=gateway-debug
kubectl get pods -n goconnect -l app=auth-service-debug
```

### 3. Get Debug Pod Names

```powershell
# Gateway debug pod
$GATEWAY_POD = kubectl get pods -n goconnect -l app=gateway-debug -o jsonpath='{.items[0].metadata.name}'
echo $GATEWAY_POD

# Auth service debug pod
$AUTH_POD = kubectl get pods -n goconnect -l app=auth-service-debug -o jsonpath='{.items[0].metadata.name}'
echo $AUTH_POD
```

## 🔌 Connect Debugger

### Debug Ports

| Service | Debug Port (Host) | Debug Port (Container) |
|---------|-------------------|------------------------|
| Gateway | localhost:30040 | 40000 |
| Auth Service | localhost:30041 | 40000 |

### VS Code Configuration

Create or update `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Gateway (Remote K8s)",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "remotePath": "/root",
      "port": 30040,
      "host": "localhost",
      "showLog": true,
      "trace": "verbose"
    },
    {
      "name": "Debug Auth Service (Remote K8s)",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "remotePath": "/root",
      "port": 30041,
      "host": "localhost",
      "showLog": true,
      "trace": "verbose"
    }
  ]
}
```

### Using VS Code Debugger

1. **Set breakpoints** in your Go code
2. **Press F5** or click "Run and Debug"
3. **Select** "Debug Gateway (Remote K8s)" or "Debug Auth Service (Remote K8s)"
4. **Wait** for "connected" message in Debug Console
5. **Trigger** the code path (make API request)
6. **Debug** when breakpoint hits!

## 🎯 Debugging Workflow

### Test Gateway Debug

```powershell
# Make request to debug gateway
curl http://localhost:30081/health

# Check debug logs
kubectl logs -f deployment/gateway-debug -n goconnect
```

### Test Auth Service Debug

```powershell
# The gateway will call auth service automatically
# Or use grpcurl (if installed):
grpcurl -plaintext localhost:30051 list

# Check debug logs
kubectl logs -f deployment/auth-service-debug -n goconnect
```

## 📝 Manual Port Forward (Alternative)

If NodePort doesn't work, use port-forward:

```powershell
# Gateway debug port
kubectl port-forward -n goconnect deployment/gateway-debug 40000:40000

# Auth service debug port (in another terminal)
kubectl port-forward -n goconnect deployment/auth-service-debug 40001:40000
```

Then update VS Code `launch.json`:
```json
"port": 40000,  // for gateway
"port": 40001,  // for auth-service
```

## 🔍 Debugging Commands

### Check Delve is Running

```powershell
# Gateway
kubectl exec -n goconnect deployment/gateway-debug -- ps aux | findstr dlv

# Auth Service
kubectl exec -n goconnect deployment/auth-service-debug -- ps aux | findstr dlv
```

### View Delve Logs

```powershell
# Gateway Delve logs
kubectl logs -n goconnect deployment/gateway-debug -f

# Auth Service Delve logs
kubectl logs -n goconnect deployment/auth-service-debug -f
```

### Test Debug Port is Open

```powershell
# Test from local machine
Test-NetConnection -ComputerName localhost -Port 30040
Test-NetConnection -ComputerName localhost -Port 30041
```

## 🛠️ Troubleshooting

### Debugger Won't Connect

1. **Check pod is running:**
   ```powershell
   kubectl get pods -n goconnect | findstr debug
   ```

2. **Check debug port is exposed:**
   ```powershell
   kubectl get svc -n goconnect | findstr debug
   ```

3. **Check Delve is listening:**
   ```powershell
   kubectl logs deployment/gateway-debug -n goconnect | findstr listening
   ```

4. **Verify port mapping in Kind:**
   ```powershell
   docker ps | findstr goconnect
   ```

### Breakpoints Not Hitting

1. **Verify source paths match:**
   - Remote path: `/root` (in container)
   - Local path: Your project root

2. **Rebuild with debug flags:**
   ```powershell
   # Ensure -gcflags="all=-N -l" is in Dockerfile
   docker build -t goconnect-gateway:debug -f build\docker\Dockerfile.gateway.debug .
   ```

3. **Check code is being executed:**
   ```powershell
   # Add log statements to verify
   kubectl logs -f deployment/gateway-debug -n goconnect
   ```

### Pod Keeps Crashing

```powershell
# Check pod events
kubectl describe pod -n goconnect -l app=gateway-debug

# Check previous logs
kubectl logs -n goconnect -l app=gateway-debug --previous

# Check security context
kubectl get pod -n goconnect -l app=gateway-debug -o yaml | findstr securityContext
```

## 📊 Debug vs Production

| Aspect | Production | Debug |
|--------|-----------|-------|
| Replicas | 2+ | 1 |
| Optimizations | Yes | No (-N -l) |
| Delve | No | Yes |
| Debug Port | No | 40000 |
| Resources | Normal | Higher |
| Security | Standard | SYS_PTRACE |

## 🔄 Switching Between Modes

### Use Debug Version

```powershell
# Scale down production
kubectl scale deployment gateway --replicas=0 -n goconnect

# Deploy debug
kubectl apply -f deployments\k8s\gateway-debug.yaml

# Update ingress to point to debug service (optional)
```

### Back to Production

```powershell
# Delete debug deployment
kubectl delete -f deployments\k8s\gateway-debug.yaml

# Scale up production
kubectl scale deployment gateway --replicas=2 -n goconnect
```

## 💡 Tips

1. **Single Replica**: Debug deployments use 1 replica for easier debugging
2. **SYS_PTRACE**: Required capability for Delve to attach
3. **Source Maps**: Ensure your local code matches the deployed version
4. **Logs**: Always check pod logs if debugger won't connect
5. **Rebuild**: After code changes, rebuild and reload the image

## 🎓 Advanced: Debug with dlv CLI

```powershell
# Connect via dlv command line
dlv connect localhost:30040

# In dlv prompt:
(dlv) break main.main
(dlv) continue
(dlv) next
(dlv) print variableName
(dlv) quit
```

## 📚 Resources

- [Delve Documentation](https://github.com/go-delve/delve)
- [VS Code Go Debugging](https://github.com/golang/vscode-go/wiki/debugging)
- [Kubernetes Debug Containers](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/)
