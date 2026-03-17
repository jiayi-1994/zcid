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

## 中国网络环境配置

国内服务器拉取 `gcr.io`、`ghcr.io`、`registry.k8s.io`、`quay.io` 等海外镜像仓库时经常超时。推荐使用 GitHub 镜像代理加速。

### 镜像代理对照表

| 原始仓库 | 代理地址 | 使用场景 |
|---------|---------|---------|
| `ghcr.io` | `docker.gh-proxy.com/ghcr.io` | zcid 平台镜像、Tekton (ghcr 版本) |
| `gcr.io` | `docker.gh-proxy.com/gcr.io` | Tekton (官方版本) |
| `registry.k8s.io` | `docker.gh-proxy.com/registry.k8s.io` | K8s 系统组件 |
| `quay.io` | `docker.gh-proxy.com/quay.io` | ArgoCD |
| `docker.io` | `docker.gh-proxy.com/docker.io` | 基础镜像 (alpine, ubuntu 等) |

### 方案 1：配置 containerd 全局镜像代理（推荐）

在所有 K8s 节点上配置 containerd 镜像代理，对所有部署操作全局生效：

```bash
# 在每个 K8s 节点执行
sudo mkdir -p /etc/containerd/certs.d/ghcr.io
sudo tee /etc/containerd/certs.d/ghcr.io/hosts.toml <<EOF
[host."https://docker.gh-proxy.com/ghcr.io"]
  capabilities = ["pull", "resolve"]
EOF

sudo mkdir -p /etc/containerd/certs.d/gcr.io
sudo tee /etc/containerd/certs.d/gcr.io/hosts.toml <<EOF
[host."https://docker.gh-proxy.com/gcr.io"]
  capabilities = ["pull", "resolve"]
EOF

sudo mkdir -p /etc/containerd/certs.d/quay.io
sudo tee /etc/containerd/certs.d/quay.io/hosts.toml <<EOF
[host."https://docker.gh-proxy.com/quay.io"]
  capabilities = ["pull", "resolve"]
EOF

sudo mkdir -p /etc/containerd/certs.d/registry.k8s.io
sudo tee /etc/containerd/certs.d/registry.k8s.io/hosts.toml <<EOF
[host."https://docker.gh-proxy.com/registry.k8s.io"]
  capabilities = ["pull", "resolve"]
EOF

# 重启 containerd
sudo systemctl restart containerd
```

> 配置后无需修改任何 YAML 文件中的镜像地址，containerd 会自动通过代理拉取。

### 方案 2：使用已替换镜像地址的 Tekton 部署文件

项目中已提供将 `gcr.io` 替换为 `ghcr.io` 的 Tekton 部署文件，在中国网络环境下可直接使用：

```bash
# 使用 ghcr.io 版本（替代 gcr.io，国内更快）
kubectl apply -f deploy/tekton/tekton-v0.62.0-ghcr.yaml

# 如果 ghcr.io 也较慢，可以手动替换为代理地址：
sed 's|ghcr.io|docker.gh-proxy.com/ghcr.io|g' deploy/tekton/tekton-v0.62.0-ghcr.yaml | kubectl apply -f -
```

### 方案 3：替换 zcid 和 ArgoCD 镜像地址

```bash
# zcid Helm chart - 使用代理拉取平台镜像
helm install zcid deploy/helm/zcid/ -n zcid \
  --set image.repository=docker.gh-proxy.com/ghcr.io/jiayi-1994/zcid \
  ...

# ArgoCD - 使用代理拉取 manifest
curl -sSL https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml \
  | sed 's|quay.io|docker.gh-proxy.com/quay.io|g' \
  | sed 's|ghcr.io|docker.gh-proxy.com/ghcr.io|g' \
  | kubectl apply -n argocd -f -
```

### 方案 4：Helm 仓库代理

如果 `charts.bitnami.com` 也无法访问：

```bash
# 使用国内 Helm 仓库镜像
helm repo add bitnami https://charts.bitnami.com/bitnami

# 或使用阿里云 Helm 镜像
helm repo add bitnami-ali https://kubernetes.oss-cn-hangzhou.aliyuncs.com/charts
```

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

适用于测试环境的一键部署脚本。

> **Bitnami 镜像迁移注意**：2025 年 8 月起，Bitnami 已将免费 Docker 镜像从 `docker.io/bitnami/` 迁移到 `docker.io/bitnamilegacy/`，且不再更新。脚本默认使用 `bitnamilegacy`。如已购买 Bitnami Secure Images 订阅，可设置 `BITNAMI_REPO=bitnami` 使用付费仓库。

```bash
# 最简部署（使用集群默认 StorageClass）
bash deploy/install.sh

# 指定 StorageClass
STORAGE_CLASS=local-path bash deploy/install.sh

# 无持久化存储（测试环境，无需 StorageClass/PV）
PERSISTENCE=false bash deploy/install.sh

# 中国环境（自动使用 docker.gh-proxy.com 镜像代理）
USE_PROXY=1 bash deploy/install.sh

# 使用付费 Bitnami 仓库
BITNAMI_REPO=bitnami bash deploy/install.sh

# 完整自定义示例
USE_PROXY=1 STORAGE_CLASS=ceph-rbd BITNAMI_REPO=bitnami bash deploy/install.sh
```

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `STORAGE_CLASS` | 空（集群默认） | 中间件 PVC 使用的 StorageClass，留空则使用集群默认 |
| `PERSISTENCE` | `true` | 设为 `false` 禁用 PVC 持久化（使用 emptyDir，适合测试环境） |
| `USE_PROXY` | 空 | 设为 `1` 启用中国镜像代理 |
| `BITNAMI_REPO` | `bitnamilegacy` | Bitnami Docker 镜像仓库前缀 |

