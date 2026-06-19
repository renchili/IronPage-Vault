#!/usr/bin/env bash
set -euo pipefail

test -s docs/swagger/docs.go
test -s docs/swagger/swagger.yaml

# CI-owned minimal API surface contract. Full API acceptance remains in the
# merge/post-merge regression flow.
grep -q '/healthz' docs/swagger/swagger.yaml
grep -q '/api/auth/login' docs/swagger/swagger.yaml
grep -q '/api/documents' docs/swagger/swagger.yaml
grep -q '/api/admin/backup/run' docs/swagger/swagger.yaml
grep -q '/api/notifications' docs/swagger/swagger.yaml

echo "PASS CI Swagger contract"
