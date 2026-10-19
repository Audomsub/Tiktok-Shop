# Remote Deploy Script for TikTok Affiliate Analytics (Windows Host to Ubuntu VM)
[CmdletBinding()]
param(
    [string]$HostName = "127.0.0.1",
    [int]$Port = 2222,
    [string]$User = "audomsub",
    [string]$KeyPath = "$HOME\.ssh\id_ed25519"
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " Triggering Remote Deployment to $User@$HostName`:$Port" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$RemoteCmd = "if [ ! -d /home/$User/tiktok-shop ]; then git clone https://github.com/Audomsub/Tiktok-Shop.git /home/$User/tiktok-shop; fi && chmod +x /home/$User/tiktok-shop/scripts/deploy.sh && /home/$User/tiktok-shop/scripts/deploy.sh"

$sshArgs = @(
    "-i", $KeyPath,
    "-p", $Port.ToString(),
    "-o", "BatchMode=yes",
    "-o", "StrictHostKeyChecking=no",
    "$User@$HostName",
    $RemoteCmd
)

& ssh @sshArgs
if ($LASTEXITCODE -ne 0) {
    Write-Host "Deployment FAILED with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "Deployment SUCCESSFUL!" -ForegroundColor Green
Write-Host "Access Platform at: http://localhost:8080" -ForegroundColor Yellow
