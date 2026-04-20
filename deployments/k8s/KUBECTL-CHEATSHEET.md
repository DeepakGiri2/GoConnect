# Kubernetes Debugging Cheat Sheet

## 📋 Quick Status Checks

### Get All Pods
```powershell
# All pods in all namespaces
kubectl get pods -A

# All pods in goconnect namespace
kubectl get pods -n goconnect

# With more details (IP, node, status)
kubectl get pods -n goconnect -o wide

# Watch pods in real-time
kubectl get pods -n goconnect -w
```

### Get Pod Details
```powershell
# Describe specific pod (shows events, status, conditions)
kubectl describe pod <pod-name> -n goconnect

# Get pod YAML
kubectl get pod <pod-name> -n goconnect -o yaml

# Get pod JSON
kubectl get pod <pod-name> -n goconnect -o json

# Get pod with labels
kubectl get pods -n goconnect --show-labels
```

## 📝 Logs Commands

### View Logs
```powershell
# Gateway logs (follow mode)
kubectl logs -f deployment/gateway -n goconnect

# Auth service logs (follow mode)
kubectl logs -f deployment/auth-service -n goconnect

# Specific pod logs
kubectl logs <pod-name> -n goconnect

# Previous container logs (if crashed)
kubectl logs <pod-name> -n goconnect --previous

# Last 100 lines
kubectl logs <pod-name> -n goconnect --tail=100

# Logs since 1 hour ago
kubectl logs <pod-name> -n goconnect --since=1h

# Logs with timestamps
kubectl logs <pod-name> -n goconnect --timestamps
```

### Multi-Container Pods
```powershell
# List containers in a pod
kubectl get pod <pod-name> -n goconnect -o jsonpath='{.spec.containers[*].name}'

# Logs from specific container
kubectl logs <pod-name> -c <container-name> -n goconnect

# All containers in a pod
kubectl logs <pod-name> -n goconnect --all-containers=true
```

### Stream All Logs
```powershell
# All gateway pod logs
kubectl logs -l app=gateway -n goconnect --all-containers=true -f

# All auth-service pod logs
kubectl logs -l app=auth-service -n goconnect --all-containers=true -f
```

## 🔍 Container Information

### Get Container Names
```powershell
# All containers in goconnect namespace
kubectl get pods -n goconnect -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{range .spec.containers[*]}{"\t"}{.name}{"\n"}{end}{end}'

# Specific pod containers
kubectl get pod <pod-name> -n goconnect -o jsonpath='{.spec.containers[*].name}' && echo
```

### Container Status
```powershell
# Container ready status
kubectl get pods -n goconnect -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{range .status.containerStatuses[*]}{.name}={.ready}{"\t"}{end}{"\n"}{end}'

# Container restart counts
kubectl get pods -n goconnect -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{range .status.containerStatuses[*]}{.restartCount}{"\t"}{end}{"\n"}{end}'
```

## 🐚 Execute Commands in Pods

### Shell Access
```powershell
# Execute bash in pod
kubectl exec -it <pod-name> -n goconnect -- /bin/sh

# Execute in specific container
kubectl exec -it <pod-name> -c <container-name> -n goconnect -- /bin/sh

# Run single command
kubectl exec <pod-name> -n goconnect -- ls -la

# PostgreSQL access
kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect

# Redis access
kubectl exec -it redis-0 -n goconnect -- redis-cli
```

## 📊 Service & Networking

### Services
```powershell
# All services
kubectl get svc -n goconnect

# Service details
kubectl describe svc gateway-service -n goconnect

# Endpoints (actual pod IPs)
kubectl get endpoints -n goconnect
```

### Ingress
```powershell
# Get ingress
kubectl get ingress -n goconnect

# Ingress details
kubectl describe ingress goconnect-ingress -n goconnect
```

### Port Forwarding (Alternative Access)
```powershell
# Forward gateway locally
kubectl port-forward -n goconnect svc/gateway-service 8080:80

# Forward specific pod
kubectl port-forward -n goconnect <pod-name> 8080:8080

# Forward Prometheus
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# Forward Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000
```

## 🔧 Debugging Commands

### Events
```powershell
# All events in namespace
kubectl get events -n goconnect --sort-by=.metadata.creationTimestamp

# Watch events
kubectl get events -n goconnect -w

# Events for specific pod
kubectl get events -n goconnect --field-selector involvedObject.name=<pod-name>
```

### Resource Usage
```powershell
# Top pods (CPU/Memory)
kubectl top pods -n goconnect

# Top nodes
kubectl top nodes

# Specific pod resources
kubectl top pod <pod-name> -n goconnect
```

