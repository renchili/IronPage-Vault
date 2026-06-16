#!/usr/bin/env bash
set -euo pipefail
if [ -x unit_tests/test_rules.sh ]; then
  unit_tests/test_rules.sh
fi
go test ./...
if [ -x API_tests/test_api_flow.sh ]; then
  API_tests/test_api_flow.sh
fi
