# Build script for general-service Lambda deployment

# Determine script folder and the project root
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$projectRoot = (Resolve-Path (Join-Path $scriptDir '..')).ProviderPath

# Prepare output paths
$buildDir = Join-Path $projectRoot 'build'
$linuxOut = Join-Path $buildDir 'bootstrap'
$zipPath = Join-Path $projectRoot 'bootstrap.zip'

# Create build directory if it doesn't exist
if (-Not (Test-Path $buildDir)) {
    New-Item -ItemType Directory -Path $buildDir | Out-Null
}

Write-Host "Building general-service for AWS Lambda..." -ForegroundColor Cyan

# Build for Linux (Lambda runtime)
$Env:GOOS = "linux"
$Env:GOARCH = "amd64"
$Env:CGO_ENABLED = "0"

go build -ldflags="-s -w" -o "$linuxOut" "$projectRoot\cmd\main.go"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Linux build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}

Write-Host "Creating deployment package..." -ForegroundColor Cyan

# Create a zip containing the linux binary (for Lambda deployment)
if (Test-Path $zipPath) {
    Remove-Item $zipPath -Force
}

if (Get-Command zip -ErrorAction SilentlyContinue) {
    Set-Location $buildDir
    zip -j "$zipPath" "bootstrap"
    Set-Location $projectRoot
} else {
    Compress-Archive -Path "$linuxOut" -DestinationPath "$zipPath" -Force
}

Write-Host "Build complete: $zipPath" -ForegroundColor Green
