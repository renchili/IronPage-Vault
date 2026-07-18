#!/usr/bin/env bash
set -u -o pipefail

: "${BASE_URL:?BASE_URL is required}"
BODY=${BODY:-/tmp/ironpage_api_body.json}

load_saved_tokens() {
  if [ -z "${ADMIN_TOKEN:-}" ] && [ -s /tmp/ironpage_admin_token.out ]; then
    export ADMIN_TOKEN="$(cat /tmp/ironpage_admin_token.out)"
  fi
  if [ -z "${EDITOR_TOKEN:-}" ] && [ -s /tmp/ironpage_editor_token.out ]; then
    export EDITOR_TOKEN="$(cat /tmp/ironpage_editor_token.out)"
  fi
  if [ -z "${REVIEWER_TOKEN:-}" ] && [ -s /tmp/ironpage_reviewer_token.out ]; then
    export REVIEWER_TOKEN="$(cat /tmp/ironpage_reviewer_token.out)"
  fi
}

load_saved_tokens

reqid() { echo "req_$(date +%s%N)_$RANDOM"; }
ts() { date -u +%Y-%m-%dT%H:%M:%SZ; }

json_field() {
  python3 -c 'import json,sys
with open(sys.argv[1]) as f: cur=json.load(f)
for part in sys.argv[2].split("."):
    if "[" in part and part.endswith("]"):
        name,idx=part[:-1].split("["); cur=cur[name][int(idx)]
    else:
        cur=cur[part]
if isinstance(cur, bool):
    print(str(cur).lower())
elif cur is None:
    print("")
else:
    print(cur)' "$BODY" "$1" 2>/dev/null
}

expect_code() {
  local name="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then echo "PASS api: $name"; return 0; fi
  echo "FAIL api: $name expected=$expected actual=$actual"
  [ -f "$BODY" ] && cat "$BODY" && echo
  return 1
}

expect_json_field() {
  local name="$1" field="$2" expected="$3"
  local actual
  actual="$(json_field "$field")"
  if [ "$actual" = "$expected" ]; then echo "PASS api: $name"; return 0; fi
  echo "FAIL api: $name field=$field expected=$expected actual=$actual"
  [ -f "$BODY" ] && cat "$BODY" && echo
  return 1
}

expect_json_nonempty() {
  local name="$1" field="$2"
  local actual
  actual="$(json_field "$field")"
  if [ -n "$actual" ]; then echo "PASS api: $name"; return 0; fi
  echo "FAIL api: $name field=$field empty"
  [ -f "$BODY" ] && cat "$BODY" && echo
  return 1
}

login_as() {
  local user="$1" password="$2"
  curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/auth/login" \
    -H 'Content-Type: application/json' \
    -H "X-Request-ID: $(reqid)" \
    -d "{\"username\":\"$user\",\"password\":\"$password\"}"
}

auth_get() {
  curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL$2" \
    -H "Authorization: Bearer $1" \
    -H "X-Request-ID: $(reqid)" \
    -H "X-Request-Timestamp: $(ts)"
}

auth_post_json() {
  curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL$2" \
    -X POST \
    -H "Authorization: Bearer $1" \
    -H 'Content-Type: application/json' \
    -H "X-Request-ID: $(reqid)" \
    -H "X-Request-Timestamp: $(ts)" \
    -d "$3"
}

auth_patch_json() {
  curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL$2" \
    -X PATCH \
    -H "Authorization: Bearer $1" \
    -H 'Content-Type: application/json' \
    -H "X-Request-ID: $(reqid)" \
    -H "X-Request-Timestamp: $(ts)" \
    -d "$3"
}
