#!/usr/bin/env bash
# install.sh — install free-kiro's SKILL.md bundle into Claude Code /
# OpenCode / Codex CLI / CodeBuddy skills directories.
#
# Mirrors the curl|bash pattern of hindsight.vectorize.io's installer.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/contrib/skills/install.sh | bash
#
# Environment overrides:
#   FREE_KIRO_SKILL_APP        all (default) | claude | opencode | codex | codebuddy
#   FREE_KIRO_SKILL_VERSION    latest (default) | v0.7.0
#   FREE_KIRO_SKILL_FROM       local path or zip URL (skip download)
#   FREE_KIRO_SKILL_DRY_RUN    1 → print plan only, write nothing
#   FREE_KIRO_SKILL_REPO       jingyu525/free-kiro (default)
#   FREE_KIRO_HOME             override home directory (for CI / tests)

set -euo pipefail

REPO="${FREE_KIRO_SKILL_REPO:-jingyu525/free-kiro}"
APP="${FREE_KIRO_SKILL_APP:-all}"
VERSION="${FREE_KIRO_SKILL_VERSION:-}"
FROM="${FREE_KIRO_SKILL_FROM:-}"
DRY_RUN="${FREE_KIRO_SKILL_DRY_RUN:-}"
if [ -n "${FREE_KIRO_HOME:-}" ]; then
  HOME_DIR="$FREE_KIRO_HOME"
elif [ -n "${HOME:-}" ]; then
  HOME_DIR="$HOME"
else
  HOME_DIR="$(eval echo ~"$(id -un)")"
fi

GITHUB_API="https://api.github.com/repos/${REPO}"
GITHUB_DL="https://github.com/${REPO}/releases/download"

SUBDIR="free-kiro"

log() { printf '  %s\n' "$*" >&2; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

# --- resolve version ----------------------------------------------------

if [ -z "$VERSION" ] && [ -z "$FROM" ]; then
  log "fetching latest release tag from GitHub…"
  VERSION=$(curl -fsSL "${GITHUB_API}/releases/latest" \
    | grep -oE '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' \
    | head -n1 \
    | sed -E 's/.*"([^"]*)"/\1/' \
    | sed 's/^v//')
  if [ -z "$VERSION" ]; then
    die "could not resolve latest version from ${GITHUB_API}/releases/latest"
  fi
fi

VERSION="${VERSION#v}"
ASSET="free-kiro-skill_${VERSION}.zip"
URL="${FROM:-${GITHUB_DL}/v${VERSION}/${ASSET}}"

log "version: ${VERSION}"
log "asset:   ${ASSET}"
log "source:  ${URL}"

# --- download + verify (skipped when --from is a local directory) ------

WORK=""
SKIP_DOWNLOAD=0
if [ -d "$FROM" ]; then
  log "using local bundle directory: $FROM"
  WORK="$FROM"
  SKIP_DOWNLOAD=1
elif [ -n "$FROM" ] && [[ "$FROM" =~ ^https?:// ]]; then
  log "fetching custom URL: $FROM"
else
  if [ -n "$DRY_RUN" ]; then
    log "(dry-run) would download ${URL}"
  else
    WORK=$(mktemp -d -t free-kiro-skill.XXXXXX)
    trap 'rm -rf "$WORK"' EXIT
    log "downloading ${ASSET}…"
    curl -fsSL -o "${WORK}/${ASSET}" "$URL" || die "download failed: $URL"
    # Verify SHA256 against SHA256SUMS.
    SUMS_URL="${GITHUB_DL}/v${VERSION}/free-kiro_${VERSION}_SHA256SUMS"
    log "verifying sha256 against ${SUMS_URL}…"
    SUMS=$(curl -fsSL "$SUMS_URL") || die "could not fetch SHA256SUMS"
    EXPECTED=$(printf '%s\n' "$SUMS" | awk -v a="$ASSET" '$2==a {print $1}')
    if [ -z "$EXPECTED" ]; then
      die "no SHA256 entry for $ASSET in SHA256SUMS"
    fi
    GOT=$(shasum -a 256 "${WORK}/${ASSET}" | awk '{print $1}')
    if [ "$EXPECTED" != "$GOT" ]; then
      die "sha256 mismatch: got=$GOT want=$EXPECTED"
    fi
    log "sha256 OK"
    # Extract zip.
    ( cd "$WORK" && unzip -q "${WORK}/${ASSET}" )
  fi
fi

# --- resolve target apps -----------------------------------------------

target_skills_dir() {
  case "$1" in
    claude|claude-code) echo "${HOME_DIR}/.claude/skills/${SUBDIR}" ;;
    opencode)           echo "${HOME_DIR}/.opencode/skills/${SUBDIR}" ;;
    codex)              echo "${HOME_DIR}/.codex/skills/${SUBDIR}" ;;
    codebuddy)          echo "${HOME_DIR}/.codebuddy/skills/${SUBDIR}" ;;
    *) die "unknown app: $1 (supported: claude, opencode, codex, codebuddy)" ;;
  esac
}

apps=()
case "$APP" in
  all)        apps=(claude opencode codex codebuddy) ;;
  claude|claude-code) apps=(claude) ;;
  opencode)   apps=(opencode) ;;
  codex)      apps=(codex) ;;
  codebuddy)  apps=(codebuddy) ;;
  *) die "unknown --app value: $APP" ;;
esac

# --- write per-app ------------------------------------------------------

for a in "${apps[@]}"; do
  dest=$(target_skills_dir "$a")
  if [ -n "$DRY_RUN" ]; then
    log "(dry-run) $a → $dest"
    continue
  fi
  if [ -z "$WORK" ]; then
    log "skipping $a (no bundle to write — dry-run mode)"
    continue
  fi
  mkdir -p "$dest" || die "mkdir failed: $dest"
  # Copy all files from extracted bundle (SKILL.md, skill.json, references/).
  ( cd "$WORK" && tar cf - --exclude="${ASSET}" . ) | ( cd "$dest" && tar xf - )
  log "installed $a → $dest"
done

log "done. verify with: free-kiro skill show"