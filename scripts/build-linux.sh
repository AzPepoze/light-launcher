#!/usr/bin/env bash
# Builds the Linux release: launcher, instance, GTK shell, and dist/light-launcher-<version>-linux-x64.tar.gz.
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

require go
require bun
require tar

VERSION=$(app_version)
STAGE="$DIST/linux"
ARCHIVE="$DIST/light-launcher-$VERSION-linux-x64.tar.gz"

rm -rf "$STAGE"
mkdir -p "$STAGE"
bash "$SCRIPTS_DIR/build-launcher.sh" "$STAGE/light-launcher"

step "build instance"
(cd "$ROOT" && go build -ldflags="-s -w" -o "$STAGE/light-launcher-instance" "$INSTANCE_PKG")

step "build GTK shell"
(cd "$ROOT/shell/gtk" && go build -ldflags="-s -w" -o "$STAGE/light-launcher-shell" .)

cp "$ROOT/LICENSE" "$ROOT/README.md" "$STAGE/"

step "pack $(basename "$ARCHIVE")"
rm -f "$ARCHIVE"
tar -C "$STAGE" -czf "$ARCHIVE" .

write_checksums
ls -l "$ARCHIVE"
