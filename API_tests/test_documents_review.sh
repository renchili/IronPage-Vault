#!/usr/bin/env bash
set -u -o pipefail
bash API_tests/test_document_upload.sh
bash API_tests/test_workflow_transitions.sh
