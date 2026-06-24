#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8080}
OUT_DIR=${IRONPAGE_ACCEPTANCE_REPORT_DIR:-artifacts/local-acceptance}
UI_OUT="$OUT_DIR/ui"
SCREENSHOT="$UI_OUT/ironpage-ui.png"
HTML_COPY="$UI_OUT/index.html"
MANIFEST="$UI_OUT/manifest.json"
REPORT="$UI_OUT/report.html"
BODY=${BODY:-/tmp/ironpage_ui_body.txt}
mkdir -p "$UI_OUT"

browser_bin() {
  if [ -n "${IRONPAGE_BROWSER_BIN:-}" ] && command -v "$IRONPAGE_BROWSER_BIN" >/dev/null 2>&1; then
    echo "$IRONPAGE_BROWSER_BIN"
    return 0
  fi
  for candidate in google-chrome-stable google-chrome chromium chromium-browser chrome; do
    if command -v "$candidate" >/dev/null 2>&1; then
      echo "$candidate"
      return 0
    fi
  done
  return 1
}

expect_http() {
  local name="$1" url="$2" expected="$3" output="$4"
  local code
  code=$(curl -sS -o "$output" -w "%{http_code}" "$url")
  if [ "$code" != "$expected" ]; then
    echo "FAIL ui: $name expected=$expected actual=$code url=$url"
    [ -f "$output" ] && cat "$output" && echo
    return 1
  fi
  echo "PASS ui: $name"
}

expect_http health "$BASE_URL/healthz" 200 "$UI_OUT/health.json"
expect_http ui_page "$BASE_URL/ui/" 200 "$HTML_COPY"

grep -q "IronPage Vault Backend Test UI" "$HTML_COPY"
grep -q "data-testid=\"screenshot-contract\"" "$HTML_COPY"

echo "PASS ui: static page contract"

BROWSER=$(browser_bin) || {
  echo "FAIL ui: no supported browser found. Install google-chrome, chromium, or set IRONPAGE_BROWSER_BIN."
  exit 1
}

set +e
"$BROWSER" \
  --headless=new \
  --no-sandbox \
  --disable-gpu \
  --disable-dev-shm-usage \
  --window-size=1440,1100 \
  --virtual-time-budget=3000 \
  --screenshot="$SCREENSHOT" \
  "$BASE_URL/ui/" >"$UI_OUT/browser.log" 2>&1
status=$?
if [ "$status" -ne 0 ]; then
  "$BROWSER" \
    --headless \
    --no-sandbox \
    --disable-gpu \
    --disable-dev-shm-usage \
    --window-size=1440,1100 \
    --virtual-time-budget=3000 \
    --screenshot="$SCREENSHOT" \
    "$BASE_URL/ui/" >>"$UI_OUT/browser.log" 2>&1
  status=$?
fi
set -e

if [ "$status" -ne 0 ]; then
  echo "FAIL ui: headless screenshot failed with $BROWSER"
  cat "$UI_OUT/browser.log"
  exit "$status"
fi

test -s "$SCREENSHOT"

python3 - "$BASE_URL" "$BROWSER" "$SCREENSHOT" "$HTML_COPY" "$MANIFEST" "$REPORT" <<'PY'
import datetime as dt
import html
import json
import os
import sys
base_url, browser, screenshot, html_copy, manifest, report = sys.argv[1:]
payload = {
    "generated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
    "base_url": base_url,
    "browser": browser,
    "page": f"{base_url}/ui/",
    "screenshot": os.path.relpath(screenshot, os.path.dirname(manifest)),
    "html_copy": os.path.relpath(html_copy, os.path.dirname(manifest)),
    "screenshot_size_bytes": os.path.getsize(screenshot),
}
with open(manifest, "w", encoding="utf-8") as f:
    json.dump(payload, f, indent=2)
with open(report, "w", encoding="utf-8") as f:
    f.write("<!doctype html><html><head><meta charset='utf-8'><title>IronPage UI Screenshot Evidence</title>")
    f.write("<style>body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;margin:32px;} img{max-width:100%;border:1px solid #d0d7de;border-radius:12px;} code{background:#f6f8fa;padding:2px 6px;border-radius:6px;}</style>")
    f.write("</head><body>")
    f.write("<h1>IronPage UI Screenshot Evidence</h1>")
    f.write(f"<p>Generated at <code>{html.escape(payload['generated_at'])}</code> using <code>{html.escape(browser)}</code>.</p>")
    f.write(f"<p>Page: <code>{html.escape(payload['page'])}</code></p>")
    f.write(f"<p>Screenshot bytes: <code>{payload['screenshot_size_bytes']}</code></p>")
    f.write(f"<img src='{html.escape(payload['screenshot'])}' alt='IronPage UI screenshot'>")
    f.write("</body></html>")
PY

echo "PASS ui: screenshot generated at $SCREENSHOT"
echo "PASS ui: evidence manifest generated at $MANIFEST"
echo "PASS ui: evidence report generated at $REPORT"
