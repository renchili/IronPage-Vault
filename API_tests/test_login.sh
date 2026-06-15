#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${LOGIN_USER:?set LOGIN_USER}"
: "${LOGIN_SECRET:?set LOGIN_SECRET}"
code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/auth/login" -H 'Content-Type: application/json' -H "X-Request-ID: $(reqid)" -d "{\"username\":\"$LOGIN_USER\",\"password\":\"$LOGIN_SECRET\"}")
expect_code "login endpoint" 200 "$code" || FAIL=$((FAIL+1))
json_field token >/tmp/ironpage_token.out || FAIL=$((FAIL+1))
exit "$FAIL"
