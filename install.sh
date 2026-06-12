#!/usr/bin/env sh
set -eu

REPO="${LLMFETCH_REPO:-T-Zevin/llmfetch}"
INSTALL_DIR="${LLMFETCH_INSTALL_DIR:-/usr/local/bin}"
TMP_DIR="${TMPDIR:-/tmp}/llmfetch-install.$$"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "llmfetch installer: missing required command: $1" >&2
    exit 1
  }
}

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

need curl
need tar
need sed
need grep
need uname

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin) os_slug="apple-darwin" ;;
  linux) os_slug="unknown-linux-gnu" ;;
  *)
    echo "llmfetch installer: unsupported OS: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  arm64|aarch64) arch_slug="aarch64" ;;
  x86_64|amd64) arch_slug="x86_64" ;;
  *)
    echo "llmfetch installer: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

latest_url="https://github.com/${REPO}/releases/latest"
version="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$latest_url" | sed 's#.*/tag/##')"

if [ -z "$version" ] || [ "$version" = "$latest_url" ]; then
  echo "llmfetch installer: failed to resolve latest release" >&2
  exit 1
fi

asset="llmfetch-${version#v}-${arch_slug}-${os_slug}.tar.gz"
base_url="https://github.com/${REPO}/releases/download/${version}"

mkdir -p "$TMP_DIR"
cd "$TMP_DIR"

echo "Downloading ${asset}..."
curl -fL -o "$asset" "${base_url}/${asset}"
curl -fL -o checksums.txt "${base_url}/checksums.txt"

expected="$(grep "  ${asset}$" checksums.txt | sed 's/ .*//')"
if [ -z "$expected" ]; then
  echo "llmfetch installer: checksum not found for ${asset}" >&2
  exit 1
fi

if command -v shasum >/dev/null 2>&1; then
  actual="$(shasum -a 256 "$asset" | sed 's/ .*//')"
elif command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$asset" | sed 's/ .*//')"
else
  echo "llmfetch installer: missing shasum or sha256sum" >&2
  exit 1
fi

if [ "$actual" != "$expected" ]; then
  echo "llmfetch installer: checksum mismatch" >&2
  exit 1
fi

tar -xzf "$asset"
binary="$(find . -type f -name llmfetch | head -n 1)"

if [ -z "$binary" ]; then
  echo "llmfetch installer: binary not found in archive" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR" 2>/dev/null || true

if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "$binary" "$INSTALL_DIR/llmfetch"
else
  echo "Installing to ${INSTALL_DIR} requires sudo..."
  sudo install -m 0755 "$binary" "$INSTALL_DIR/llmfetch"
fi

echo "llmfetch installed to ${INSTALL_DIR}/llmfetch"
echo "Run: llmfetch"
