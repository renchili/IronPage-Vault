#!/usr/bin/env bash
set -u -o pipefail
bash tests/api/test_document_upload.sh
bash tests/api/test_workflow_transitions.sh
