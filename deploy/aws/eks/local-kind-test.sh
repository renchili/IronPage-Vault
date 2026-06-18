#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-ironpage-vault}"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind is required for local EKS simulation" >&2
  exit 2
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required for local EKS simulation" >&2
  exit 2
fi

if ! kind get clusters | grep -qx "$CLUSTER_NAME"; then
  kind create cluster --name "$CLUSTER_NAME"
fi

docker build -t ironpage-vault:local .
kind load docker-image ironpage-vault:local --name "$CLUSTER_NAME"

kubectl apply -k deploy/aws/eks
kubectl -n ironpage rollout status deployment/ironpage-vault --timeout=180s
kubectl -n ironpage get pods,svc

echo "Run this to test the API:"
echo "kubectl -n ironpage port-forward svc/ironpage-vault 8080:8080"
echo "curl http://127.0.0.1:8080/healthz"
