#!/usr/bin/env bash
set -euo pipefail

. API_tests/lib.sh

APP_SERVICE=${APP_SERVICE:-ironpage}
: "${SEED_EDITOR_PASSWORD:?SEED_EDITOR_PASSWORD is required}"

DB_USER_IN_CONTAINER=$(docker compose exec -T "$APP_SERVICE" printenv POSTGRES_USER | tr -d '\r')
DB_NAME_IN_CONTAINER=$(docker compose exec -T "$APP_SERVICE" printenv POSTGRES_DB | tr -d '\r')

psql_exec() {
  docker compose exec -T "$APP_SERVICE" \
    psql -X -v ON_ERROR_STOP=1 -U "$DB_USER_IN_CONTAINER" -d "$DB_NAME_IN_CONTAINER" "$@"
}

psql_scalar() {
  psql_exec -qAt -c "$1" | tr -d '\r'
}

restore_fault_objects() {
  set +e
  psql_exec -q -c 'ALTER TABLE IF EXISTS login_attempts_fault RENAME TO login_attempts;' >/dev/null 2>&1
  psql_exec -q -c 'ALTER TABLE IF EXISTS jwt_blacklist_fault RENAME TO jwt_blacklist;' >/dev/null 2>&1
  psql_exec -q -c 'ALTER TABLE IF EXISTS request_replay_guard_fault RENAME TO request_replay_guard;' >/dev/null 2>&1
  psql_exec -q -c 'ALTER TABLE IF EXISTS sessions_fault RENAME TO sessions;' >/dev/null 2>&1
  psql_exec -q -c 'DROP TRIGGER IF EXISTS reject_session_revoke ON sessions;' >/dev/null 2>&1
  psql_exec -q -c 'DROP FUNCTION IF EXISTS reject_session_revoke();' >/dev/null 2>&1
  set -e
}
trap restore_fault_objects EXIT

expect_error_code() {
  local name="$1" expected="$2"
  expect_json_field "$name" error.code "$expected"
}

editor_id_sql="(SELECT id FROM users WHERE role='Editor' ORDER BY created_at LIMIT 1)"
psql_exec -q -c "DELETE FROM login_attempts WHERE user_id=$editor_id_sql; UPDATE users SET failed_attempts=0, locked_until=NULL WHERE id=$editor_id_sql;"

wrong_password="not-the-editor-password-$RANDOM-$RANDOM"

for attempt in 1 2 3 4; do
  code=$(login_as editor "$wrong_password")
  expect_code "rolling-window initial failure $attempt" 401 "$code"
done

psql_exec -q -c "UPDATE login_attempts SET attempted_at=NOW()-INTERVAL '16 minutes' WHERE user_id=$editor_id_sql;"
old_count=$(psql_scalar "SELECT COUNT(*) FROM login_attempts WHERE user_id=$editor_id_sql AND attempted_at < NOW()-INTERVAL '15 minutes';")
if [ "$old_count" != "4" ]; then
  echo "FAIL api: expected four expired login attempts, got $old_count"
  exit 1
fi
echo "PASS api: expired attempts prepared outside rolling window"

for attempt in 1 2 3 4; do
  code=$(login_as editor "$wrong_password")
  expect_code "rolling-window fresh failure $attempt" 401 "$code"
done

code=$(login_as editor "$wrong_password")
expect_code "rolling-window fifth failure locks account" 423 "$code"
expect_error_code "rolling-window lock error code" ACCOUNT_LOCKED

lock_state=$(psql_scalar "SELECT failed_attempts::text || ':' || (locked_until > NOW())::text FROM users WHERE id=$editor_id_sql;")
if [ "$lock_state" != "5:true" ]; then
  echo "FAIL api: unexpected lock state $lock_state"
  exit 1
fi
echo "PASS api: five in-window failures persisted a 15-minute lock"

code=$(login_as editor "$SEED_EDITOR_PASSWORD")
expect_code "correct password remains blocked during lock" 423 "$code"
expect_error_code "active lock error code" ACCOUNT_LOCKED

psql_exec -q -c "UPDATE users SET locked_until=NOW()-INTERVAL '1 minute' WHERE id=$editor_id_sql;"
code=$(login_as editor "$SEED_EDITOR_PASSWORD")
expect_code "login succeeds after lock expiry" 200 "$code"
editor_token=$(json_field token)
if [ -z "$editor_token" ]; then
  echo "FAIL api: successful login after lock expiry returned no token"
  exit 1
