# Start APIHub backend with China-friendly Go proxy (if default proxy fails)
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOSUMDB = "off"

Write-Host "Running go mod tidy..."
go mod tidy

Write-Host "Starting APIHub on http://localhost:8080 ..."
go run ./cmd/apihub
