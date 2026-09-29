#!/usr/bin/env bash
# compute-skill-sha.sh — populate sha256 map in skill.json with hashes of
# every file in the skill bundle. Idempotent: re-running rewrites the map.
#
# Called as a goreleaser `before` hook:
#   cmd: ./contrib/skills/scripts/compute-skill-sha.sh
#
# The script computes sha256 for every file under SKILL_SRC (default:
# contrib/skills/free-kiro) and rewrites the `sha256` key in skill.json.
# Files are referenced relative to SKILL_SRC root (matches what the
# installer writes to each app's skills/<name>/ directory).

set -euo pipefail

SKILL_SRC="${SKILL_SRC:-contrib/skills/free-kiro}"

if [ ! -d "$SKILL_SRC" ]; then
  echo "error: $SKILL_SRC does not exist" >&2
  exit 1
fi

MANIFEST="$SKILL_SRC/skill.json"
if [ ! -f "$MANIFEST" ]; then
  echo "error: $MANIFEST not found" >&2
  exit 1
fi

# Collect files: SKILL.md, references/*.md. Exclude skill.json from the
# sha256 map — its content includes the map itself, so its hash would
# differ after we splice the new map in (self-referential chicken/egg).
mapfile -t FILES < <(cd "$SKILL_SRC" && find . -type f \( -name '*.md' -o -name '*.sh' \) ! -path './skill.json' | LC_ALL=C sort)

# Build the sha256 map with jq (present on every CI image + macOS dev box).
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

{
  echo '{'
  first=1
  for f in "${FILES[@]}"; do
    # Strip leading "./" for prettier keys.
    key="${f#./}"
    hash=$(shasum -a 256 "$SKILL_SRC/$f" | awk '{print $1}')
    if [ "$first" -eq 1 ]; then
      first=0
    else
      echo ','
    fi
    printf '  %s' "\"$key\": \"$hash\""
  done
  echo ''
  echo '}'
} > "$TMP"

# Splice the new sha256 map into skill.json, preserving everything else.
# jq's --argfile / --slurpfile avoids loading the file twice.
jq --slurpfile sha "$TMP" '.sha256 = $sha[0]' "$MANIFEST" > "$MANIFEST.tmp"
mv "$MANIFEST.tmp" "$MANIFEST"

# Re-format with 2-space indent so committed diffs stay clean.
jq . "$MANIFEST" > "$MANIFEST.tmp"
mv "$MANIFEST.tmp" "$MANIFEST"

echo "computed sha256 for ${#FILES[@]} files in $SKILL_SRC"