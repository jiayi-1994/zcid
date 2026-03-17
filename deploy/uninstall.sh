#!/bin/bash
set -e
NAMESPACE=zcid

# 是否删除持久化数据（PVC）。默认保留，设置 DELETE_DATA=1 彻底清除
DELETE_DATA=${DELETE_DATA:-}

echo "===== 卸载 zcid 平台 ====="
echo ""

# Step 1: zcid 应用
echo "===== Step 1: 卸载 zcid ====="
helm uninstall zcid -n $NAMESPACE 2>/dev/null && echo "  zcid 已卸载" || echo "  zcid 未安装，跳过"

# Step 2: ArgoCD
echo "===== Step 2: 卸载 ArgoCD ====="
kubectl delete -f deploy/argocd/zcid-argocd-project.yaml 2>/dev/null || true
kubectl delete -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.0/manifests/install.yaml 2>/dev/null \
  && echo "  ArgoCD 已卸载" || echo "  ArgoCD 未安装或已卸载，跳过"

# Step 3: Tekton
echo "===== Step 3: 卸载 Tekton ====="
kubectl delete -f deploy/tekton/zcid-tekton-rbac.yaml 2>/dev/null || true
kubectl delete -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.62.0/release.yaml 2>/dev/null \
  && echo "  Tekton 已卸载" || echo "  Tekton 未安装或已卸载，跳过"

# Step 4: 中间件
echo "===== Step 4: 卸载中间件 ====="
helm uninstall minio -n $NAMESPACE 2>/dev/null && echo "  MinIO 已卸载" || echo "  MinIO 未安装，跳过"
helm uninstall redis -n $NAMESPACE 2>/dev/null && echo "  Redis 已卸载" || echo "  Redis 未安装，跳过"
helm uninstall postgresql -n $NAMESPACE 2>/dev/null && echo "  PostgreSQL 已卸载" || echo "  PostgreSQL 未安装，跳过"

# Step 5: 清理 PVC（可选）
if [ -n "$DELETE_DATA" ]; then
  echo "===== Step 5: 删除持久化数据 (PVC) ====="
  kubectl delete pvc --all -n $NAMESPACE 2>/dev/null && echo "  PVC 已清理" || echo "  无 PVC 需要清理"
else
  echo "===== Step 5: 保留持久化数据 ====="
  echo "  提示: 如需删除数据，重新执行: DELETE_DATA=1 bash deploy/uninstall.sh"
  REMAINING=$(kubectl get pvc -n $NAMESPACE --no-headers 2>/dev/null | wc -l)
  if [ "$REMAINING" -gt 0 ]; then
    echo "  当前保留的 PVC:"
    kubectl get pvc -n $NAMESPACE 2>/dev/null
  fi
fi

# Step 6: 删除 Namespace（可选）
if [ -n "$DELETE_DATA" ]; then
  echo "===== Step 6: 删除 Namespace ====="
  kubectl delete namespace argocd 2>/dev/null && echo "  argocd namespace 已删除" || true
  kubectl delete namespace tekton-pipelines 2>/dev/null && echo "  tekton-pipelines namespace 已删除" || true
  kubectl delete namespace $NAMESPACE 2>/dev/null && echo "  $NAMESPACE namespace 已删除" || true
else
  echo "===== Step 6: 保留 Namespace ====="
  echo "  提示: Namespace 已保留以便重新部署。如需彻底清除，设置 DELETE_DATA=1"
fi

echo ""
echo "===== 卸载完成 ====="
echo ""
if [ -z "$DELETE_DATA" ]; then
  echo "数据已保留。重新部署: bash deploy/install.sh"
  echo "彻底清除:   DELETE_DATA=1 bash deploy/uninstall.sh"
else
  echo "所有资源和数据已彻底清除。"
fi