fi

after_success=$(psql_scalar "SELECT failed_attempts::text || ':' || (locked_until IS NULL)::text || ':' || (SELECT COUNT(*) FROM login_attempts WHERE user_id=$editor_id_sql)::text FROM users WHERE id=$editor_id_sql;")
if [ "$after_success" != "0:true:0" ]; then
  echo "FAIL api: successful login did not clear lockout state: $after_success"
  exit 1
fi
echo "PASS api: successful login clears rolling-window state"

psql_exec -q -c 'ALTER TABLE login_attempts RENAME TO login_attempts_fault;'
code=$(login_as editor "$wrong_password")
psql_exec -q -c 'ALTER TABLE login_attempts_fault RENAME TO login_attempts;'
expect_code "failed-login persistence error fails closed" 500 "$code"
expect_error_code "failed-login persistence error code" LOGIN_ATTEMPT_WRITE_ERROR

psql_exec -q -c 'ALTER TABLE login_attempts RENAME TO login_attempts_fault;'
code=$(login_as editor "$SEED_EDITOR_PASSWORD")
psql_exec -q -c 'ALTER TABLE login_attempts_fault RENAME TO login_attempts;'
expect_code "successful-login persistence error fails closed" 500 "$code"
expect_error_code "successful-login persistence error code" LOGIN_STATE_WRITE_ERROR

code=$(login_as editor "$SEED_EDITOR_PASSWORD")
expect_code "fresh token after persistence fault recovery" 200 "$code"
editor_token=$(json_field token)

psql_exec -q -c 'ALTER TABLE jwt_blacklist RENAME TO jwt_blacklist_fault;'
code=$(auth_get "$editor_token" /api/auth/me)
psql_exec -q -c 'ALTER TABLE jwt_blacklist_fault RENAME TO jwt_blacklist;'
expect_code "blacklist read error fails closed" 500 "$code"
expect_error_code "blacklist read error code" AUTH_STATE_READ_ERROR

psql_exec -q -c 'ALTER TABLE request_replay_guard RENAME TO request_replay_guard_fault;'
code=$(auth_get "$editor_token" /api/auth/me)
psql_exec -q -c 'ALTER TABLE request_replay_guard_fault RENAME TO request_replay_guard;'
expect_code "replay persistence error fails closed" 500 "$code"
expect_error_code "replay persistence error code" REPLAY_GUARD_ERROR

psql_exec -q -c 'ALTER TABLE sessions RENAME TO sessions_fault;'
code=$(auth_get "$editor_token" /api/auth/me)
psql_exec -q -c 'ALTER TABLE sessions_fault RENAME TO sessions;'
expect_code "session activity error fails closed" 500 "$code"
expect_error_code "session activity error code" SESSION_UPDATE_ERROR

psql_exec -q <<'SQL'
CREATE OR REPLACE FUNCTION reject_session_revoke()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
  IF NEW.revoked_at IS NOT NULL THEN
    RAISE EXCEPTION 'forced session revoke failure';
  END IF;
  RETURN NEW;
END;
$function$;
CREATE TRIGGER reject_session_revoke
BEFORE UPDATE ON sessions
FOR EACH ROW EXECUTE FUNCTION reject_session_revoke();
SQL

code=$(auth_post_json "$editor_token" /api/auth/logout '{}')
psql_exec -q -c 'DROP TRIGGER reject_session_revoke ON sessions; DROP FUNCTION reject_session_revoke();'
expect_code "logout persistence error fails closed" 500 "$code"
expect_error_code "logout persistence error code" LOGOUT_WRITE_ERROR

code=$(auth_get "$editor_token" /api/auth/me)
expect_code "failed logout transaction leaves session usable" 200 "$code"

code=$(auth_post_json "$editor_token" /api/auth/logout '{}')
expect_code "successful logout" 200 "$code"

code=$(auth_get "$editor_token" /api/auth/me)
expect_code "logged-out token is rejected" 401 "$code"
expect_error_code "logged-out token error code" TOKEN_REVOKED

echo "PASS api-suite: rolling lockout and authentication persistence faults"
