#!/usr/bin/env bash
set -euo pipefail

run_script() {
  local script="$1"
  if [ -f "$script" ]; then
    bash "$script"
  fi
}

load_api_tokens() {
  if [ -s /tmp/ironpage_admin_token.out ]; then
    export ADMIN_TOKEN="$(cat /tmp/ironpage_admin_token.out)"
  fi
  if [ -s /tmp/ironpage_editor_token.out ]; then
    export EDITOR_TOKEN="$(cat /tmp/ironpage_editor_token.out)"
  fi
  if [ -s /tmp/ironpage_reviewer_token.out ]; then
    export REVIEWER_TOKEN="$(cat /tmp/ironpage_reviewer_token.out)"
  fi
}

prepare_swagger() {
  mkdir -p docs/swagger
  printf 'package swagger\n' > docs/swagger/docs.go

  if command -v swag >/dev/null 2>&1; then
    SWAG_BIN="$(command -v swag)" bash scripts/generate_swagger.sh
    return
  fi

  go install github.com/swaggo/swag/cmd/swag@v1.16.4
  SWAG_BIN="$(go env GOPATH)/bin/swag" bash scripts/generate_swagger.sh
}

prepare_swagger

if [ "${IRONPAGE_RUN_TESTS_CONTRACT_PROBE:-}" = "1" ]; then
  test -s docs/swagger/docs.go
  test -s docs/swagger/swagger.yaml
  go test -mod=mod ./internal/core
  echo "PASS run_tests local entrypoint contract"
  exit 0
fi

run_script unit_tests/test_rules.sh

go test -mod=mod ./...

run_script API_tests/test_api_flow.sh
load_api_tokens

run_script API_tests/test_api_contracts.sh
load_api_tokens

run_script API_tests/test_static_review_reject_flows.sh
load_api_tokens

run_script API_tests/test_acceptance_denials.sh
load_api_tokens

run_script API_tests/test_compare_acceptance.sh
load_api_tokens

run_script API_tests/test_finalized_immutability.sh
load_api_tokens

run_script API_tests/test_redaction_coordinate_ciphertext.sh
load_api_tokens

run_script API_tests/test_pdf_content_acceptance.sh
load_api_tokens

run_script API_tests/test_notification_mention_side_effect.sh
load_api_tokens

run_script unit_tests/test_structure_rules.sh
run_script API_tests/test_strict_dependency_failures.sh
run_script API_tests/test_bates_sequence_multi_doc.sh
