param([string]$Output = "")
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
if (-not $Output) { $Output = Join-Path $Root "dist\PrismTranscode.exe" }
$Output = [System.IO.Path]::GetFullPath($Output)
New-Item -ItemType Directory -Force -Path (Split-Path $Output) | Out-Null
$oldOS=$env:GOOS; $oldArch=$env:GOARCH; $oldCGO=$env:CGO_ENABLED
Push-Location $Root
try {
    & go test ./...
    if ($LASTEXITCODE -ne 0) { throw "Unit tests failed" }
    $env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
    & go build -trimpath -buildvcs=false -ldflags "-s -w -H=windowsgui" -o $Output .
    if ($LASTEXITCODE -ne 0) { throw "Windows build failed" }
    Get-FileHash -Algorithm SHA256 $Output | Format-List
    Write-Host "Built: $Output (FFmpeg/FFprobe are not bundled)"
} finally {
    $env:GOOS=$oldOS; $env:GOARCH=$oldArch; $env:CGO_ENABLED=$oldCGO
    Pop-Location
}
