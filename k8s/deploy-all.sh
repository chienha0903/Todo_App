#!/usr/bin/env bash
set -euo pipefail

write_step() {
  local n="$1"
  local total="$2"
  local msg="$3"
  echo ""
  echo "=== [$n/$total] $msg ==="
}

write_step 1 8 "Build Docker images trong minikube context"
eval "$(minikube -p minikube docker-env)"

docker build -t todo-bff:latest            -f services/todo-bff/Dockerfile .
docker build -t todo-users:latest          -f services/users/Dockerfile .
docker build -t todo-users-worker:latest   -f services/users/Dockerfile.worker .
docker build -t todo-todos:latest          -f services/todos/Dockerfile .
docker build -t todo-todos-consumer:latest -f services/todos/Dockerfile.consumer .

write_step 2 8 "Namespace"
kubectl apply -f k8s/namespace.yaml

write_step 3 8 "Secrets & ConfigMaps"
kubectl apply -f k8s/secrets/
kubectl apply -f k8s/configmaps/

write_step 4 8 "Storage PVCs"
kubectl apply -f k8s/storage/

write_step 5 8 "Infrastructure: Postgres, RabbitMQ, Redis"
kubectl apply -f k8s/infrastructure/

echo ""
echo "Chờ infrastructure sẵn sàng..."
kubectl rollout status statefulset/postgres-todo  -n todo-app --timeout=120s
kubectl rollout status statefulset/postgres-users -n todo-app --timeout=120s
kubectl rollout status statefulset/rabbitmq        -n todo-app --timeout=120s
kubectl rollout status statefulset/redis           -n todo-app --timeout=120s

write_step 6 8 "Application Services"
kubectl apply -f k8s/services/

write step 7 8 "Network Policies"
kubectl apply -f k8s/network-policies/

write_step 8 8 "Ingress"
kubectl apply -f k8s/ingress.yaml

echo ""
echo "=== DONE! ==="
kubectl get all -n todo-app

echo ""
echo "BFF URL:"
minikube service todo-bff-svc -n todo-app --url