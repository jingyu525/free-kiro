#!/usr/bin/env bash
# install.sh — one-line installer for free-kiro.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/install.sh | bash
#
# Environment variables:
#   FREE_KIRO_VERSION   Specific version to install (default: latest)
#   FREE_KIRO_DIR       Install location (default: ~/.local/bin)
#   FREE_KIRO_REPO      GitHub repo (default: jingyu525/free-kiro)
#
# This script detects the platform, downloads the matching binary
# from GitHub Releases, verifies the SHA256, and drops it in
# ~/.local/bin (or FREE_KIRO_DIR). Idempotent — re-running
# overwrites the previous binary.

set -euo pipefail

REPO="${FREE_KIRO_REPO:-jingyu525/free-kiro}"
INSTALL_DIR="${FREE_KIRO_DIR:-$HOME/.local/bin}"
VERSION="${FREE_KIRO_VERSION:-latest}"
BINARY="free-kiro"

# --- platform detection ------------------------------------------------------
detect_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"
  case "$os" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *)      echo "error: unsupported OS '$os' (only linux + darwin supported)" >&2; exit 1 ;;
  esac
  case "$arch" in
    x86_64|amd64)  arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)             echo "error: unsupported architecture '$arch'" >&2; exit 1 ;;
  esac
  echo "${os}_${arch}"
}

# --- version selection -------------------------------------------------------
if [ "$VERSION" = "latest" ]; then
  echo "→ resolving latest release from $REPO …"
  VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep -oE '"tag_name":\s*"v[^"]+"' | head -1 | sed -E 's/.*"v([^"]+)".*/\1/')"
  if [ -z "$VERSION" ]; then
    echo "error: could not determine latest version" >&2; exit 1
  fi
fi

PLATFORM="$(detect_platform)"
TARBALL="${BINARY}_${VERSION}_${PLATFORM}.tar.gz"
URL_BASE="https://github.com/$REPO/releases/download/v${VERSION}"
URL="${URL_BASE}/${TARBALL}"
SHA_URL="${URL_BASE}/${BINARY}_${VERSION}_SHA256SUMS"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "→ downloading $TARBALL …"
curl -fsSL -o "$TMPDIR/$TARBALL" "$URL"
curl -fsSL -o "$TMPDIR/SHA256SUMS" "$SHA_URL"

echo "→ verifying checksum …"
( cd "$TMPDIR" && grep "  $TARBALL" SHA256SUMS | sha256sum -c - )

echo "→ extracting …"
tar -xzf "$TMPDIR/$TARBALL" -C "$TMPDIR"

echo "→ installing to $INSTALL_DIR …"
mkdir -p "$INSTALL_DIR"
install -m 0755 "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"

echo ""
echo "✓ free-kiro v$VERSION installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Next steps:"
echo "  1. Make sure $INSTALL_DIR is on your PATH:"
echo "       export PATH=\"\$HOME/.local/bin:\$PATH\"  # add to ~/.zshrc or ~/.bashrc"
echo ""
echo "  2. Verify:"
echo "       $BINARY --version"
echo ""
echo "  3. Initialize a project:"
echo "       cd your-project && $BINARY init"
echo ""