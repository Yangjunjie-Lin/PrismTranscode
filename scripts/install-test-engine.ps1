param([string]$Destination = ".cache/ffmpeg-test")
$ErrorActionPreference = "Stop"
$Destination = [IO.Path]::GetFullPath($Destination)
New-Item -ItemType Directory -Force -Path $Destination | Out-Null
$metadata = Invoke-RestMethod 'https://api.github.com/repos/BtbN/FFmpeg-Builds/releases/latest'
$asset = $metadata.assets | Where-Object name -eq 'ffmpeg-master-latest-win64-gpl.zip'
$checks = $metadata.assets | Where-Object name -eq 'checksums.sha256'
$checksText = (Invoke-WebRequest $checks.browser_download_url).Content
if ($checksText -is [byte[]]) { $checksText = [Text.Encoding]::UTF8.GetString($checksText) }
$line = ($checksText -split "`n" | Where-Object { $_ -match ' ffmpeg-master-latest-win64-gpl.zip$' })
if (-not $line) { throw 'Missing publisher checksum' }
$expected = ($line.Trim() -split '\s+')[0]
$zip = Join-Path $Destination 'engine.zip'
Invoke-WebRequest $asset.browser_download_url -OutFile $zip
$actual = (Get-FileHash -LiteralPath $zip -Algorithm SHA256).Hash.ToLower()
if ($actual -ne $expected) { throw 'Engine checksum mismatch; archive not executed' }
Expand-Archive -LiteralPath $zip -DestinationPath $Destination
$engine = Get-ChildItem -LiteralPath $Destination -Recurse -Filter ffmpeg.exe | Select-Object -First 1
Write-Output "Validated engine: $($engine.FullName)"
Write-Output "SHA256: $actual"
& $engine.FullName -version | Select-Object -First 3
