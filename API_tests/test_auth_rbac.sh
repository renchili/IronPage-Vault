#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/healthz")
expect_code "health" 200 "$code" || FAIL=$((FAIL+1))
exit "$FAIL"
