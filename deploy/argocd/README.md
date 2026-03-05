# ArgoCD 部署方案

ArgoCD 是 zcid 的 CD（持续部署）引擎。以下提供完整的部署和集成配置方案。

## 前置条件

- Kubernetes 集群 1.25+
- kubectl 已配置
- Helm 3.10+

## 1. 安装 ArgoCD

### 方案 A：使用官方 Manifest（推荐测试环境）

```bash
# 创建 namespace
kubectl create namespace argocd

# 安装 ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml

# 等待就绪
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server \
  -n argocd --timeout=300s
```

### 方案 B：使用 Helm（推荐生产环境）

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

helm install argocd argo/argo-cd \
  --namespace argocd \
  --create-namespace \
  --set server.service.type=ClusterIP \
  --set configs.params."server\.insecure"=true \
  --set server.resources.requests.cpu=100m \
  --set server.resources.requests.memory=128Mi \
  --set controller.resources.requests.cpu=250m \
  --set controller.resources.requests.memory=256Mi \
  --wait --timeout=300s
```

## 2. 访问 ArgoCD

```bash
# 获取初始 admin 密码
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d; echo

# 端口转发
kubectl port-forward svc/argocd-server -n argocd 8443:443

# 浏览器访问 https://localhost:8443
# 用户名: admin
# 密码: 上面获取的密码
```

### 安装 CLI（可选）

```bash
# macOS
brew install argocd

# Linux
curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/download/v2.13.0/argocd-linux-amd64
chmod +x argocd && sudo mv argocd /usr/local/bin/

# 登录
argocd login localhost:8443 --username admin --password <password> --insecure
```

## 3. 为 zcid 配置 ArgoCD

### 3.1 创建 ArgoCD Account

为 zcid 创建专用的 ArgoCD API 账号（而非使用 admin）：

```bash
# 修改 argocd-cm 添加账号
kubectl patch configmap argocd-cm -n argocd --type merge -p '{
  "data": {
    "accounts.zcid": "apiKey, login",
    "accounts.zcid.enabled": "true"
  }
}'

# 设置 RBAC 权限
kubectl patch configmap argocd-rbac-cm -n argocd --type merge -p '{
  "data": {
    "policy.csv": "p, role:zcid-role, applications, *, */*, allow\np, role:zcid-role, projects, get, *, allow\ng, zcid, role:zcid-role"
  }
}'

# 重启 argocd-server 使配置生效
kubectl rollout restart deployment argocd-server -n argocd

# 为 zcid 账号设置密码
argocd account update-password --account zcid \
  --current-password <admin-password> \
  --new-password <zcid-password>

# 生成 API Token
argocd account generate-token --account zcid
# 保存输出的 token，配置到 zcid 的环境变量 ARGOCD_TOKEN
```

### 3.2 创建 ArgoCD Project

```yaml
# deploy/argocd/zcid-argocd-project.yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: zcid-managed
  namespace: argocd
spec:
  description: "Projects managed by zcid platform"
  sourceRepos:
    - '*'
  destinations:
    - namespace: '*'
      server: https://kubernetes.default.svc
  clusterResourceWhitelist:
    - group: ''
      kind: Namespace
  namespaceResourceWhitelist:
    - group: '*'
      kind: '*'
```

```bash
kubectl apply -f deploy/argocd/zcid-argocd-project.yaml
```

### 3.3 Git 仓库凭证（如需私有仓库）

```bash
argocd repo add https://gitlab.example.com/org/deploy-manifests.git \
  --username deploy-token \
  --password <token>
```

## 4. zcid 连接配置

在 zcid 的 Helm values 或环境变量中配置 ArgoCD 连接：

```yaml
# deploy/helm/zcid/values.yaml 中添加：
config:
  argocdUrl: "argocd-server.argocd.svc.cluster.local:443"
  argocdInsecure: "true"

secrets:
  argocdToken: "<上面生成的 API Token>"
```

或直接设置环境变量：

```bash
export ARGOCD_SERVER=argocd-server.argocd.svc.cluster.local:443
export ARGOCD_TOKEN=<your-token>
export ARGOCD_INSECURE=true
```

## 5. 应用部署流程

zcid 通过 ArgoCD API 管理部署：

```
用户触发部署
  → zcid 调用 ArgoCD gRPC API
    → CreateOrUpdateApp (创建/更新 Application)
    → SyncApp (触发同步)
    → GetAppStatus (监听状态)
  → zcid 展示部署状态
```

### Application 示例

zcid 会自动生成类似以下的 ArgoCD Application：

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: zcid-mall-user-service-dev
  namespace: argocd
  labels:
    zcid.io/managed-by: zcid
    zcid.io/project: mall-platform
    zcid.io/environment: dev
spec:
  project: zcid-managed
  source:
    repoURL: https://gitlab.example.com/org/deploy-manifests.git
    targetRevision: main
    path: envs/dev/user-service
  destination:
    server: https://kubernetes.default.svc
    namespace: mall-dev
  syncPolicy:
    automated:
      prune: false
      selfHeal: false
```

## 6. 完整安装脚本

```bash
#!/bin/bash
set -e

echo "=== Installing ArgoCD ==="
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server \
  -n argocd --timeout=300s

echo "=== Configuring zcid account ==="
kubectl patch configmap argocd-cm -n argocd --type merge -p '{
  "data": {
    "accounts.zcid": "apiKey, login",
    "accounts.zcid.enabled": "true"
  }
}'

kubectl patch configmap argocd-rbac-cm -n argocd --type merge -p '{
  "data": {
    "policy.csv": "p, role:zcid-role, applications, *, */*, allow\np, role:zcid-role, projects, get, *, allow\ng, zcid, role:zcid-role"
  }
}'

echo "=== Creating ArgoCD Project ==="
kubectl apply -f deploy/argocd/zcid-argocd-project.yaml

echo "=== Restarting ArgoCD server ==="
kubectl rollout restart deployment argocd-server -n argocd
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server \
  -n argocd --timeout=120s

echo ""
echo "=== ArgoCD deployed ==="
echo "Get admin password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
echo "Port forward: kubectl port-forward svc/argocd-server -n argocd 8443:443"
echo ""
echo "Next steps:"
echo "  1. Login to ArgoCD and set zcid account password"
echo "  2. Generate API token: argocd account generate-token --account zcid"
echo "  3. Configure token in zcid Helm values or environment"
```

## 7. 版本兼容性

| ArgoCD 版本 | gRPC API | zcid 兼容 |
|-------------|---------|----------|
| v2.8 - v2.10 | v1 | 基础支持 |
| v2.11 - v2.13 | v1 | 推荐 |
| v2.14+ | v1 | 需要验证 |

## 8. 资源估算

| 组件 | CPU Request | Memory Request |
|------|------------|----------------|
| argocd-server | 100m | 128Mi |
| argocd-controller | 250m | 256Mi |
| argocd-repo-server | 100m | 128Mi |
| argocd-redis | 50m | 64Mi |
| argocd-dex-server | 50m | 64Mi |
| **合计** | **550m** | **640Mi** |
