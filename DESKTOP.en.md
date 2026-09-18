# PrismTranscode desktop guide

[简体中文](DESKTOP.md) · **English** · [Online guide](docs/USAGE.en.md)

The desktop edition is a local, offline-first batch media workbench compiled with Go. It opens a localhost page in your system browser; it is not a hosted converter or a native WebView application. The online edition has English/Chinese switching; the desktop interface is still Chinese. Chinese button labels below help you find the controls.

**Release: 2.1.0-beta.1, unsigned community prerelease.** Windows x64 testing covers 59 native-media checks, 12 browser GUI checks, packaged-program startup, exclusive data-directory locking and lock recovery after a crash. Physical GPU matrices, manual native-dialog acceptance, independent security auditing and commercial legal review remain incomplete. See [the current release checklist](docs/RELEASE.md); `docs/ACCEPTANCE.md` records the older 2.0 tests.

## Getting started on Windows

1. Download the Windows ZIP from [Releases](https://github.com/Yangjunjie-Lin/PrismTranscode/releases), check its SHA-256 and extract it fully. Run `PrismTranscode.exe`. End users do not need Go, Python or .NET.
2. Open **转换引擎** (Conversion engine). Choose a folder containing both `ffmpeg.exe` and `ffprobe.exe`, or explicitly choose an engine download. The desktop ZIP does **not** include FFmpeg. Downloads are checked against the distributor’s SHA-256; a failed download is not reported as installed.
3. Use **添加文件** (Add files), **添加文件夹** (Add folder) or **路径导入** (Import paths). Native path input avoids making an extra media copy. Drag-and-drop/browser imports copy files to a local cache, never to the internet.
4. Set defaults before import, or change individual rows. The right-hand settings update existing jobs only when you click **将以上参数应用到选中文件** (Apply to selected files). Choose an output directory, select jobs and click **开始转换** (Start).
5. Open the output folder when done. **详情** (Details) shows detected streams, execution arguments, warnings, output SHA-256 and engine logs. Failed jobs do not stop other jobs. Use **选择失败 / 停止项** (Select failed/stopped jobs) to retry; reapply settings before reconverting a completed file.

BtbN provides a full GPL development build; Gyan provides release essentials. They are not equivalent: essentials may omit SoXR or SVT-AV1, disabling dependent presets. For reproducible work, select a fixed, validated engine and retain its hash/configuration. Windows 10/11 x64 is recommended; the chosen engine determines additional requirements. Engines may also be copied, with their notices, into a `tools/` folder next to the executable.

## Capabilities

- Actual container/stream detection, codec/duration/dimensions/sample-rate/channel/bit-depth/HDR information, NCM decoding before probing.
- Mixed-format imports, deduplication, recursive folders, per-file formats, shared settings for selected jobs, 1–4 workers, progress, cancellation, failure isolation and queue recovery.
- Audio: MP3, AAC/M4A/ADTS, FLAC, PCM WAV/AIFF, ALAC, Vorbis, Opus, WMA, AC-3, WavPack and original-codec audio extraction.
- Video: H.264/H.265/AV1/VP9, ProRes 422 HQ and FFV1, using appropriate MP4/MKV/WebM/MOV/AVI/MPG/TS presets; MP4/MKV remuxing.
- Images/subtitles: PNG, JPEG, WebP, lossless WebP, AVIF, lossless JPEG XL, TIFF, BMP, GIF, SRT, WebVTT and ASS. Still-image presets output a single frame.
- Sample-rate/bit-depth/channel/bitrate settings, optional SoXR, video quality/size/fps/trim controls, hardware encoding and explicitly recorded CPU fallback.
- Read-only originals, no overwriting existing outputs, staged output commits, structure verification, default full decoding checks, optional hashes and JSON/CSV reports.

The 38 presets do not guarantee every input/codec combination. See [the format and quality matrix](docs/FORMATS_QUALITY.md) (Chinese).

## Common tasks and fidelity limits

**Video to MP3:** select an input containing audio and choose MP3. High-quality transcoding uses 320 kbps by default. If the source already contains compatible MP3 audio and no transformation is needed, the app can copy it. Silent video fails rather than producing fake audio.

**Extract without re-encoding:** choose **原码提取 · 不重编码**. AAC normally produces M4A, while MP3 produces MP3. Converting AAC to MP3 necessarily re-encodes it.

**NCM batches:** choose MP3, FLAC or another supported audio target. MP3-to-FLAC cannot restore lost detail. NCM cover art is not migrated in this version; title, artist and album text can be retained.

Ordinary video transcoding does not promise every subtitle, attachment, cover or private metadata field survives. Prefer compatible remuxing for track preservation. HDR is blocked for ordinary lossy workflows rather than silently producing incorrect colors. No AI restoration, interpolation or upscaling is provided.

## Data, shutdown and safety

Closing the browser does not stop the program; use **退出** (Quit). Settings and queue data normally live in `%APPDATA%\PrismTranscode`. If the browser does not open, `last-session-url.txt` contains the current localhost URL. Do not share its session token.

Only one process may use a data directory. OS locks are released on crash; do not manually delete `instance.lock`. Imports and queue history persist locally; deleting a queue entry does not delete the original media. Stop the app before manually cleaning caches. Never expose or reverse-proxy this localhost backend publicly.

No document/archive conversion, RAW development, image-subtitle OCR, cloud video downloads or Widevine/FairPlay/QMC/KGM unlocking is included. HEIC and unusual inputs depend on the engine and are not comprehensively validated. Process trusted files as a standard user: native FFmpeg is not sandboxed. The executable is not code-signed.

## Building and tests

The module baseline is Go 1.23; the verified release uses Go 1.27.1. Desktop has no third-party Go modules and needs no npm build.

```sh
go test ./...
go vet ./...
go run . --no-browser --ffmpeg /path/to/ffmpeg --data-dir ./dev-data
```

Windows builds use `scripts/build_windows.ps1`. Linux/macOS can be built from source, with manually provided FFmpeg/FFprobe and path/browser imports; automatic engine installation is Windows x64 only. Cross-compilation is not platform runtime certification.

`scripts/smoke_test.py` generates synthetic media and exercises the real API/engine. `scripts/ui_test.py` tests real browser interactions; native OS dialogs remain outside that automated test. `scripts/runtime_windows_test.py` checks packaged-program startup and lock/crash behavior. Reproduction details are in [docs/RELEASE.md](docs/RELEASE.md).

Project-owned code is MIT licensed with the original NCM-MP3 Batch notices retained. Go, FFmpeg and codecs retain their own licenses. No LosslessCut or Shutter Encoder GUI code is included. Redistribution and commercial use require independent license, patent and media-rights review. See [third-party notices](THIRD_PARTY_NOTICES.en.md) and [security](SECURITY.en.md).
