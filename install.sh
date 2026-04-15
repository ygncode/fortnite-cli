#!/bin/sh
# Install the latest fortnite-cli release into $HOME/.local/bin.
# Re-run to upgrade.

set -eu

REPO="ygncode/fortnite-cli"
BIN="fortnite"
INSTALL_DIR="${FORTNITE_INSTALL_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin) os=darwin ;;
  linux)  os=linux ;;
  mingw*|msys*|cygwin*) os=windows ;;
  *) echo "unsupported OS: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  armv7l) arch=armv7 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

ext="tar.gz"
if [ "$os" = "windows" ]; then ext="zip"; fi

# Resolve latest tag.
if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1"; }
else
  fetch() { wget -qO- "$1"; }
fi
tag=$(fetch "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": *"[^"]*"' | head -1 | sed 's/.*"\(.*\)"/\1/')
if [ -z "${tag:-}" ]; then
  echo "failed to resolve latest release tag" >&2
  exit 1
fi

archive="${BIN}_${tag#v}_${os}_${arch}.${ext}"
url="https://github.com/$REPO/releases/download/$tag/$archive"
checksums_url="https://github.com/$REPO/releases/download/$tag/checksums.txt"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $url..."
fetch "$url" > "$tmp/$archive"
fetch "$checksums_url" > "$tmp/checksums.txt"

# Verify checksum.
if command -v sha256sum >/dev/null 2>&1; then
  sha256_cmd="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  sha256_cmd="shasum -a 256"
else
  echo "neither sha256sum nor shasum found; skipping checksum verification" >&2
  sha256_cmd=""
fi
if [ -n "$sha256_cmd" ]; then
  expected=$(grep "  $archive\$" "$tmp/checksums.txt" | awk '{print $1}')
  actual=$(cd "$tmp" && $sha256_cmd "$archive" | awk '{print $1}')
  if [ "$expected" != "$actual" ]; then
    echo "checksum mismatch: expected $expected, got $actual" >&2
    exit 1
  fi
fi

# Extract.
case "$ext" in
  tar.gz) tar -xzf "$tmp/$archive" -C "$tmp" ;;
  zip)    unzip -q "$tmp/$archive" -d "$tmp" ;;
esac

mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/$BIN" "$INSTALL_DIR/$BIN" 2>/dev/null || cp "$tmp/$BIN" "$INSTALL_DIR/$BIN"
chmod 0755 "$INSTALL_DIR/$BIN"

echo "Installed $BIN $tag to $INSTALL_DIR/$BIN"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Note: add $INSTALL_DIR to your PATH to run \`$BIN\`." ;;
esac
