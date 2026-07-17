#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8080}
OUT_DIR=${IRONPAGE_UI_EVIDENCE_DIR:-artifacts/regression/ui-interaction}
: "${SEED_EDITOR_PASSWORD:?SEED_EDITOR_PASSWORD is required}"

browser_bin() {
  if [ -n "${IRONPAGE_BROWSER_BIN:-}" ] && command -v "$IRONPAGE_BROWSER_BIN" >/dev/null 2>&1; then
    printf '%s\n' "$IRONPAGE_BROWSER_BIN"
    return 0
  fi
  local candidate
  for candidate in google-chrome-stable google-chrome chromium chromium-browser chrome; do
    if command -v "$candidate" >/dev/null 2>&1; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done
  return 1
}

browser=$(browser_bin) || {
  echo "FAIL ui-suite: no supported Chrome/Chromium binary found" >&2
  exit 1
}

mkdir -p "$OUT_DIR"
python3 API_tests/ui_interaction_cdp.py \
  --browser "$browser" \
  --base-url "$BASE_URL" \
  --username editor \
  --password "$SEED_EDITOR_PASSWORD" \
  --output-dir "$OUT_DIR"
