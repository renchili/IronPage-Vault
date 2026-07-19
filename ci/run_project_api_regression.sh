#!/usr/bin/env bash
set -euo pipefail

# Full-regression API runner. It executes the project-owned acceptance suites
# only inside the serialized full-regression workflow/container.
run_script() {
  local script="$1"
  if [ ! -f "$script" ]; then
    echo "ERROR: required API acceptance script is missing: $script" >&2
    exit 1
  fi
  bash "$script"
}

load_api_tokens() {
  if [ -s /tmp/ironpage_admin_token.out ]; then export ADMIN_TOKEN="$(cat /tmp/ironpage_admin_token.out)"; fi
  if [ -s /tmp/ironpage_editor_token.out ]; then export EDITOR_TOKEN="$(cat /tmp/ironpage_editor_token.out)"; fi
  if [ -s /tmp/ironpage_reviewer_token.out ]; then export REVIEWER_TOKEN="$(cat /tmp/ironpage_reviewer_token.out)"; fi
}

command -v curl >/dev/null
command -v python3 >/dev/null
command -v pdftotext >/dev/null
python3 - <<'PY'
import reportlab
import pypdf
import PIL
PY

run_script tests/api/test_api_flow.sh
load_api_tokens
run_script tests/api/test_api_contracts.sh
load_api_tokens
run_script tests/api/test_static_review_reject_flows.sh
load_api_tokens
run_script tests/api/test_acceptance_denials.sh
load_api_tokens
run_script tests/api/test_compare_acceptance.sh
load_api_tokens
run_script tests/api/test_finalized_immutability.sh
load_api_tokens
run_script tests/api/test_redaction_coordinate_ciphertext.sh
load_api_tokens
run_script tests/api/test_pdf_content_acceptance.sh
load_api_tokens
run_script tests/api/test_notification_mention_side_effect.sh
load_api_tokens
run_script tests/api/test_strict_dependency_failures.sh
load_api_tokens
run_script tests/api/test_bates_sequence_multi_doc.sh
load_api_tokens
run_script tests/api/test_request_guard_edges.sh
