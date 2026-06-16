#!/usr/bin/env bash
set -u -o pipefail
exec API_tests/test_compare_self_contained.sh
