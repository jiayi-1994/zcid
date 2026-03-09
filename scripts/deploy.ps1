# zcid 快速部署脚本 (PowerShell)
# 用法: .\deploy.ps1 [-Branch <branch>]
#
# 会自动: 1) git push 到远端 2) 等待 GitHub Actions 完成 3) 更新 k8s 集群

param(
    [string]$Branch = "main"
)

$ErrorActionPreference = "Stop"

# ========== 配置 ==========
$KubeconfigPath = "config/kubeconfig"
$HelmChartPath = "deploy/helm/zcid"
$HelmNamespace = "zcicd"
$RepoName = "jiayi-1994/zcid"
# ========================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  zcid 快速部署脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "分支: $Branch"
Write-Host ""

# 检查 kubeconfig
if (-not (Test-Path $KubeconfigPath)) {
    Write-Host "错误: 未找到 kubeconfig 文件: $KubeconfigPath" -ForegroundColor Red
    exit 1
}

$env:KUBECONFIG = $KubeconfigPath

# 检查 helm chart
if (-not (Test-Path $HelmChartPath)) {
    Write-Host "错误: 未找到 Helm chart: $HelmChartPath" -ForegroundColor Red
    exit 1
}

# 设置代理 (如果 v2rayN 在运行)
$ProxyPort = 10808
$TestProxy = Test-NetConnection -ComputerName 127.0.0.1 -Port $ProxyPort -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
if ($TestProxy.TcpTestSucceeded) {
    Write-Host "检测到 v2rayN 代理, 设置 HTTP 代理..." -ForegroundColor Yellow
    $env:HTTP_PROXY = "http://127.0.0.1:$ProxyPort"
    $env:HTTPS_PROXY = "http://127.0.0.1:$ProxyPort"
    Write-Host "代理: $env:HTTP_PROXY" -ForegroundColor Gray
}

# ========== Step 1: Git Push ==========
Write-Host ""
Write-Host "[1/4] 推送代码到远端..." -ForegroundColor Green

git push origin $Branch

$CommitHash = git rev-parse --short HEAD
Write-Host "已推送: $CommitHash" -ForegroundColor Gray

# ========== Step 2: 等待 GitHub Actions ==========
Write-Host ""
Write-Host "[2/4] 等待 GitHub Actions 构建完成..." -ForegroundColor Green

# 获取最新的 workflow run
$RunList = gh run list -L 1 --branch $Branch --json databaseId,status,conclusion
$RunData = $RunList | ConvertFrom-Json
$RunId = $RunData[0].databaseId

if (-not $RunId) {
    Write-Host "错误: 无法获取 GitHub Actions run ID" -ForegroundColor Red
    exit 1
}

Write-Host "GitHub Actions Run ID: $RunId" -ForegroundColor Gray

# 轮询等待构建完成
while ($true) {
    $RunView = gh run view $RunId --json status,conclusion | ConvertFrom-Json
    $Status = $RunView.status
    $Conclusion = $RunView.conclusion

    Write-Host "状态: $Status (conclusion: $Conclusion)" -ForegroundColor Gray

    if ($Status -eq "completed") {
        if ($Conclusion -eq "success") {
            Write-Host "构建成功!" -ForegroundColor Green
            break
        } else {
            Write-Host "错误: 构建失败, conclusion: $Conclusion" -ForegroundColor Red
            exit 1
        }
    }

    Write-Host "等待 30 秒后重试..." -ForegroundColor Yellow
    Start-Sleep -Seconds 30
}

# ========== Step 3: 获取构建产物信息 ==========
Write-Host ""
Write-Host "[3/4] 获取构建产物..." -ForegroundColor Green

Write-Host "镜像: ghcr.io/$RepoName:$CommitHash" -ForegroundColor Gray

# ========== Step 4: 更新 K8s 集群 ==========
Write-Host ""
Write-Host "[4/4] 更新 K8s 集群..." -ForegroundColor Green

# 检查当前 helm release 状态
$HelmList = helm list -n $HelmNamespace -o json | ConvertFrom-Json
$CurrentRevision = ($HelmList | Where-Object { $_.name -eq "zcid" }).revision
Write-Host "当前版本: $CurrentRevision" -ForegroundColor Gray

# 执行 helm upgrade
Write-Host "执行 helm upgrade..." -ForegroundColor Yellow
helm upgrade zcid $HelmChartPath -n $HelmNamespace --wait --timeout 5m

# 获取新版本信息
$HelmList = helm list -n $HelmNamespace -o json | ConvertFrom-Json
$NewRevision = ($HelmList | Where-Object { $_.name -eq "zcid" }).revision
Write-Host "新版本: $NewRevision" -ForegroundColor Gray

# 等待 pod 就绪
Write-Host "等待新 Pod 就绪..." -ForegroundColor Yellow
kubectl rollout status deployment/zcid -n $HelmNamespace --timeout=120s

# 获取 pod 信息
$Pod = kubectl get pods -n $HelmNamespace -l app.kubernetes.io=name=zcid -o json | ConvertFrom-Json
$PodName = $Pod.items[0].metadata.name
$PodImage = $Pod.items[0].spec.containers[0].image
$PodStatus = $Pod.items[0].status.phase

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  部署完成!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Pod: $PodName" -ForegroundColor White
Write-Host "镜像: $PodImage" -ForegroundColor White
Write-Host "状态: $PodStatus" -ForegroundColor White
Write-Host ""
