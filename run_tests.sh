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

run_script unit_tests/test_rules.sh

go test -mod=mod ./...

run_script API_tests/test_api_flow.sh
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
