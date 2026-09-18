# Security and privacy

[简体中文](SECURITY.md) · **English**

Process only trusted media you have the right to use. The desktop HTTP backend is a single-user localhost service, not an internet service or a security sandbox.

## Online edition

Vercel serves a static site. File API data enters a WebAssembly Worker in browser memory; there is no media upload endpoint, analytics code or third-party runtime script. The host can still log requests for the site and engine, including IP, time and User-Agent. Download results before closing or refreshing the tab.

Manual language selection stores only `zh` or `en` under `prismtranscode.language` in localStorage. Browser-default mode removes that setting. No media, queue or conversion result is persisted there. If localStorage is blocked, the interface remains usable with a session-only language choice.

The same-origin Content Security Policy restricts scripts, connections and Workers. Camera, microphone, location access and embedding are disabled by policy. File names and translations are rendered as text, never evaluated as HTML. Downloads use sanitized names. Limits are 128 MiB per input, 384 MiB / 50 inputs per queue; decoding and results can consume more memory. Cancellation terminates the Worker; each probing, encoding and verification phase uses an isolated engine heap. A main-thread watchdog terminates jobs exceeding five minutes.

The pinned `@ffmpeg/core 0.12.10` is based on FFmpeg 5.1.4 and includes older C/C++ libraries. `npm audit` does not scan those compiled libraries. Browser isolation reduces exposure of local files, but it does not guarantee freedom from vulnerabilities. Online libopus/VP9 are not used after observed WebAssembly failures; the UI discloses experimental native Opus and VP8/Vorbis WebM instead. Independent review and a maintained codec-upgrade strategy are required before treating this as a certified commercial service.

## Desktop protections

- Listens only on `127.0.0.1`; `--listen` cannot expose it on `0.0.0.0`.
- Each launch creates a random session token. API checks enforce Bearer authentication, Host and Origin. Do not share `last-session-url.txt` or token-bearing URLs.
- FFmpeg is invoked with argument arrays, not shell commands assembled from filenames. Inputs are local regular files, and allowed protocols are `file,pipe`.
- Originals are read-only. Output commits do not replace existing files. Input parameters, JSON, concurrency and queue sizes are limited; CSV output escapes common formula prefixes.
- Engine downloads require explicit action, HTTPS host/redirect allowlists and a distributor SHA-256 match. Archive and extracted sizes are limited; only intended executables/notices are extracted.
- An OS-owned lock prevents two processes from writing the same data directory. The OS releases it on crash; do not delete the lock file manually.

## Remaining risks

Native FFmpeg is not isolated with AppContainer, seccomp, a container or a dedicated account. Do not run as administrator to process unknown files. Independent penetration testing, a full security audit and code signing have not been completed. GPU/driver matrices, native-dialog acceptance and large-file/disk-failure testing remain incomplete.

Checksums detect mismatched or damaged downloads, not a compromise of both the distributor’s archive and checksum publication. The installer does not verify FFmpeg source PGP signatures or imply that third-party Windows binaries are officially signed by FFmpeg.

Imports, temporary decoded media, queue paths and metadata can contain private information on disk. Normal shutdown removes its session directory but retains imports/history; crashes can leave temporary data. Removing a queue item does not remove original media. Clean caches only after stopping all instances, and never commit runtime data, media or tokens to GitHub.

Release assets include `build-manifest.json` and `SHA256SUMS.txt`. The 2.1 beta was rebuilt with Go 1.27.1 and passed its GitHub Actions release checks; these checks are not an independent security or legal certification. Codec licenses, patents and media rights require review for the actual distribution/use case.

To report vulnerabilities, use GitHub private vulnerability reporting if enabled. Otherwise open an issue without sensitive media, tokens, local paths or publicly exploitable details.
