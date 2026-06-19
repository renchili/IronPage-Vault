#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8080}
BODY=${BODY:-/tmp/ironpage_ci_body.json}

reqid() { echo "ci_req_$(date +%s%N)_$RANDOM"; }
ts() { date -u +%Y-%m-%dT%H:%M:%SZ; }

expect_code() {
  local name="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "PASS ci-smoke: $name"
    return 0
  fi
  echo "FAIL ci-smoke: $name expected=$expected actual=$actual"
  [ -f "$BODY" ] && cat "$BODY" && echo
  return 1
}

expect_json_error_code() {
  local name="$1" expected="$2"
  python3 - "$BODY" "$expected" <<'PY'
import json, sys
body_path, expected = sys.argv[1:]
with open(body_path, encoding='utf-8') as f:
    body = json.load(f)
actual = body.get('error', {}).get('code')
if actual != expected:
    print(f'expected error.code={expected} actual={actual}')
    print(json.dumps(body, indent=2))
    sys.exit(1)
PY
  echo "PASS ci-smoke: $name"
}

curl -fsS "$BASE_URL/healthz" >/dev/null
echo "PASS ci-smoke: healthz"

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" \
  -H "X-Request-ID: $(reqid)" \
  -H "X-Request-Timestamp: $(ts)")
expect_code "unauthenticated document list is rejected" 401 "$code"
expect_json_error_code "unauthenticated document list error envelope" UNAUTHORIZED

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/not-a-real-route" \
  -H "X-Request-ID: $(reqid)" \
  -H "X-Request-Timestamp: $(ts)")
expect_code "unknown route returns not found" 404 "$code"
expect_json_error_code "unknown route error envelope" NOT_FOUND

echo "PASS ci-smoke: neutral runtime contract"
