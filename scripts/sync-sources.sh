#!/usr/bin/env bash
# Copies the app source from thrillee/aegisbox and thrillee/aegisbox-portal into src/,
# so Dokploy builds everything from this one repository (no submodules).
#
# Usage: scripts/sync-sources.sh [aegisbox-ref] [portal-ref]
#   Refs default to main. Review the result with git status / git diff, then commit and push.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
AEGISBOX_REF="${1:-main}"
PORTAL_REF="${2:-main}"
GITHUB="${GITHUB_BASE:-https://github.com/thrillee}"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

sync() {
  local repo="$1" ref="$2" dest="$ROOT/src/$1"
  shift 2
  git clone --quiet --filter=blob:none --no-checkout "$GITHUB/$repo.git" "$WORK/$repo"
  local commit
  commit="$(git -C "$WORK/$repo" rev-parse "origin/$ref" 2>/dev/null || git -C "$WORK/$repo" rev-parse "$ref")"
  rm -rf "$dest"
  mkdir -p "$dest"
  # "$@" holds pathspecs to leave out, e.g. build output committed by mistake
  git -C "$WORK/$repo" archive "$commit" -- . "$@" | tar -x -C "$dest"
  echo "$repo $commit" >> "$WORK/versions"
  echo "synced $repo at $commit"
}

sync aegisbox "$AEGISBOX_REF" ':!smpp-gateway'
sync aegisbox-portal "$PORTAL_REF"

{
  echo "# App source copied into src/ by scripts/sync-sources.sh. Do not edit src/ by hand;"
  echo "# change the app repos and sync again."
  cat "$WORK/versions"
} > "$ROOT/SOURCE_VERSIONS"

echo "Done. Review with: git -C \"$ROOT\" status"
