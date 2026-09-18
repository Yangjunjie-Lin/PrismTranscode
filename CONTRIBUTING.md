# Contributing

Run `go test ./...`, `go vet ./...`, `npm ci`, `npm run check` and `npm run test:e2e` before proposing changes. Install Chromium with `npx playwright install chromium` first. A complete native FFmpeg installation is required for `scripts/smoke_test.py`.

Never commit personal media, session URLs, local paths from runtime logs, credentials, node_modules, downloaded engines or build caches. Fixtures must be synthetic or licensed for redistribution. Preserve the MIT notices and third-party licenses. Codec support, losslessness and platform compatibility claims need actual test evidence.

Report reproducible issues with OS/browser versions, selected output settings and a small synthetic example where possible. Do not attach sensitive media.
