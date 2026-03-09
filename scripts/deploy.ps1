# zcid quick deploy script (PowerShell)
# Usage: .\deploy.ps1 [-Branch <branch>]
#
# Auto: 1) git push 2) wait GitHub Actions 3) update k8s cluster

param(
    [string]$Branch = "main"
)

$ErrorActionPreference = "Stop"

# ========== Config ==========
$KubeconfigPath = "config/kubeconfig"
$HelmChartPath = "deploy/helm/zcid"
$HelmNamespace = "zcicd"
$RepoName = "jiayi-1994/zcid"
# ========================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  zcid Quick Deploy Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Branch: $Branch"
Write-Host ""

# Check kubeconfig
if (-not (Test-Path $KubeconfigPath)) {
    Write-Host "Error: kubeconfig not found: $KubeconfigPath" -ForegroundColor Red
    exit 1
}

$env:KUBECONFIG = $KubeconfigPath

# Check helm chart
if (-not (Test-Path $HelmChartPath)) {
    Write-Host "Error: Helm chart not found: $HelmChartPath" -ForegroundColor Red
    exit 1
}

# Set proxy (if v2rayN is running)
$ProxyPort = 10808
$TestProxy = Test-NetConnection -ComputerName 127.0.0.1 -Port $ProxyPort -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
if ($TestProxy.TcpTestSucceeded) {
    Write-Host "Detected v2rayN proxy, setting HTTP proxy..." -ForegroundColor Yellow
    $env:HTTP_PROXY = "http://127.0.0.1:$ProxyPort"
    $env:HTTPS_PROXY = "http://127.0.0.1:$ProxyPort"
    Write-Host "Proxy: $env:HTTP_PROXY" -ForegroundColor Gray
}

# ========== Step 1: Git Push ==========
Write-Host ""
Write-Host "[1/4] Push to remote..." -ForegroundColor Green

git push origin $Branch

$CommitHash = git rev-parse --short HEAD
Write-Host "Pushed: $CommitHash" -ForegroundColor Gray

# ========== Step 2: Wait GitHub Actions ==========
Write-Host ""
Write-Host "[2/4] Waiting for GitHub Actions..." -ForegroundColor Green

# Get latest workflow run
$RunList = gh run list -L 1 --branch $Branch --json databaseId,status,conclusion
$RunData = $RunList | ConvertFrom-Json
$RunId = $RunData[0].databaseId

if (-not $RunId) {
    Write-Host "Error: Cannot get GitHub Actions run ID" -ForegroundColor Red
    exit 1
}

Write-Host "GitHub Actions Run ID: $RunId" -ForegroundColor Gray

# Poll until build completes
while ($true) {
    $RunView = gh run view $RunId --json status,conclusion | ConvertFrom-Json
    $Status = $RunView.status
    $Conclusion = $RunView.conclusion

    Write-Host "Status: $Status (conclusion: $Conclusion)" -ForegroundColor Gray

    if ($Status -eq "completed") {
        if ($Conclusion -eq "success") {
            Write-Host "Build success!" -ForegroundColor Green
            break
        } else {
            Write-Host "Error: Build failed, conclusion: $Conclusion" -ForegroundColor Red
            exit 1
        }
    }

    Write-Host "Waiting 30 seconds..." -ForegroundColor Yellow
    Start-Sleep -Seconds 30
}

# ========== Step 3: Get build artifacts ==========
Write-Host ""
Write-Host "[3/4] Getting build artifacts..." -ForegroundColor Green

Write-Host "Image: ghcr.io/${RepoName}:${CommitHash}" -ForegroundColor Gray

# ========== Step 4: Update K8s cluster ==========
Write-Host ""
Write-Host "[4/4] Updating K8s cluster..." -ForegroundColor Green

# Check current helm release
$HelmList = helm list -n $HelmNamespace -o json | ConvertFrom-Json
$CurrentRevision = ($HelmList | Where-Object { $_.name -eq "zcid" }).revision
Write-Host "Current revision: $CurrentRevision" -ForegroundColor Gray

# Execute helm upgrade
Write-Host "Running helm upgrade..." -ForegroundColor Yellow
helm upgrade zcid $HelmChartPath -n $HelmNamespace --wait --timeout 5m

# Get new revision
$HelmList = helm list -n $HelmNamespace -o json | ConvertFrom-Json
$NewRevision = ($HelmList | Where-Object { $_.name -eq "zcid" }).revision
Write-Host "New revision: $NewRevision" -ForegroundColor Gray

# Wait for pod ready
Write-Host "Waiting for Pod ready..." -ForegroundColor Yellow
kubectl rollout status deployment/zcid -n $HelmNamespace --timeout=120s

# Get pod info
$Pod = kubectl get pods -n $HelmNamespace -l app.kubernetes.io/instance=zcid -o json 2>$null | ConvertFrom-Json
if ($Pod.items -and $Pod.items.Count -gt 0) {
    $PodName = $Pod.items[0].metadata.name
    $PodImage = $Pod.items[0].spec.containers[0].image
    $PodStatus = $Pod.items[0].status.phase
} else {
    $PodName = "N/A"
    $PodImage = "N/A"
    $PodStatus = "N/A"
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Deploy Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Pod: $PodName" -ForegroundColor White
Write-Host "Image: $PodImage" -ForegroundColor White
Write-Host "Status: $PodStatus" -ForegroundColor White
Write-Host ""
