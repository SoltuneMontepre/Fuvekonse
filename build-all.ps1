# Master build script for all Lambda services

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Building all Lambda services" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$servicesDir = Join-Path $scriptDir "services"

$services = @(
    "general-service",
    "rbac-service",
    "sqs-worker"
)

foreach ($service in $services) {
    Write-Host ""
    Write-Host "Building $service..." -ForegroundColor Yellow
    
    $buildScript = Join-Path $servicesDir "$service\scripts\build.ps1"
    
    if (Test-Path $buildScript) {
        & $buildScript
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to build $service"
            exit $LASTEXITCODE
        }
    } else {
        Write-Warning "Build script not found: $buildScript"
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "All services built successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Deployment packages created:" -ForegroundColor Cyan
foreach ($service in $services) {
    $zipPath = Join-Path $servicesDir "$service\bootstrap.zip"
    if (Test-Path $zipPath) {
        $size = (Get-Item $zipPath).Length / 1MB
        Write-Host "  ✓ $service`: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    }
}
