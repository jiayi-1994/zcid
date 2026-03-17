#!/bin/bash
set -e
NAMESPACE=zcid

# 镜像代理设置（中国网络环境设置 USE_PROXY=1）
PROXY=${USE_PROXY:+docker.gh-proxy.com/}

# Bitnami 镜像仓库（2025-08 起 bitnami 已迁移至 bitnamilegacy，不再更新）
# 如需使用付费 Bitnami Secure Images，将此值改为 bitnami
BITNAMI_REPO=${BITNAMI_REPO:-bitnamilegacy}

echo "===== Step 1: Namespace ====="
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

echo "===== Step 2: Middleware (PostgreSQL + Redis + MinIO) ====="
echo "  (使用镜像仓库: docker.io/${BITNAMI_REPO})"
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm upgrade --install postgresql bitnami/postgresql -n $NAMESPACE \
  --set global.imageRegistry=docker.io \
  --set image.registry=docker.io --set image.repository=${BITNAMI_REPO}/postgresql \
  --set auth.database=zcicd --set auth.username=postgres \
  --set auth.postgresPassword=zcicd123 \
  --set primary.persistence.size=10Gi --wait --timeout=300s

helm upgrade --install redis bitnami/redis -n $NAMESPACE \
  --set global.imageRegistry=docker.io \
  --set image.registry=docker.io --set image.repository=${BITNAMI_REPO}/redis \
  --set architecture=standalone --set auth.enabled=false \
  --set master.persistence.size=1Gi --wait --timeout=300s

helm upgrade --install minio bitnami/minio -n $NAMESPACE \
  --set global.imageRegistry=docker.io \
  --set image.registry=docker.io --set image.repository=${BITNAMI_REPO}/minio \
  --set auth.rootUser=minioadmin --set auth.rootPassword=zcicd123 \
  --set defaultBuckets=zcid-logs --set persistence.size=10Gi --wait --timeout=300s

echo "===== Step 3: Tekton ====="
if [ -n "$USE_PROXY" ]; then
  echo "  (使用 ghcr 镜像 + 代理)"
  sed "s|ghcr.io|${PROXY}ghcr.io|g" deploy/tekton/tekton-v0.62.0-ghcr.yaml | kubectl apply -f -
else
  kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.62.0/release.yaml
fi
kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=tekton-pipelines \
  -n tekton-pipelines --timeout=300s
kubectl apply -f deploy/tekton/zcid-tekton-rbac.yaml

echo "===== Step 4: ArgoCD ====="
if [ -n "$USE_PROXY" ]; then
  echo "  (使用镜像代理)"
  curl -sSL https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml \
    | sed "s|quay.io|${PROXY}quay.io|g" \
    | sed "s|ghcr.io|${PROXY}ghcr.io|g" \
    | kubectl apply -n argocd -f -
else
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml
fi
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server \
  -n argocd --timeout=300s
kubectl apply -f deploy/argocd/zcid-argocd-project.yaml

echo "===== Step 5: zcid ====="
ZCID_IMAGE="ghcr.io/jiayi-1994/zcid"
[ -n "$USE_PROXY" ] && ZCID_IMAGE="${PROXY}${ZCID_IMAGE}"

helm upgrade --install zcid deploy/helm/zcid/ -n $NAMESPACE \
  --set image.repository=$ZCID_IMAGE \
  --set secrets.dbPassword=zcicd123 \
  --set secrets.redisPassword="" \
  --set secrets.minioSecretKey=zcicd123 \
  --set secrets.jwtSecret=zcid-jwt-test-$(date +%s) \
  --set config.encryptionKey=0123456789abcdef0123456789abcdef \
  --wait --timeout=300s

echo "===== Done ====="
echo ""
kubectl get pods -n $NAMESPACE
echo ""
echo "访问方式:"
echo "  kubectl port-forward svc/zcid -n $NAMESPACE 8080:8080"
echo "  浏览器打开 http://localhost:8080"
echo "  账号: admin / admin123"
