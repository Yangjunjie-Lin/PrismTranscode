param([string]$Version = "2.1.0-beta.1")
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
$binary = Join-Path $Root 'dist/PrismTranscode.exe'
if (-not (Test-Path -LiteralPath $binary)) { throw 'Run build_windows.ps1 first' }
$stage = Join-Path $Root "artifacts/PrismTranscode-Windows-x64-$Version"
if (Test-Path -LiteralPath $stage) { throw "Packaging directory already exists: $stage. Choose a new version or inspect it manually." }
New-Item -ItemType Directory -Path $stage | Out-Null
Copy-Item -LiteralPath $binary -Destination $stage
foreach ($name in @('LICENSE', 'README.md', 'DESKTOP.md', 'SECURITY.md', 'THIRD_PARTY_NOTICES.md', 'CHANGELOG.md', 'licenses', 'presets')) { Copy-Item -LiteralPath (Join-Path $Root $name) -Destination $stage -Recurse }
New-Item -ItemType Directory -Path (Join-Path $stage 'docs') | Out-Null
foreach ($name in @('RELEASE.md', 'FORMATS_QUALITY.md', 'API.md', 'ARCHITECTURE.md')) { Copy-Item -LiteralPath (Join-Path $Root "docs/$name") -Destination (Join-Path $stage 'docs') }
$archive = Join-Path $Root "artifacts/PrismTranscode-Windows-x64-$Version.zip"
Compress-Archive -LiteralPath $stage -DestinationPath $archive
$sha = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLower()
$exeSha = (Get-FileHash -LiteralPath $binary -Algorithm SHA256).Hash.ToLower()
"$sha  $([IO.Path]::GetFileName($archive))" | Set-Content -LiteralPath (Join-Path $Root 'artifacts/SHA256SUMS.txt') -Encoding utf8
@{ version=$Version; os='windows'; arch='amd64'; go=(& go version); commit=(& git -C $Root rev-parse HEAD); binary_sha256=$exeSha; archive_sha256=$sha; signed=$false; ffmpeg_bundled=$false } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $Root 'artifacts/build-manifest.json') -Encoding utf8
Write-Output "Release package: $archive"
