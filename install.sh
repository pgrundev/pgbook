#!/bin/sh
# pgbook installer — https://pgbook.dev
#
#   curl -fsSL https://pgbook.dev/install.sh | sh
#
# Installs the latest pgbook release for macOS (arm64/amd64) and
# Linux (arm64/amd64) into /usr/local/bin (or $PGBOOK_INSTALL_DIR).
set -eu

REPO="pgrundev/pgbook"
INSTALL_DIR="${PGBOOK_INSTALL_DIR:-/usr/local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$os" in
  darwin|linux) ;;
  *) echo "pgbook: unsupported OS: $os" >&2; exit 1 ;;
esac
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "pgbook: unsupported architecture: $arch" >&2; exit 1 ;;
esac

tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
  grep -m1 '"tag_name"' | cut -d'"' -f4)
[ -n "$tag" ] || { echo "pgbook: cannot determine latest release" >&2; exit 1; }

name="pgbook_${tag#v}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$tag/$name"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading pgbook $tag ($os/$arch)…"
curl -fsSL "$url" -o "$tmp/$name"
curl -fsSL "https://github.com/$REPO/releases/download/$tag/checksums.txt" -o "$tmp/checksums.txt"

(
  cd "$tmp"
  want=$(grep " $name\$" checksums.txt | cut -d' ' -f1)
  [ -n "$want" ] || { echo "pgbook: no published checksum for $name" >&2; exit 1; }
  if command -v sha256sum >/dev/null 2>&1; then
    got=$(sha256sum "$name" | cut -d' ' -f1)
  else
    got=$(shasum -a 256 "$name" | cut -d' ' -f1)
  fi
  [ "$got" = "$want" ] || { echo "pgbook: checksum mismatch for $name" >&2; exit 1; }
  tar -xzf "$name"
)

if [ -w "$INSTALL_DIR" ]; then
  install "$tmp/pgbook" "$INSTALL_DIR/pgbook"
else
  echo "Installing to $INSTALL_DIR (requires sudo)…"
  sudo install "$tmp/pgbook" "$INSTALL_DIR/pgbook"
fi

echo "✓ pgbook $tag installed to $INSTALL_DIR/pgbook"
echo "  Try: pgbook list"
