#Requires -Version 5.1
<#
.SYNOPSIS
  Deploy toàn bộ Todo App lên minikube theo đúng thứ tự.
.NOTES
  Chạy từ root của project: .\k8s\deploy-all.ps1
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-Step($n, $total, $msg) {
    Write-Host "`n=== [$n/$total] $msg ===" -ForegroundColor Cyan
}

# ── Bước 1: Build images trong context minikube ──────────────────────────────
Write-Step 1 7 "Build Docker images trong minikube context"
& minikube -p minikube docker-env --shell powershell | Invoke-Expression
docker build -t todo-bff:latest            -f services/todo-bff/Dockerfile .
docker build -t todo-users:latest          -f services/users/Dockerfile .
docker build -t todo-users-worker:latest   -f services/users/Dockerfile.worker .
docker build -t todo-todos:latest          -f services/todos/Dockerfile .
docker build -t todo-todos-consumer:latest -f services/todos/Dockerfile.consumer .

# ── Bước 2: Namespace ─────────────────────────────────────────────────────────
Write-Step 2 7 "Namespace"
kubectl apply -f k8s/namespace.yaml

# ── Bước 3: Secrets & ConfigMaps ─────────────────────────────────────────────
Write-Step 3 7 "Secrets & ConfigMaps"
kubectl apply -f k8s/secrets/
kubectl apply -f k8s/configmaps/

# ── Bước 4: Storage (PVCs) ───────────────────────────────────────────────────
Write-Step 4 7 "Storage (PVCs)"
kubectl apply -f k8s/storage/

# ── Bước 5: Infrastructure ───────────────────────────────────────────────────
Write-Step 5 7 "Infrastructure (Postgres, RabbitMQ, Redis)"
kubectl apply -f k8s/infrastructure/

Write-Host "`nChờ infrastructure sẵn sàng..." -ForegroundColor Yellow
kubectl rollout status statefulset/postgres-todo  -n todo-app --timeout=120s
kubectl rollout status statefulset/postgres-users -n todo-app --timeout=120s
kubectl rollout status statefulset/rabbitmq        -n todo-app --timeout=120s
kubectl rollout status statefulset/redis           -n todo-app --timeout=120s

# ── Bước 6: Application Services ─────────────────────────────────────────────
Write-Step 6 7 "Application Services"
kubectl apply -f k8s/services/

# ── Bước 7: Ingress ───────────────────────────────────────────────────────────
Write-Step 7 7 "Ingress"
kubectl apply -f k8s/ingress.yaml

# ── Done ──────────────────────────────────────────────────────────────────────
Write-Host "`n=== DONE! ===" -ForegroundColor Green
kubectl get all -n todo-app
$bffUrl = minikube service todo-bff-svc -n todo-app --url
Write-Host "`nBFF URL: $bffUrl" -ForegroundColor Green
