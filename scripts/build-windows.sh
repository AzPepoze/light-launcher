#!/usr/bin/env bash
# Cross-builds the Windows release from Linux: launcher, instance, WebView2 shell,
# dist/light-launcher-<version>-windows-x64.zip, and the NSIS installer when makensis is installed.
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

require go
require bun
require zip

VERSION=$(app_version)
STAGE="$DIST/windows"
PORTABLE="$DIST/light-launcher-$VERSION-windows-x64.zip"
INSTALLER="$DIST/light-launcher-$VERSION-windows-x64-setup.exe"

export GOOS=windows GOARCH=amd64 CGO_ENABLED=0

trap clean_ui EXIT
build_ui

step "build launcher ($VERSION)"
rm -rf "$STAGE"
mkdir -p "$STAGE"
(cd "$ROOT" && go build -tags embedui -ldflags="-s -w -X main.version=$VERSION" -o "$STAGE/light-launcher.exe" "$LAUNCHER_PKG")

step "build instance"
(cd "$ROOT" && go build -ldflags="-s -w" -o "$STAGE/light-launcher-instance.exe" "$INSTANCE_PKG")

step "build WebView2 shell"
(cd "$ROOT/shell/webview2" && go build -ldflags="-s -w -H=windowsgui" -o "$STAGE/light-launcher-shell.exe" .)

cp "$ROOT/LICENSE" "$ROOT/README.md" "$STAGE/"

step "pack $(basename "$PORTABLE")"
rm -f "$PORTABLE"
(cd "$STAGE" && zip -qr "$PORTABLE" .)

if command -v makensis >/dev/null 2>&1; then
	step "build $(basename "$INSTALLER")"
	rm -f "$INSTALLER"
	makensis -V2 \
		-DVERSION="$VERSION" \
		-DSTAGE_DIR="$STAGE" \
		-DOUTFILE="$INSTALLER" \
		"$ROOT/install/windows/installer.nsi"
else
	printf 'warning: makensis not found, skipping the installer\n' >&2
fi

write_checksums
ls -l "$PORTABLE"
