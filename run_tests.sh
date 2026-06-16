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
