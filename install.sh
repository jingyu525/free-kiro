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
#   FREE_KIRO_SKIP_PATH Set to 1 to skip PATH configuration
#
# This script detects the platform, downloads the matching binary
# from GitHub Releases, verifies the SHA256, and drops it in
# ~/.local/bin (or FREE_KIRO_DIR). Idempotent — re-running
# overwrites the previous binary. Also auto-appends ~/.local/bin to
# the user's shell rc (zshrc / bashrc / fish config) when missing.

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

# --- PATH auto-configuration -------------------------------------------------
# Detect shell rc files and append a single export line if ~/.local/bin
# isn't already on PATH. Skip when FREE_KIRO_SKIP_PATH=1 or when running
# in a non-interactive context (e.g. CI / curl|bash with --norc).

configure_path() {
  if [ "${FREE_KIRO_SKIP_PATH:-0}" = "1" ]; then
    return
  fi
  # Already on PATH?
  case ":$PATH:" in
    *":$INSTALL_DIR:"*) return ;;
  esac

  # Pick the right rc for this shell. Don't try to be clever — just
  # update the ones that exist, preferring the user's $SHELL.
  local rc_files=()
  case "${SHELL:-}" in
    */zsh)  rc_files=("$HOME/.zshrc") ;;
    */bash) rc_files=("$HOME/.bashrc") ;;
    */fish) rc_files=("$HOME/.config/fish/config.fish") ;;
  esac
  # Also try the others as fallback — many macOS users have bash as $SHELL
  # but a .zshrc lying around from defaults.
  [ -f "$HOME/.zshrc" ] && rc_files+=("$HOME/.zshrc")
  [ -f "$HOME/.bashrc" ] && rc_files+=("$HOME/.bashrc")
  [ -f "$HOME/.bash_profile" ] && rc_files+=("$HOME/.bash_profile")
  [ -f "$HOME/.config/fish/config.fish" ] && rc_files+=("$HOME/.config/fish/config.fish")

  # Filter to ones that actually exist.
  local existing=()
  for rc in "${rc_files[@]}"; do
    [ -f "$rc" ] && existing+=("$rc")
  done
  if [ ${#existing[@]} -eq 0 ]; then
    echo ""
    echo "⚠️  $INSTALL_DIR is not on PATH and no shell rc file was found."
    echo "   Add it manually:  export PATH=\"\$HOME/.local/bin:\$PATH\""
    return
  fi

  local updated=()
  for rc in "${existing[@]}"; do
    # Skip if this rc already has our export (idempotent across rcs).
    if grep -qF "$INSTALL_DIR" "$rc" 2>/dev/null; then
      continue
    fi
    local marker="# free-kiro"
    if grep -qF "$marker" "$rc" 2>/dev/null; then
      # We've appended before — just append the export (idempotent).
      :
    else
      printf '\n%s\n' "$marker" >> "$rc"
    fi
    if [[ "$rc" == *"fish"* ]]; then
      printf 'set -gx PATH "%s" $PATH\n' "$INSTALL_DIR" >> "$rc"
    else
      printf 'export PATH="%s:$PATH"\n' "$INSTALL_DIR" >> "$rc"
    fi
    updated+=("$rc")
  done

  if [ ${#updated[@]} -gt 0 ]; then
    echo ""
    echo "✓ Added $INSTALL_DIR to PATH in: ${updated[*]}"
    echo "  →  reload your shell (e.g. \`source ~/.zshrc\`) or open a new terminal"
  else
    echo ""
    echo "✓ $INSTALL_DIR is already configured in your shell rc"
  fi
}

configure_path

echo ""
echo "Next steps:"
echo "  1. Verify:"
echo "       $BINARY --version"
echo ""
echo "  2. Initialize a project:"
echo "       cd your-project && $BINARY init --ide auto"
echo ""
echo "  3. Diagnose any issues:"
echo "       $BINARY doctor"
echo ""