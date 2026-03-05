# Tekton 部署方案

Tekton 是 zcid 的 CI 执行引擎。以下提供完整的 Tekton 部署和配置方案。

## 前置条件

- Kubernetes 集群 1.25+
- kubectl 已配置
- 集群至少 2 CPU / 4Gi 内存可用

## 1. 安装 Tekton Pipelines

```bash
# 安装 Tekton Pipelines v0.62 (支持 v1 API)
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.62.0/release.yaml

# 等待就绪
kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=tekton-pipelines \
  -n tekton-pipelines --timeout=300s
```

### 验证安装

```bash
kubectl get pods -n tekton-pipelines
# 应看到：
# tekton-pipelines-controller-xxx   Running
# tekton-pipelines-webhook-xxx      Running

# 检查 CRD
kubectl get crd | grep tekton
# 应包含：
# pipelines.tekton.dev
# pipelineruns.tekton.dev
# tasks.tekton.dev
# taskruns.tekton.dev
```

## 2. 安装 Tekton Dashboard（可选）

```bash
kubectl apply --filename https://storage.googleapis.com/tekton-releases/dashboard/previous/v0.46.0/release-full.yaml

# 端口转发
kubectl port-forward svc/tekton-dashboard -n tekton-pipelines 9097:9097
# 浏览器访问 http://localhost:9097
```

## 3. 配置 Tekton

### 3.1 Feature Flags

```bash
kubectl patch configmap feature-flags -n tekton-pipelines \
  --type merge -p '{"data":{
    "disable-affinity-assistant": "true",
    "running-in-environment-with-injected-sidecars": "true",
    "enable-api-fields": "beta"
  }}'
```

### 3.2 默认超时

```bash
kubectl patch configmap config-defaults -n tekton-pipelines \
  --type merge -p '{"data":{
    "default-timeout-minutes": "60",
    "default-managed-by-label-value": "zcid"
  }}'
```

### 3.3 清理策略（PipelineRun TTL）

```bash
kubectl patch configmap feature-flags -n tekton-pipelines \
  --type merge -p '{"data":{
    "keep-pod-on-cancel": "false"
  }}'
```

> zcid 内置 CRDCleaner 定时清理过期 PipelineRun，默认保留 7 天。

## 4. RBAC 配置

为 zcid 创建最小权限 ServiceAccount：

```yaml
# deploy/tekton/zcid-tekton-rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: zcid-tekton-sa
  namespace: zcid
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: zcid-tekton-role
rules:
  # Tekton PipelineRun 管理
  - apiGroups: ["tekton.dev"]
    resources: ["pipelineruns", "taskruns"]
    verbs: ["create", "get", "list", "watch", "delete", "update", "patch"]
  - apiGroups: ["tekton.dev"]
    resources: ["pipelines", "tasks"]
    verbs: ["get", "list"]
  # Pod 日志读取
  - apiGroups: [""]
    resources: ["pods", "pods/log"]
    verbs: ["get", "list", "watch"]
  # Secret 临时注入
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["create", "delete", "get"]
  # Namespace 查询
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: zcid-tekton-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: zcid-tekton-role
subjects:
  - kind: ServiceAccount
    name: zcid-tekton-sa
    namespace: zcid
```

```bash
kubectl apply -f deploy/tekton/zcid-tekton-rbac.yaml
```

## 5. Harbor 镜像仓库凭证

为 Tekton 任务配置 Harbor 推送凭证：

```bash
kubectl create secret docker-registry harbor-creds \
  --namespace zcid \
  --docker-server=harbor.example.com \
  --docker-username=admin \
  --docker-password=Harbor12345

# 关联到 zcid-tekton-sa
kubectl patch serviceaccount zcid-tekton-sa -n zcid \
  -p '{"secrets": [{"name": "harbor-creds"}]}'
```

## 6. 完整安装脚本

```bash
#!/bin/bash
set -e

echo "=== Installing Tekton Pipelines ==="
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.62.0/release.yaml
kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=tekton-pipelines \
  -n tekton-pipelines --timeout=300s

echo "=== Installing Tekton Dashboard ==="
kubectl apply --filename https://storage.googleapis.com/tekton-releases/dashboard/previous/v0.46.0/release-full.yaml

echo "=== Configuring Tekton ==="
kubectl patch configmap feature-flags -n tekton-pipelines \
  --type merge -p '{"data":{"disable-affinity-assistant":"true","running-in-environment-with-injected-sidecars":"true","enable-api-fields":"beta"}}'

kubectl patch configmap config-defaults -n tekton-pipelines \
  --type merge -p '{"data":{"default-timeout-minutes":"60","default-managed-by-label-value":"zcid"}}'

echo "=== Creating RBAC ==="
kubectl apply -f deploy/tekton/zcid-tekton-rbac.yaml

echo "=== Tekton deployment complete ==="
kubectl get pods -n tekton-pipelines
```

## 7. 验证 Tekton + zcid 集成

```bash
# 创建一个测试 PipelineRun
cat <<EOF | kubectl apply -f -
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  name: zcid-test-run
  namespace: zcid
  labels:
    zcid.io/managed-by: zcid
spec:
  pipelineSpec:
    tasks:
      - name: hello
        taskSpec:
          steps:
            - name: echo
              image: alpine:3.21
              command: ["echo", "zcid + Tekton integration works!"]
EOF

# 等待完成
kubectl wait --for=condition=succeeded pipelinerun/zcid-test-run -n zcid --timeout=120s

# 查看日志
kubectl logs -n zcid -l tekton.dev/pipelineRun=zcid-test-run --all-containers

# 清理
kubectl delete pipelinerun zcid-test-run -n zcid
```

## 8. 版本兼容性

| Tekton 版本 | API 版本 | zcid 兼容 |
|-------------|---------|----------|
| v0.44 - v0.55 | v1 | 基础支持 |
| v0.56 - v0.62 | v1 + beta | 推荐 |
| v0.63+ | v1 | 需要验证 |

zcid 在启动时会检测 `tekton.dev/v1` CRD 是否注册，如未安装 Tekton 则以 mock 模式运行。
