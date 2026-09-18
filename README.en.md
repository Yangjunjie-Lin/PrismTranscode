# PrismTranscode · 流光转码

[简体中文](README.md) · **English**

A local-first media workbench: a Windows desktop app and an online edition that converts media inside your browser without uploading it.

[Open online](https://prismtranscode.vercel.app) · [Online guide](docs/USAGE.en.md) · [Windows downloads](https://github.com/Yangjunjie-Lin/PrismTranscode/releases) · [Desktop guide](DESKTOP.en.md) · [Security](SECURITY.en.md)

Current release: **2.1.0-beta.1, community prerelease**. This is not a code-signed, independently audited or fully hardware-certified commercial release.

## English and Chinese interface

The online edition supports English and Simplified Chinese. On a first visit, it matches the browser’s preferred language list against `en` and `zh`, including regional variants such as `en-GB` and `zh-TW`. If neither language matches, it uses English. Use the top-right selector to choose **English / 中文 / Browser default**.

A manual choice is saved in this browser’s `localStorage` and takes precedence on future visits. **Browser default** removes that preference. If browser storage is blocked, switching still works for the current visit. Switching does not reload the page or change your queue, options, running conversion or downloads. All Chinese variants use Simplified Chinese. This language selector applies to the online edition; the desktop app’s interface is still Chinese.

## Choose an edition

| | Online | Windows desktop |
|---|---|---|
| Processing | FFmpeg WebAssembly Worker in your browser | Local Go application + external FFmpeg/FFprobe |
| Output presets | 14: MP3/WAV/FLAC/M4A/OGG/Opus, MP4/WebM, GIF/PNG/JPEG/WebP, SRT/VTT | 38, including HEVC, AV1, ProRes, FFV1, AVIF, JXL and remuxing |
| Files | ≤128 MiB per file; ≤384 MiB / 50 files per queue; sequential | Direct local paths, large files, NCM, 1–4 parallel jobs |
| Results | Individual downloads, JSON report, SHA-256 and structure checks | Local output folder, JSON/CSV, full decoding checks and SHA-256 |
| Privacy | No media uploads; refresh discards files and results | Persistent local queue; backend must remain private |

Presets are not a promise that every input converts to every output. Silent video cannot produce audio, and text subtitle conversion does not include OCR. Online video keeps the first video and audio streams; there is no HDR tone mapping or GPU encoding. Lossless formats cannot restore details already lost.

Online WebM uses **VP8 + Vorbis**. Opus uses a clearly labeled **experimental encoder at 48 kHz stereo**. Use desktop for demanding Opus work, VP9 or larger files.

## What is included in 2.1

- Drag-and-drop imports, deduplication, per-file formats, quality, trimming, size limits, progress, cancellation/retry, downloads and JSON reports.
- English and Chinese navigation, format descriptions, statuses, errors, downloads and report descriptions, with automatic detection and saved manual preference.
- A pinned, same-origin WebAssembly engine with capability detection, output structure verification and SHA-256; no runtime CDN dependency.
- Desktop OS-owned data-directory locks, invalid-upload cleanup, strict JSON parsing and strict handling of explicitly configured engines.
- Windows real-media regression, cross-platform tests, vulnerability scanning, browser E2E, checksums and a GitHub Release workflow.

## Local development

Desktop: Go 1.27.1 is the verified toolchain. Install FFmpeg and FFprobe separately.

```sh
go test ./...
go vet ./...
go run . --no-browser --ffmpeg /path/to/ffmpeg --data-dir ./dev-data
```

Online: Node.js 24 LTS.

```sh
npm ci
npm run build
npm run dev
npm test
npx playwright install chromium
npm run test:e2e
```

Run `build` once before the first development-server start; it copies the pinned engine into same-origin static assets. Production output is `online/dist`. Translation messages live in `online/i18n.js`. Tests cover language selection, regional variants, manual overrides, blocked storage, switching during conversion and responsive layouts.

## Testing and deployment

```powershell
./scripts/build_windows.ps1
./scripts/package_windows.ps1
```

GitHub Actions runs Go tests and vet on Windows/macOS/Linux, Linux race checks, govulncheck, npm audit, real browser conversions and Windows native FFmpeg regression. Version tags trigger the release workflow; checks must pass before producing a Windows ZIP, checksums and build manifest. Hyphenated versions are marked as prereleases.

Vercel hosts only the static online edition using `vercel.json`; pushes to `main` trigger deployment. Never expose the desktop API through a public reverse proxy. The detailed release checklist is in [docs/RELEASE.md](docs/RELEASE.md) (Chinese).

## Privacy and licensing

Media is processed on your device. Requests for the website and engine remain subject to the hosting provider’s normal access logs. The desktop ZIP does not bundle FFmpeg. The online engine is based on FFmpeg 5.1.4, not the latest native FFmpeg, and its compiled C/C++ libraries have not undergone a complete vulnerability audit.

Project-owned code is MIT licensed and preserves the original NCM-MP3 Batch notices. The online FFmpeg core is **GPL-2.0-or-later**; other components retain their own licenses. See [third-party notices](THIRD_PARTY_NOTICES.en.md) for source archives, build recipes and license texts. No independent legal, patent or security certification is provided. Test media is synthetic.

## Bilingual documentation

| Topic | 简体中文 | English |
|---|---|---|
| Project and development | [README](README.md) | [README](README.en.md) |
| Online usage and language | [使用指南](docs/USAGE.zh-CN.md) | [User guide](docs/USAGE.en.md) |
| Desktop usage | [桌面指南](DESKTOP.md) | [Desktop guide](DESKTOP.en.md) |
| Security and privacy | [安全说明](SECURITY.md) | [Security and privacy](SECURITY.en.md) |
| Components and licenses | [许可说明](THIRD_PARTY_NOTICES.md) | [Third-party notices](THIRD_PARTY_NOTICES.en.md) |