### Deployments
```powershell
# Get deployments
kubectl get deployments -n goconnect

# Deployment status
kubectl rollout status deployment/gateway -n goconnect

# Deployment history
kubectl rollout history deployment/gateway -n goconnect

# Scale deployment
kubectl scale deployment gateway --replicas=3 -n goconnect
```

### StatefulSets
```powershell
# Get statefulsets
kubectl get statefulsets -n goconnect

# StatefulSet status
kubectl get sts postgres -n goconnect

# StatefulSet pods
kubectl get pods -l app=postgres -n goconnect
```

## 🔄 Common Operations

### Restart Pods
```powershell
# Restart deployment (rolling restart)
kubectl rollout restart deployment/gateway -n goconnect
kubectl rollout restart deployment/auth-service -n goconnect

# Delete pod (will auto-recreate)
kubectl delete pod <pod-name> -n goconnect
```

### Update Configuration
```powershell
# Edit deployment
kubectl edit deployment gateway -n goconnect

# Apply changes from file
kubectl apply -f deployments/k8s/gateway.yaml

# Replace resource
kubectl replace -f deployments/k8s/gateway.yaml
```

### Copy Files
```powershell
# Copy from pod to local
kubectl cp goconnect/<pod-name>:/path/to/file ./local-file

# Copy from local to pod
kubectl cp ./local-file goconnect/<pod-name>:/path/to/file
```

## 🎯 Quick Pod Names

### Get Pod Names Quickly
```powershell
# Gateway pods
kubectl get pods -n goconnect -l app=gateway -o jsonpath='{.items[0].metadata.name}'

# Auth service pods
kubectl get pods -n goconnect -l app=auth-service -o jsonpath='{.items[0].metadata.name}'

# Postgres pod
kubectl get pods -n goconnect -l app=postgres -o jsonpath='{.items[0].metadata.name}'

# Redis pod
kubectl get pods -n goconnect -l app=redis -o jsonpath='{.items[0].metadata.name}'

# All pod names in namespace
kubectl get pods -n goconnect -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}'
```

## 🚨 Troubleshooting

### Pod Not Starting
```powershell
# Check pod status and events
kubectl describe pod <pod-name> -n goconnect

# Check logs for errors
kubectl logs <pod-name> -n goconnect

# Check previous logs if crashed
kubectl logs <pod-name> -n goconnect --previous
```

### Service Not Reachable
```powershell
# Check service endpoints
kubectl get endpoints gateway-service -n goconnect

# Test from another pod
kubectl run -it --rm debug --image=busybox --restart=Never -n goconnect -- wget -O- gateway-service

# Check ingress
kubectl describe ingress goconnect-ingress -n goconnect
```

### Database Issues
```powershell
# Check postgres logs
kubectl logs postgres-0 -n goconnect

# Connect to postgres
kubectl exec -it postgres-0 -n goconnect -- psql -U postgres -d goconnect

# Run SQL query
kubectl exec postgres-0 -n goconnect -- psql -U postgres -d goconnect -c "SELECT * FROM users LIMIT 5;"
```

## 📦 Monitoring Stack

### Prometheus
```powershell
# Prometheus pods
kubectl get pods -n monitoring -l app=prometheus

# Prometheus logs
kubectl logs -f deployment/prometheus -n monitoring

# Access: http://localhost:9090
```

### Grafana
```powershell
# Grafana pods
kubectl get pods -n monitoring -l app=grafana

# Grafana logs
kubectl logs -f deployment/grafana -n monitoring

# Access: http://localhost:3000 (admin/admin123)
```

### Traefik
```powershell
# Traefik pods
kubectl get pods -n traefik -l app=traefik

# Traefik logs
kubectl logs -f deployment/traefik -n traefik

# Dashboard: http://localhost:8888
```

## 💡 Useful Aliases (Add to PowerShell Profile)

```powershell
# Add to $PROFILE
function k { kubectl $args }
function kgp { kubectl get pods -n goconnect $args }
function kgpa { kubectl get pods -A $args }
function kdp { kubectl describe pod $args -n goconnect }
function kl { kubectl logs -f $args -n goconnect }
function kex { kubectl exec -it $args -n goconnect -- /bin/sh }
function kpf { kubectl port-forward -n goconnect $args }

# Reload profile: . $PROFILE
```

## 🔍 Complete Cluster Overview

```powershell
# Everything in one command
kubectl get all -n goconnect

# All resources across all namespaces
kubectl get all -A

# Cluster info
kubectl cluster-info

# Node status
kubectl get nodes -o wide
```
