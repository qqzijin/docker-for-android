#!/usr/bin/env bash
# Update docker/arm64_bin binaries from CDN based on version.txt
# - Fetches version.txt from CDN (not cached)
# - Parses version and arm64 package info
# - Downloads tarball from CDN (fallback to origin on failure)
# - Verifies sha256
# - Extracts to docker/ (overwriting docker/arm64_bin)

set -euo pipefail

# Configurable endpoints (can be overridden via env)
CDN_URL=${CDN_URL:-"https://fw.kspeeder.com/binary/docker-for-android"}
ORIGIN_SERVER_URL=${ORIGIN_SERVER_URL:-"https://fw.koolcenter.com/binary/docker-for-android"}

# Paths (relative to repo root)
REPO_ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
DOCKER_DIR="$REPO_ROOT_DIR/docker"
ARM64_BIN_DIR="$DOCKER_DIR/arm64_bin"
TMP_DIR="$(mktemp -d -t dfa-arm64-XXXX)"
CLEANUP() { rm -rf "$TMP_DIR" 2>/dev/null || true; }
trap CLEANUP EXIT

info()  { printf "[info] %s\n" "$*"; }
warn()  { printf "[warn] %s\n" "$*" >&2; }
error() { printf "[error] %s\n" "$*" >&2; exit 1; }

need_cmd() { command -v "$1" >/dev/null 2>&1 || error "required command not found: $1"; }

need_cmd curl
need_cmd tar
# Prefer shasum on macOS, fallback to sha256sum if available
if command -v shasum >/dev/null 2>&1; then
  SHASUM_CMD=(shasum -a 256)
elif command -v sha256sum >/dev/null 2>&1; then
  SHASUM_CMD=(sha256sum)
else
  error "neither shasum nor sha256sum found"
fi

mkdir -p "$ARM64_BIN_DIR"

# Step 1: fetch version.txt (from CDN since .txt is not cached there)
info "Fetching version.txt from CDN: $CDN_URL/version.txt"
if ! curl -fsSL "$CDN_URL/version.txt" -o "$TMP_DIR/version.txt"; then
  error "failed to download version.txt from CDN"
fi

# Step 2: parse version and package info
VERSION=$(grep -E '^VERSION=' "$TMP_DIR/version.txt" | head -n1 | cut -d'=' -f2- || true)
ARM64_PACKAGE=$(grep -E '^ARM64_PACKAGE=' "$TMP_DIR/version.txt" | head -n1 | cut -d'=' -f2- || true)
ARM64_SHA256=$(grep -E '^ARM64_SHA256=' "$TMP_DIR/version.txt" | head -n1 | cut -d'=' -f2- || true)

[ -n "$ARM64_PACKAGE" ] || error "ARM64_PACKAGE not found in version.txt"
[ -n "$ARM64_SHA256" ]  || error "ARM64_SHA256 not found in version.txt"

info "Latest version: ${VERSION:-unknown}"
info "ARM64 package: $ARM64_PACKAGE"

# Step 3: download tarball (CDN first, fallback to origin)
ARM64_TGZ="$TMP_DIR/$ARM64_PACKAGE"
PKG_URL_CDN="$CDN_URL/$ARM64_PACKAGE"
PKG_URL_ORIGIN="$ORIGIN_SERVER_URL/$ARM64_PACKAGE"

info "Downloading arm64 package from CDN..."
if ! curl -fSL "$PKG_URL_CDN" -o "$ARM64_TGZ"; then
  warn "CDN download failed, trying origin server..."
  curl -fSL "$PKG_URL_ORIGIN" -o "$ARM64_TGZ" || error "failed to download package from both CDN and origin"
fi

# Step 4: verify checksum
info "Verifying sha256 checksum..."
# The checksum file may contain only the hash; we construct a line for verification
CHECKSUM_LINE="$ARM64_SHA256  $ARM64_TGZ"
if ! printf "%s\n" "$CHECKSUM_LINE" | "${SHASUM_CMD[@]}" -c - >/dev/null 2>&1; then
  # Some shasum implementations don't support -c; do manual compare
  CALC=$("${SHASUM_CMD[@]}" "$ARM64_TGZ" | awk '{print $1}')
  [ "$CALC" = "$ARM64_SHA256" ] || error "checksum mismatch: expected $ARM64_SHA256 got $CALC"
fi
info "Checksum OK"

# Step 5: extract to docker/ (archive root is arm64_bin)
info "Extracting into $DOCKER_DIR ..."
mkdir -p "$DOCKER_DIR"
tar -xzf "$ARM64_TGZ" -C "$DOCKER_DIR"

# Step 6: report
info "Updated binaries in: $ARM64_BIN_DIR"
ls -l "$ARM64_BIN_DIR" || true

info "Done."
