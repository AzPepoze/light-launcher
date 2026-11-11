#!/usr/bin/env bash
# Builds the launcher with the UI copied into the package and embedded.
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

OUTPUT="${1:-$ROOT/bin/light-launcher}"
[[ "$OUTPUT" = /* ]] || OUTPUT="$PWD/$OUTPUT"

require go
require bun

VERSION=$(app_version)

trap clean_ui EXIT
build_ui

step "build launcher ($VERSION)"
mkdir -p "$(dirname "$OUTPUT")"
(cd "$ROOT" && go build -tags embedui -ldflags="-s -w -X main.version=$VERSION" -o "$OUTPUT" "$LAUNCHER_PKG")
