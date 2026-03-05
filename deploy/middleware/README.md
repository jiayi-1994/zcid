# zcid 中间件部署方案

zcid 依赖三个中间件：**PostgreSQL**、**Redis**、**MinIO**。以下提供基于 Helm 的 K8s 部署方案。

## 前置条件

- Kubernetes 集群 1.25+
- Helm 3.10+
- 可用的 StorageClass（用于 PVC 持久化）

## 1. PostgreSQL

使用 Bitnami PostgreSQL Helm Chart：

```bash
# 添加 Bitnami 仓库
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# 创建 namespace
kubectl create namespace zcid

# 安装 PostgreSQL
helm install postgresql bitnami/postgresql \
  --namespace zcid \
  --set auth.postgresPassword=admin-password \
  --set auth.database=zcid \
  --set auth.username=zcid \
  --set auth.password=zcid-password \
  --set primary.persistence.size=20Gi \
  --set primary.resources.requests.memory=256Mi \
  --set primary.resources.requests.cpu=250m
```

### 验证

```bash
kubectl get pods -n zcid -l app.kubernetes.io/name=postgresql
kubectl exec -it postgresql-0 -n zcid -- psql -U zcid -d zcid -c "SELECT version();"
```

### 连接信息

| 配置项 | 值 |
|--------|-----|
| Host | `postgresql.zcid.svc.cluster.local` |
| Port | `5432` |
| Database | `zcid` |
| Username | `zcid` |
| Password | `zcid-password` |

---

## 2. Redis

使用 Bitnami Redis Helm Chart：

```bash
helm install redis bitnami/redis \
  --namespace zcid \
  --set architecture=standalone \
  --set auth.enabled=false \
  --set master.persistence.size=2Gi \
  --set master.resources.requests.memory=128Mi \
  --set master.resources.requests.cpu=100m
```

> 生产环境建议启用认证：`--set auth.enabled=true --set auth.password=your-redis-password`

### 验证

```bash
kubectl get pods -n zcid -l app.kubernetes.io/name=redis
kubectl exec -it redis-master-0 -n zcid -- redis-cli ping
```

### 连接信息

| 配置项 | 值 |
|--------|-----|
| Host | `redis-master.zcid.svc.cluster.local` |
| Port | `6379` |
| Password | （空，或自定义） |

---

## 3. MinIO

使用 Bitnami MinIO Helm Chart：

```bash
helm install minio bitnami/minio \
  --namespace zcid \
  --set auth.rootUser=minioadmin \
  --set auth.rootPassword=minioadmin \
  --set defaultBuckets=zcid-logs \
  --set persistence.size=50Gi \
  --set resources.requests.memory=256Mi \
  --set resources.requests.cpu=250m
```

### 验证

```bash
kubectl get pods -n zcid -l app.kubernetes.io/name=minio

# 端口转发访问 MinIO Console
kubectl port-forward svc/minio -n zcid 9001:9001
# 浏览器访问 http://localhost:9001 (minioadmin / minioadmin)
```

### 连接信息

| 配置项 | 值 |
|--------|-----|
| Endpoint | `minio.zcid.svc.cluster.local:9000` |
| Bucket | `zcid-logs` |
| Access Key | `minioadmin` |
| Secret Key | `minioadmin` |
| Use SSL | `false` |

---

## 4. 一键部署脚本

```bash
#!/bin/bash
set -e

NAMESPACE=zcid

echo "=== Creating namespace ==="
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

echo "=== Adding Bitnami repo ==="
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

echo "=== Installing PostgreSQL ==="
helm upgrade --install postgresql bitnami/postgresql \
  --namespace $NAMESPACE \
  --set auth.postgresPassword=admin-password \
  --set auth.database=zcid \
  --set auth.username=zcid \
  --set auth.password=zcid-password \
  --set primary.persistence.size=20Gi \
  --wait --timeout=300s

echo "=== Installing Redis ==="
helm upgrade --install redis bitnami/redis \
  --namespace $NAMESPACE \
  --set architecture=standalone \
  --set auth.enabled=false \
  --set master.persistence.size=2Gi \
  --wait --timeout=300s

echo "=== Installing MinIO ==="
helm upgrade --install minio bitnami/minio \
  --namespace $NAMESPACE \
  --set auth.rootUser=minioadmin \
  --set auth.rootPassword=minioadmin \
  --set defaultBuckets=zcid-logs \
  --set persistence.size=50Gi \
  --wait --timeout=300s

echo "=== All middleware deployed ==="
kubectl get pods -n $NAMESPACE
```

---

## 5. 资源估算

| 组件 | CPU Request | Memory Request | 磁盘 |
|------|------------|----------------|------|
| PostgreSQL | 250m | 256Mi | 20Gi |
| Redis | 100m | 128Mi | 2Gi |
| MinIO | 250m | 256Mi | 50Gi |
| **合计** | **600m** | **640Mi** | **72Gi** |

---

## 6. 生产环境建议

1. **PostgreSQL**: 启用 `replication.enabled=true`，配置 WAL 备份到 S3/MinIO
2. **Redis**: 启用认证，考虑 Sentinel 架构
3. **MinIO**: 启用 distributed 模式（4+ 节点），配置 TLS
4. **所有组件**: 设置 `PodDisruptionBudget`，配置 `anti-affinity`
5. **备份**: 使用 Velero 定期备份 PVC
