# Shared helpers for the build scripts. Source this file, do not run it.

SCRIPTS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOT=$(cd "$SCRIPTS_DIR/.." && pwd)
DIST="$ROOT/dist"
# The launcher package embeds the UI with the embedui tag. Set LAUNCHER_PKG to test with another package.
LAUNCHER_PKG="${LAUNCHER_PKG:-./core/cmd/launcher}"
INSTANCE_PKG="./core/cmd/instance"
LAUNCHER_DIR="$ROOT/${LAUNCHER_PKG#./}"
UI_COPY="$LAUNCHER_DIR/ui"

step() {
	printf '==> %s\n' "$1"
}

fail() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

require() {
	command -v "$1" >/dev/null 2>&1 || fail "$1 is not installed"
}

# app_version prints the version from the git tags, without a leading v.
app_version() {
	local version
	version=$(git -C "$ROOT" describe --tags --always)
	echo "${version#v}"
}

# build_ui builds the Svelte UI and copies it into the launcher package. Call clean_ui on exit.
build_ui() {
	step "build UI"
	[[ -d "$LAUNCHER_DIR" ]] || fail "launcher package not found: $LAUNCHER_PKG"
	(cd "$ROOT/src/renderer" && bun run build)
	rm -rf "$UI_COPY"
	cp -r "$ROOT/src/renderer/dist" "$UI_COPY"
}

clean_ui() {
	rm -rf "$UI_COPY"
}

# write_checksums writes dist/SHA256SUMS for every release archive in dist/.
write_checksums() {
	step "write SHA256SUMS"
	(
		cd "$DIST"
		shopt -s nullglob
		local files=(*.tar.gz *.zip *-setup.exe)
		if (( ${#files[@]} > 0 )); then
			sha256sum "${files[@]}" > SHA256SUMS
		fi
	)
}