---

## 一键卸载

```bash
# 卸载所有组件，保留数据（PVC）以便重新部署
bash deploy/uninstall.sh

# 彻底清除，包括持久化数据和 Namespace
DELETE_DATA=1 bash deploy/uninstall.sh
```

卸载顺序与安装相反：zcid → ArgoCD → Tekton → MinIO → Redis → PostgreSQL。默认保留 PVC 数据，方便重装后恢复。

---

## 分步部署

### Step 1: 创建 Namespace

```bash
kubectl create namespace zcid
kubectl create namespace argocd
```

### Step 2: 部署中间件

详细文档：[`middleware/README.md`](middleware/README.md)

> **Bitnami 镜像迁移**：2025 年 8 月起，`docker.io/bitnami/*` 已迁移至 `docker.io/bitnamilegacy/*` 且停止更新。部署时需通过 `--set image.repository=bitnamilegacy/<chart>` 覆盖镜像地址，否则会拉取失败。如已购买 Bitnami Secure Images 订阅，可继续使用 `bitnami`。

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update

# PostgreSQL（注意 image.repository 使用 bitnamilegacy）
# storageClass 留空使用集群默认，或指定为集群中已有的 StorageClass
helm install postgresql bitnami/postgresql -n zcid \
  --set image.registry=docker.io --set image.repository=bitnamilegacy/postgresql \
  --set auth.database=zcicd \
  --set auth.username=postgres \
  --set auth.postgresPassword=<your-password> \
  --set primary.persistence.size=20Gi

# Redis
helm install redis bitnami/redis -n zcid \
  --set image.registry=docker.io --set image.repository=bitnamilegacy/redis \
  --set architecture=standalone --set auth.enabled=false

# MinIO
helm install minio bitnami/minio -n zcid \
  --set image.registry=docker.io --set image.repository=bitnamilegacy/minio \
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
# 海外环境：
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.62.0/release.yaml

# 中国环境（使用项目内 ghcr 版本 + 代理）：
sed 's|ghcr.io|docker.gh-proxy.com/ghcr.io|g' deploy/tekton/tekton-v0.62.0-ghcr.yaml | kubectl apply -f -

# 等待就绪
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
# 海外环境：
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml

# 中国环境（使用镜像代理）：
curl -sSL https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml \
  | sed 's|quay.io|docker.gh-proxy.com/quay.io|g' \
  | sed 's|ghcr.io|docker.gh-proxy.com/ghcr.io|g' \
  | kubectl apply -n argocd -f -

# 等待就绪
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

### Q: 中国环境部署时镜像拉取超时怎么办？

三种方案（任选其一）：

1. **containerd 全局代理**（推荐）— 配置一次，所有拉取自动走代理，见上方「中国网络环境配置」
2. **一键脚本加代理**：`USE_PROXY=1 bash deploy/install.sh`
3. **手动 sed 替换**：在 `kubectl apply` 前用 `sed` 将镜像地址替换为 `docker.gh-proxy.com/` 前缀

### Q: Bitnami 中间件镜像拉取失败 (ImagePullBackOff)？

2025 年 8 月起，`docker.io/bitnami/*` 已停止更新并迁移至 `docker.io/bitnamilegacy/*`。

**方案 1**（推荐）：使用一键部署脚本，已自动处理：
```bash
bash deploy/install.sh
```

**方案 2**：手动指定 bitnamilegacy 仓库：
```bash
helm install postgresql bitnami/postgresql -n zcid \
  --set image.registry=docker.io --set image.repository=bitnamilegacy/postgresql ...
```

**方案 3**：如已购买 Bitnami Secure Images 订阅，可继续使用 `bitnami`：
```bash
BITNAMI_REPO=bitnami bash deploy/install.sh
```

> ⚠️ `bitnamilegacy` 镜像不再接收安全更新。生产环境建议购买 Bitnami 订阅或迁移到其他维护中的 Helm chart（如 CloudNativePG、Dragonfly 等）。

### Q: 中间件 Pod 一直 Pending，提示 "unbound immediate PersistentVolumeClaims"？

集群中没有可用的 StorageClass 动态供给器（如 `local-path-provisioner`、`nfs-provisioner`、`ceph-csi` 等），PVC 无法自动绑定 PV。

**方案 1**（推荐，测试环境）：禁用持久化存储，使用 emptyDir：
```bash
PERSISTENCE=false bash deploy/install.sh
```
> ⚠️ Pod 重启后数据会丢失，仅适合测试/演示。

**方案 2**：安装存储供给器（如 Rancher Local Path Provisioner）：
```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.30/deploy/local-path-storage.yaml
# 设置为默认 StorageClass（如果还没有默认的）
kubectl patch storageclass local-path -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
# 然后重新部署
STORAGE_CLASS=local-path bash deploy/install.sh
```

**方案 3**：指定集群中已有的 StorageClass：
```bash
# 查看可用的 StorageClass
kubectl get storageclass
# 指定其中一个
STORAGE_CLASS=<your-storageclass> bash deploy/install.sh
```

**方案 4**：手动创建 hostPath PV（适合单节点测试集群）：
```bash
for name in data-postgresql-0 data-redis-master-0 data-minio-0; do
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ${name}
spec:
  capacity:
    storage: 10Gi
  accessModes: [ReadWriteOnce]
  hostPath:
    path: /data/k8s-pv/${name}
  claimRef:
    namespace: zcid
    name: ${name}
EOF
sudo mkdir -p /data/k8s-pv/${name}
done
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
