#!/usr/bin/env bash
# Package the cross-compiled binaries into per-platform archives.
# Requires build/ to be produced first (make build).
#
# Usage: scripts/package.sh [build_dir] [dist_dir] [target]
set -euo pipefail

BUILD_DIR="${1:-build}"
DIST_DIR="${2:-dist}"
TARGET="${3:-modstart-cli}"

# Keep in sync with PLATFORMS in the Makefile.
PLATFORMS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64"

# Use absolute paths: several steps below change directory.
DIST_ABS="$(cd "$(dirname "$DIST_DIR")" 2>/dev/null && pwd)/$(basename "$DIST_DIR")"
BUILD_ABS="$(cd "$BUILD_DIR" && pwd)"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

if ! command -v zip >/dev/null 2>&1; then
	echo "zip not found: required for the Windows archives" >&2
	exit 1
fi

rm -rf "$DIST_ABS"
mkdir -p "$DIST_ABS"

echo "Packaging $TARGET from $BUILD_ABS -> $DIST_ABS"

for p in $PLATFORMS; do
	os="${p%/*}"
	arch="${p#*/}"
	ext=""
	[ "$os" = "windows" ] && ext=".exe"

	src="$BUILD_ABS/$TARGET-$os-$arch$ext"
	if [ ! -f "$src" ]; then
		echo "missing binary: $src" >&2
		exit 1
	fi

	name="$TARGET-$os-$arch"
	stage="$tmp/$name"
	mkdir -p "$stage"
	cp "$src" "$stage/$TARGET$ext"
	chmod +x "$stage/$TARGET$ext"

	# Ship the license (and README) with every binary, as Apache-2.0 requires.
	contents=("$TARGET$ext")
	for doc in LICENSE README.md; do
		if [ -f "$doc" ]; then
			cp "$doc" "$stage/$doc"
			contents+=("$doc")
		fi
	done

	if [ "$os" = "windows" ]; then
		(cd "$stage" && zip -q "$DIST_ABS/$name.zip" "${contents[@]}")
	else
		tar -C "$stage" -czf "$DIST_ABS/$name.tar.gz" "${contents[@]}"
	fi

	# Also keep the raw executable: some scripts prefer it over extracting an archive.
	cp "$src" "$DIST_ABS/$name$ext"

	echo "  packaged $name"
done

# sha256 of every artifact, for verifying downloads.
(
	cd "$DIST_ABS"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum ./* >checksums.txt
	else
		shasum -a 256 ./* >checksums.txt
	fi
	sed -i.bak 's@\./@@g' checksums.txt 2>/dev/null || true
	rm -f checksums.txt.bak
)

echo "Done:"
ls -1 "$DIST_ABS"
