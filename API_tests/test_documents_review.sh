#!/usr/bin/env bash
set -u -o pipefail
if [ -z "${EDITOR_TOKEN:-}" ] || [ -z "${REVIEWER_TOKEN:-}" ]; then
  echo "SKIP api: document review suite requires tokens from login flow"
  exit 0
fi
API_tests/test_document_upload.sh
