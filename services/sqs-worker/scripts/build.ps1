# Determine script folder and the project root (parent of build/)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$projectRoot = (Resolve-Path (Join-Path $scriptDir '..')).ProviderPath

# Prepare output paths inside the build directory
$linuxOut = Join-Path $scriptDir '..\build\main'
$windowsOut = Join-Path $scriptDir '..\build\main.exe'
$zipPath = Join-Path $scriptDir '..\build\bootstrap.zip'

$Env:GOOS = "linux"
$Env:GOARCH = "amd64"
$Env:CGO_ENABLED = "0"
go build -o "$linuxOut" "$projectRoot"
if ($LASTEXITCODE -ne 0) {
	exit $LASTEXITCODE
}

$Env:GOOS = "windows"
$Env:GOARCH = "amd64"
$Env:CGO_ENABLED = "0"
go build -o "$windowsOut" "$projectRoot"
if ($LASTEXITCODE -ne 0) {
	Write-Error "windows build failed with exit code $LASTEXITCODE"
	exit $LASTEXITCODE
}

# Create a zip containing the linux binary (for Lambda deployment)
if (Get-Command zip -ErrorAction SilentlyContinue) {
	zip -j -r "$zipPath" "$linuxOut"
} else {
	Compress-Archive -Path "$linuxOut" -DestinationPath "$zipPath" -Force
}
