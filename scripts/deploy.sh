#!/bin/bash
# zcid 快速部署脚本
# 用法: ./deploy.sh [git-branch]
#
# 会自动: 1) git push 到远端 2) 等待 GitHub Actions 完成 3) 更新 k8s 集群

set -e

# ========== 配置 ==========
KUBECONFIG_PATH="config/kubeconfig"
HELM_CHART_PATH="deploy/helm/zcid"
HELM_NAMESPACE="zcicd"
REPO_NAME="jiayi-1994/zcid"
# ========================

# 解析参数
BRANCH="${1:-main}"

echo "========================================"
echo "  zcid 快速部署脚本"
echo "========================================"
echo "分支: $BRANCH"
echo ""

# 检查 kubeconfig
if [ ! -f "$KUBECONFIG_PATH" ]; then
    echo "错误: 未找到 kubeconfig 文件: $KUBECONFIG_PATH"
    exit 1
fi

export KUBECONFIG="$KUBECONFIG_PATH"

# 检查 helm chart
if [ ! -d "$HELM_CHART_PATH" ]; then
    echo "错误: 未找到 Helm chart: $HELM_CHART_PATH"
    exit 1
fi

# 检查代理设置 (如果有 v2rayN)
if [ -n "$HTTP_PROXY" ]; then
    echo "使用 HTTP 代理: $HTTP_PROXY"
fi

# ========== Step 1: Git Push ==========
echo ""
echo "[1/4] 推送代码到远端..."
git push origin "$BRANCH"

# 获取当前 commit hash
COMMIT_HASH=$(git rev-parse --short HEAD)
echo "已推送: $COMMIT_HASH"

# ========== Step 2: 等待 GitHub Actions ==========
echo ""
echo "[2/4] 等待 GitHub Actions 构建完成..."

# 获取最新的 workflow run
RUN_ID=$(gh run list -L 1 --branch "$BRANCH" --json databaseId --jq '.[0].databaseId')
if [ -z "$RUN_ID" ]; then
    echo "错误: 无法获取 GitHub Actions run ID"
    exit 1
fi

echo "GitHub Actions Run ID: $RUN_ID"

# 轮询等待构建完成
while true; do
    STATUS=$(gh run view "$RUN_ID" --json status --jq '.status')
    CONCLUSION=$(gh run view "$RUN_ID" --json conclusion --jq '.conclusion')

    echo "状态: $STATUS (conclusion: $CONCLUSION)"

    if [ "$STATUS" = "completed" ]; then
        if [ "$CONCLUSION" = "success" ]; then
            echo "构建成功!"
            break
        else
            echo "错误: 构建失败, conclusion: $CONCLUSION"
            exit 1
        fi
    fi

    echo "等待 30 秒后重试..."
    sleep 30
done

# ========== Step 3: 获取构建产物信息 ==========
echo ""
echo "[3/4] 获取构建产物..."

# 从日志中提取镜像信息
LOG_FILE=$(gh run view "$RUN_ID" --log 2>/dev/null | head -1 || true)
echo "镜像: ghcr.io/$REPO_NAME:$COMMIT_HASH"

# ========== Step 4: 更新 K8s 集群 ==========
echo ""
echo "[4/4] 更新 K8s 集群..."

# 检查当前 helm release 状态
CURRENT_REVISION=$(helm list -n "$HELM_NAMESPACE" -o json | jq -r '.[] | select(.name=="zcid") | .revision')
echo "当前版本: $CURRENT_REVISION"

# 执行 helm upgrade
helm upgrade zcid "$HELM_CHART_PATH" \
    -n "$HELM_NAMESPACE" \
    --wait \
    --timeout 5m

# 获取新版本信息
NEW_REVISION=$(helm list -n "$HELM_NAMESPACE" -o json | jq -r '.[] | select(.name=="zcid") | .revision')
echo "新版本: $NEW_REVISION"

# 等待 pod 就绪
echo "等待新 Pod 就绪..."
kubectl rollout status deployment/zcid -n "$HELM_NAMESPACE" --timeout=120s

# 获取 pod 信息
POD_NAME=$(kubectl get pods -n "$HELM_NAMESPACE" -l app.kubernetes.io/name=zcid -o jsonpath='{.items[0].metadata.name}')
POD_IMAGE=$(kubectl get pod "$POD_NAME" -n "$HELM_NAMESPACE" -o jsonpath='{.spec.containers[0].image}')
POD_STATUS=$(kubectl get pod "$POD_NAME" -n "$HELM_NAMESPACE" -o jsonpath='{.status.phase}')

echo ""
echo "========================================"
echo "  部署完成!"
echo "========================================"
echo "Pod: $POD_NAME"
echo "镜像: $POD_IMAGE"
echo "状态: $POD_STATUS"
echo ""
