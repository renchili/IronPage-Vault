#!/usr/bin/env bash
set -euo pipefail

# Post-merge/manual regression runner. This intentionally executes the
# project-owned API acceptance suites after code has merged or during manual
# replay. It is not used as a pre-merge PR gate.

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

command -v curl >/dev/null
command -v python3 >/dev/null
command -v pdftotext >/dev/null
python3 - <<'PY'
import reportlab
import pypdf
import PIL
PY

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

run_script API_tests/test_strict_dependency_failures.sh
load_api_tokens

run_script API_tests/test_bates_sequence_multi_doc.sh
load_api_tokens

run_script API_tests/test_request_guard_edges.sh
