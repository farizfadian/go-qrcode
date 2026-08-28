#!/usr/bin/env bash
# Cross-compile the qrgen CLI for every supported platform into $1.
#
# This lives in one script rather than being written out twice, so CI and the
# release workflow cannot drift apart: if a build breaks, it breaks on the pull
# request rather than at tag time when it is too late to fix quietly.
set -euo pipefail

DIST="${1:-dist}"
mkdir -p "$DIST"

# LDFLAGS strips the symbol table and DWARF data, roughly halving the binaries.
LDFLAGS="-s -w"

platforms=(
  "windows amd64 .exe"
  "windows arm64 .exe"
  "windows 386   .exe"
  "linux   amd64 ''"
  "linux   arm64 ''"
  "linux   386   ''"
  "darwin  amd64 ''"
  "darwin  arm64 ''"
  "freebsd amd64 ''"
  "freebsd arm64 ''"
  "freebsd 386   ''"
)

for entry in "${platforms[@]}"; do
  read -r goos goarch ext <<<"$entry"
  [ "$ext" = "''" ] && ext=""
  out="$DIST/qrgen-${goos}-${goarch}${ext}"
  echo "building $out"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -ldflags="$LDFLAGS" -o "$out" ./cmd/qrgen
done

echo
echo "built $(find "$DIST" -type f | wc -l) binaries:"
ls -la "$DIST"
