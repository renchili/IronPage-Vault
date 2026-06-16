#!/usr/bin/env bash
set -euo pipefail

if [ -x unit_tests/test_rules.sh ]; then
  unit_tests/test_rules.sh
fi

go test ./...

if [ -x API_tests/test_api_flow.sh ]; then
  API_tests/test_api_flow.sh
fi

if [ -x API_tests/test_static_review_reject_flows.sh ]; then
  API_tests/test_static_review_reject_flows.sh
fi

if [ -x API_tests/test_acceptance_denials.sh ]; then
  API_tests/test_acceptance_denials.sh
fi

if [ -x API_tests/test_compare_acceptance.sh ]; then
  API_tests/test_compare_acceptance.sh
fi

if [ -x API_tests/test_finalized_immutability.sh ]; then
  API_tests/test_finalized_immutability.sh
fi

if [ -x API_tests/test_compare_self_contained.sh ]; then
  API_tests/test_compare_self_contained.sh
fi

if [ -x API_tests/test_redaction_coordinate_ciphertext.sh ]; then
  API_tests/test_redaction_coordinate_ciphertext.sh
fi

if [ -x API_tests/test_pdf_content_acceptance.sh ]; then
  API_tests/test_pdf_content_acceptance.sh
fi

if [ -x API_tests/test_notification_mention_side_effect.sh ]; then
  API_tests/test_notification_mention_side_effect.sh
fi

if [ -x unit_tests/test_structure_rules.sh ]; then
  unit_tests/test_structure_rules.sh
fi

if [ -x API_tests/test_strict_dependency_failures.sh ]; then
  API_tests/test_strict_dependency_failures.sh
fi

if [ -x API_tests/test_bates_sequence_multi_doc.sh ]; then
  API_tests/test_bates_sequence_multi_doc.sh
fi
