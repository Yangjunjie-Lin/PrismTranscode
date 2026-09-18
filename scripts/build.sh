#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/.."
go test ./...
go vet ./...
mkdir -p dist
CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags '-s -w' -o dist/prismtranscode .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags '-s -w -H=windowsgui' -o dist/PrismTranscode.exe .
printf '%s\n' 'Built current-platform executable and Windows x64 executable. No FFmpeg engine is bundled.'
