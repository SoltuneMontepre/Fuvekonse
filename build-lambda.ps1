# Docker-based build script for all Lambda services
# This script builds Go Lambda functions using Docker (no local Go installation required)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Building Lambda services with Docker" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

$services = @(
    @{
        Name = "general-service"
        Path = "services/general-service"
        MainFile = "cmd/main.go"
        BuildPath = "./cmd"
    },
    @{
        Name = "rbac-service"
        Path = "services/rbac-service"
        MainFile = "cmd/main.go"
        BuildPath = "./cmd"
    },
    @{
        Name = "sqs-worker"
        Path = "services/sqs-worker"
        MainFile = "main.go"
        BuildPath = "."
    }
)

foreach ($service in $services) {
    Write-Host ""
    Write-Host "Building $($service.Name)..." -ForegroundColor Yellow
    
    $servicePath = Join-Path $scriptDir $service.Path
    $outputZip = Join-Path $servicePath "bootstrap.zip"
    
    # Remove old build artifacts
    if (Test-Path $outputZip) {
        Remove-Item $outputZip -Force
        Write-Host "  Removed old bootstrap.zip" -ForegroundColor Gray
    }
    
    # Build using Docker
    Write-Host "  Compiling Go binary..." -ForegroundColor Cyan
    
    # Determine swag init command based on service structure
    $swagCmd = if ($service.Name -eq "sqs-worker") {
        "swag init -g main.go -o docs --parseDependency --parseInternal || true"
    } else {
        "swag init -g cmd/main.go -o docs --parseDependency --parseInternal"
    }
    
    $dockerCmd = @"
docker run --rm -v "$($servicePath):/build" -w /build golang:1.25-alpine sh -c "
    apk add --no-cache zip && 
    go install github.com/swaggo/swag/cmd/swag@latest && 
    $swagCmd && 
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -o bootstrap $($service.BuildPath) && 
    zip bootstrap.zip bootstrap && 
    rm bootstrap
"
"@
    
    Invoke-Expression $dockerCmd
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build $($service.Name)"
        exit $LASTEXITCODE
    }
    
    if (Test-Path $outputZip) {
        $size = (Get-Item $outputZip).Length / 1MB
        Write-Host "  ✓ Built successfully: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Error "Bootstrap.zip not created for $($service.Name)"
        exit 1
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "All services built successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Deployment packages:" -ForegroundColor Cyan
foreach ($service in $services) {
    $zipPath = Join-Path $scriptDir "$($service.Path)\bootstrap.zip"
    if (Test-Path $zipPath) {
        $size = (Get-Item $zipPath).Length / 1MB
        Write-Host "  $($service.Name): $([math]::Round($size, 2)) MB" -ForegroundColor White
    }
}
Write-Host ""
Write-Host "You can now run: doppler run -- terraform apply --auto-approve -var-file .\envs\prod.tfvars" -ForegroundColor Yellow
