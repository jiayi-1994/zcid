#!/bin/bash
set -e
NAMESPACE=zcid

# 镜像代理设置（中国网络环境设置 USE_PROXY=1）
PROXY=${USE_PROXY:+docker.gh-proxy.com/}

# Bitnami 镜像仓库（2025-08 起 bitnami 已迁移至 bitnamilegacy，不再更新）
# 如需使用付费 Bitnami Secure Images，将此值改为 bitnami
BITNAMI_REPO=${BITNAMI_REPO:-bitnamilegacy}

# StorageClass 设置
#   未设置 = 自动探测（优先默认 SC，否则使用集群唯一 SC）
#   设为具体值 = 使用指定 StorageClass（如 edge-lvm, local-path, standard 等）
STORAGE_CLASS=${STORAGE_CLASS:-}

# 持久化存储设置
#   PERSISTENCE=true  (默认) 使用 PVC 持久化数据
#   PERSISTENCE=false 使用 emptyDir，无需 StorageClass/PV（适合测试环境）
PERSISTENCE=${PERSISTENCE:-true}

# 自动探测 StorageClass（仅当 PERSISTENCE=true 且未显式指定 STORAGE_CLASS 时）
if [ "$PERSISTENCE" != "false" ] && [ -z "$STORAGE_CLASS" ]; then
  DEFAULT_SC=$(kubectl get storageclass -o jsonpath='{range .items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")]}{.metadata.name}{end}' 2>/dev/null)
  if [ -n "$DEFAULT_SC" ]; then
    STORAGE_CLASS="$DEFAULT_SC"
    echo "  自动探测到默认 StorageClass: ${STORAGE_CLASS}"
  else
    SC_LIST=$(kubectl get storageclass -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
    SC_COUNT=$(echo "$SC_LIST" | wc -w)
    if [ "$SC_COUNT" -eq 1 ]; then
      STORAGE_CLASS="$SC_LIST"
      echo "  集群仅有一个 StorageClass，自动使用: ${STORAGE_CLASS}"
    elif [ "$SC_COUNT" -eq 0 ]; then
      echo "  ⚠ 警告: 集群中没有 StorageClass，PVC 可能无法绑定"
      echo "  建议: 设置 PERSISTENCE=false 或安装存储供给器"
    else
      echo "  ⚠ 警告: 集群中有多个 StorageClass 但无默认值:"
      echo "    $SC_LIST"
      echo "  建议: 通过 STORAGE_CLASS=<name> 显式指定，否则 PVC 可能无法绑定"
    fi
  fi
fi

echo "===== Step 1: Namespace ====="
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

echo "===== Step 2: Middleware (PostgreSQL + Redis + MinIO) ====="
echo "  (镜像仓库: docker.io/${BITNAMI_REPO})"

# 构建 StorageClass 参数
STORAGE_ARGS=""
if [ -n "$STORAGE_CLASS" ]; then
  STORAGE_ARGS="--set global.storageClass=${STORAGE_CLASS}"
  echo "  StorageClass: ${STORAGE_CLASS}"
else
  echo "  StorageClass: (未指定)"
fi

# 构建持久化参数
if [ "$PERSISTENCE" = "false" ]; then
  echo "  持久化: 已禁用 (emptyDir)"
  PG_PERSIST_ARGS="--set primary.persistence.enabled=false"
  REDIS_PERSIST_ARGS="--set master.persistence.enabled=false"
  MINIO_PERSIST_ARGS="--set persistence.enabled=false"
else
  echo "  持久化: 已启用 (PVC)"
  PG_PERSIST_ARGS="--set primary.persistence.size=10Gi"
  REDIS_PERSIST_ARGS="--set master.persistence.size=1Gi"
  MINIO_PERSIST_ARGS="--set persistence.size=10Gi"
fi

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm upgrade --install postgresql bitnami/postgresql -n $NAMESPACE \
  --set global.imageRegistry=docker.io \
  --set image.registry=docker.io --set image.repository=${BITNAMI_REPO}/postgresql \
  $STORAGE_ARGS \
  --set auth.database=zcicd --set auth.username=postgres \
  --set auth.postgresPassword=zcicd123 \
  $PG_PERSIST_ARGS --wait --timeout=300s

helm upgrade --install redis bitnami/redis -n $NAMESPACE \
  --set global.imageRegistry=docker.io \
  --set image.registry=docker.io --set image.repository=${BITNAMI_REPO}/redis \
  $STORAGE_ARGS \
  --set architecture=standalone --set auth.enabled=false \
  $REDIS_PERSIST_ARGS --wait --timeout=300s

helm upgrade --install minio bitnami/minio -n $NAMESPACE \
  --set global.imageRegistry=docker.io \
  --set image.registry=docker.io --set image.repository=${BITNAMI_REPO}/minio \
  $STORAGE_ARGS \
  --set auth.rootUser=minioadmin --set auth.rootPassword=zcicd123 \
  --set defaultBuckets=zcid-logs $MINIO_PERSIST_ARGS --wait --timeout=300s

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
