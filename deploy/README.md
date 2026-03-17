# zcid 部署指南

完整的 Kubernetes 测试/生产环境部署方案。

## 目录

- [环境要求](#环境要求)
- [部署架构](#部署架构)
- [快速部署（一键脚本）](#快速部署)
- [分步部署](#分步部署)
  - [Step 1: 创建 Namespace](#step-1-创建-namespace)
  - [Step 2: 部署中间件](#step-2-部署中间件)
  - [Step 3: 部署 Tekton](#step-3-部署-tekton)
  - [Step 4: 部署 ArgoCD](#step-4-部署-argocd)
  - [Step 5: 部署 zcid](#step-5-部署-zcid)
  - [Step 6: 验证](#step-6-验证)
- [Helm Values 参考](#helm-values-参考)
- [常见问题](#常见问题)
- [资源估算](#资源估算)
- [生产环境建议](#生产环境建议)

---

## 环境要求

| 组件 | 最低版本 | 用途 |
|------|---------|------|
| Kubernetes | 1.25+ | 容器编排 |
| Helm | 3.10+ | 包管理 |
| kubectl | 1.25+ | 集群操作 |
| StorageClass | 任意 | PVC 持久化 |

**集群最低资源**: 4 CPU / 4Gi Memory / 80Gi Disk

---

## 部署架构

```
                    ┌─────────────────────────────────┐
                    │        K8s Cluster               │
                    │                                   │
┌─────────┐        │  ┌──────────────────────────┐    │
│ Browser  │───────▶│  │   zcid (Helm Chart)      │    │
│          │        │  │   Pod: zcid-server        │    │
└─────────┘        │  │   Port: 8080              │    │
                    │  └──────┬───────────────────┘    │
                    │         │                         │
         ┌─────────┼─────────┼──────────┐              │
         ▼         ▼         ▼          ▼              │
  ┌────────┐ ┌───────┐ ┌───────┐ ┌──────────┐         │
  │ PG 16  │ │Redis 7│ │ MinIO │ │Tekton    │         │
  │ (data) │ │(cache)│ │(logs) │ │Pipelines │         │
  └────────┘ └───────┘ └───────┘ └──────────┘         │
                    │                                   │
                    │  ┌──────────────────────────┐    │
                    │  │  ArgoCD (argocd ns)       │    │
                    │  │  CD engine                │    │
                    │  └──────────────────────────┘    │
                    └─────────────────────────────────┘
```

所有组件部署在同一个 K8s 集群内。zcid 通过 ServiceAccount 访问 Tekton CRD，通过 REST API 访问 ArgoCD。**Helm chart 自动配置所有内部连接地址，无需手动设置环境变量。**

---

## 快速部署

适用于测试环境的一键部署脚本：

```bash
#!/bin/bash
set -e
NAMESPACE=zcid

echo "===== Step 1: Namespace ====="
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

echo "===== Step 2: Middleware (PostgreSQL + Redis + MinIO) ====="
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm upgrade --install postgresql bitnami/postgresql -n $NAMESPACE \
  --set auth.database=zcicd --set auth.username=postgres \
  --set auth.postgresPassword=zcicd123 \
  --set primary.persistence.size=10Gi --wait --timeout=300s

helm upgrade --install redis bitnami/redis -n $NAMESPACE \
  --set architecture=standalone --set auth.enabled=false \
  --set master.persistence.size=1Gi --wait --timeout=300s

helm upgrade --install minio bitnami/minio -n $NAMESPACE \
  --set auth.rootUser=minioadmin --set auth.rootPassword=zcicd123 \
  --set defaultBuckets=zcid-logs --set persistence.size=10Gi --wait --timeout=300s

echo "===== Step 3: Tekton ====="
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.62.0/release.yaml
kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=tekton-pipelines \
  -n tekton-pipelines --timeout=300s
kubectl apply -f deploy/tekton/zcid-tekton-rbac.yaml

echo "===== Step 4: ArgoCD ====="
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server \
  -n argocd --timeout=300s
kubectl apply -f deploy/argocd/zcid-argocd-project.yaml

echo "===== Step 5: zcid ====="
helm upgrade --install zcid deploy/helm/zcid/ -n $NAMESPACE \
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
```

保存为 `deploy/install.sh`，执行 `bash deploy/install.sh` 即可。

---

## 分步部署

### Step 1: 创建 Namespace

```bash
kubectl create namespace zcid
kubectl create namespace argocd
```

### Step 2: 部署中间件

详细文档：[`middleware/README.md`](middleware/README.md)

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update

# PostgreSQL
helm install postgresql bitnami/postgresql -n zcid \
  --set auth.database=zcicd \
  --set auth.username=postgres \
  --set auth.postgresPassword=<your-password> \
  --set primary.persistence.size=20Gi

# Redis
helm install redis bitnami/redis -n zcid \
  --set architecture=standalone --set auth.enabled=false

# MinIO
helm install minio bitnami/minio -n zcid \
  --set auth.rootUser=minioadmin \
  --set auth.rootPassword=<your-password> \
  --set defaultBuckets=zcid-logs
```

**验证中间件就绪：**

```bash
kubectl get pods -n zcid
# 确保 postgresql-0, redis-master-0, minio-xxx 都是 Running
```

### Step 3: 部署 Tekton

详细文档：[`tekton/README.md`](tekton/README.md)

```bash
# 安装 Tekton Pipelines v0.62
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.62.0/release.yaml
kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=tekton-pipelines \
  -n tekton-pipelines --timeout=300s

# 创建 zcid RBAC（允许 zcid 管理 PipelineRun）
kubectl apply -f deploy/tekton/zcid-tekton-rbac.yaml

# 配置 Tekton（可选）
kubectl patch configmap feature-flags -n tekton-pipelines \
  --type merge -p '{"data":{"disable-affinity-assistant":"true"}}'
```

### Step 4: 部署 ArgoCD

详细文档：[`argocd/README.md`](argocd/README.md)

```bash
# 安装 ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server \
  -n argocd --timeout=300s

# 创建 zcid 专用 Project
kubectl apply -f deploy/argocd/zcid-argocd-project.yaml

# 获取 admin 密码（用于生成 zcid API token）
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d; echo
```

**生成 zcid API Token（推荐）：**

```bash
# 配置 zcid 账号
kubectl patch configmap argocd-cm -n argocd --type merge -p '{
  "data": {"accounts.zcid": "apiKey, login", "accounts.zcid.enabled": "true"}
}'
kubectl patch configmap argocd-rbac-cm -n argocd --type merge -p '{
  "data": {"policy.csv": "p, role:zcid-role, applications, *, */*, allow\np, role:zcid-role, projects, get, *, allow\ng, zcid, role:zcid-role"}
}'
kubectl rollout restart deployment argocd-server -n argocd

# 生成 token（需要先安装 argocd CLI）
argocd login <argocd-server>:443 --username admin --password <admin-password> --insecure
argocd account generate-token --account zcid
# 将输出的 token 配置到 Step 5 的 secrets.argocdToken
```

### Step 5: 部署 zcid

```bash
helm install zcid deploy/helm/zcid/ --namespace zcid \
  --set secrets.dbPassword=<postgresql-password> \
  --set secrets.redisPassword=<redis-password-if-enabled> \
  --set secrets.minioSecretKey=<minio-password> \
  --set secrets.jwtSecret=<random-jwt-signing-key> \
  --set secrets.argocdToken=<argocd-api-token> \
  --set config.encryptionKey=<32-byte-hex-key>
```

**Helm chart 自动处理的事项（不需要手动配置）：**

| 配置 | 说明 | Helm 默认值 |
|------|------|------------|
| K8s 集成 | Pod 在集群内运行，通过 ServiceAccount 访问 K8s API | `k8sEnabled: true` |
| ArgoCD 地址 | 使用集群内 Service DNS | `argocd-server.argocd.svc.cluster.local` |
| PostgreSQL 地址 | 连接同 namespace 的 PostgreSQL | `zcicd-pg-postgresql:5432` |
| Redis 地址 | 连接同 namespace 的 Redis | `zcicd-redis-master:6379` |
| MinIO 地址 | 连接同 namespace 的 MinIO | `zcicd-minio:9000` |
| 数据库迁移 | Helm pre-install hook 自动执行 | 自动 |
| 健康检查 | liveness `/healthz` + readiness `/readyz` | 已配置 |

### Step 6: 验证

```bash
# 检查所有 Pod
kubectl get pods -n zcid
kubectl get pods -n tekton-pipelines
kubectl get pods -n argocd

# 检查 zcid 就绪
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=zcid -n zcid --timeout=120s

# 端口转发访问
kubectl port-forward svc/zcid -n zcid 8080:8080

# 浏览器打开
open http://localhost:8080
# 登录: admin / admin123
```

**验证 CI 功能：**
1. 登录 → 创建项目 → 创建流水线（选择模板）
2. 点击「运行」触发 Pipeline Run
3. 查看运行历史，确认状态从 queued → running → succeeded

**验证 CD 功能：**
1. 创建环境（dev/staging/production）
2. 触发部署，选择环境和镜像
3. 查看部署状态，确认 ArgoCD 同步完成

---

## Helm Values 参考

完整的 `values.yaml` 见 [`helm/zcid/values.yaml`](helm/zcid/values.yaml)。

常用覆盖参数：

```bash
helm install zcid deploy/helm/zcid/ -n zcid \
  # 镜像
  --set image.repository=your-registry/zcid \
  --set image.tag=v1.0.0 \
  # 密钥（必须修改）
  --set secrets.dbPassword=strong-password \
  --set secrets.jwtSecret=random-32-char-string \
  --set config.encryptionKey=0123456789abcdef0123456789abcdef \
  # Ingress
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=zcid.example.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix \
  # 资源
  --set resources.requests.cpu=200m \
  --set resources.requests.memory=256Mi
```

---

## 常见问题

### Q: 不安装 Tekton/ArgoCD 可以运行吗？

**可以。** zcid 自动检测环境：
- 无 K8s → Mock 模式（模拟 Pipeline Run 生命周期）
- 有 K8s 但无 Tekton → Mock 模式
- 有 Tekton → 真实 CI 执行
- 有 ArgoCD → 真实 CD 部署

### Q: Helm 部署后需要手动设置环境变量吗？

**不需要。** Helm chart 通过 `values.yaml` 自动注入所有环境变量到 Pod。K8s 集群内的 Pod 通过 ServiceAccount 自动获取 K8s API 访问权限，无需 KUBECONFIG。

### Q: 如何更新 zcid？

```bash
helm upgrade zcid deploy/helm/zcid/ -n zcid --set image.tag=new-version
# 数据库迁移由 Helm hook 自动执行
```

### Q: 如何查看日志？

```bash
kubectl logs -f deploy/zcid -n zcid
```

### Q: 如何连接外部数据库？

```bash
helm install zcid deploy/helm/zcid/ -n zcid \
  --set config.dbHost=external-db.example.com \
  --set config.dbPort=5432 \
  --set secrets.dbPassword=external-password
```

---

## 资源估算

| 组件 | CPU | Memory | Disk | Namespace |
|------|-----|--------|------|-----------|
| PostgreSQL | 250m | 256Mi | 20Gi | zcid |
| Redis | 100m | 128Mi | 2Gi | zcid |
| MinIO | 250m | 256Mi | 50Gi | zcid |
| Tekton | 500m | 512Mi | — | tekton-pipelines |
| ArgoCD | 550m | 640Mi | — | argocd |
| zcid | 100m | 128Mi | — | zcid |
| **合计** | **1750m** | **1920Mi** | **72Gi** | |

测试环境可以适当降低 persistence.size。

---

## 生产环境建议

1. **密钥管理** — 使用外部 Secret 管理（Vault、AWS Secrets Manager），不要在 values.yaml 中明文存储
2. **PostgreSQL** — 启用主从复制，配置 WAL 备份
3. **Redis** — 启用认证，考虑 Sentinel 高可用
4. **MinIO** — 启用分布式模式（4+ 节点），配置 TLS
5. **Ingress** — 配置 TLS 证书（cert-manager + Let's Encrypt）
6. **资源限制** — 设置合理的 requests/limits，配置 HPA
7. **备份** — 使用 Velero 定期备份 PVC 和数据库
8. **监控** — 接入 Prometheus + Grafana，zcid 暴露 `/healthz` 和 `/readyz`

---

## 子文档

| 文档 | 内容 |
|------|------|
| [`middleware/README.md`](middleware/README.md) | PostgreSQL, Redis, MinIO 部署详情 |
| [`tekton/README.md`](tekton/README.md) | Tekton Pipelines 安装 + RBAC + 验证 |
| [`argocd/README.md`](argocd/README.md) | ArgoCD 安装 + zcid 账号 + Project 配置 |
| [`helm/zcid/values.yaml`](helm/zcid/values.yaml) | Helm chart 完整参数 |
