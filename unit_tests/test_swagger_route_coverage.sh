#!/usr/bin/env bash
set -euo pipefail

sources=$(cat internal/app/swagger_*.go)

require_route() {
  local method="$1"
  local route="$2"
  if ! printf '%s\n' "$sources" | grep -Fq "@Router $route [$method]"; then
    echo "missing Swaggo route: $method $route" >&2
    exit 1
  fi
}

require_route get /healthz
require_route post /api/auth/login
require_route post /api/auth/logout
require_route get /api/auth/me
require_route post /api/admin/users
require_route get /api/admin/users
require_route get /api/admin/config
require_route patch /api/admin/config/{key}
require_route get /api/admin/workflow-statuses
require_route get /api/admin/notification-templates
require_route patch /api/admin/notification-templates/{key}
require_route post /api/admin/backup/run
require_route get /api/admin/backup/jobs
require_route post /api/admin/backup/restore
require_route get /api/documents
require_route post /api/documents
require_route post /api/documents/batch
require_route post /api/documents/compare
require_route get /api/documents/{id}
require_route get /api/documents/{id}/file
require_route get /api/documents/{id}/versions
require_route post /api/documents/{id}/rollback
require_route post /api/documents/{id}/finalize
require_route post /api/documents/{id}/workflow/transition
require_route post /api/documents/{id}/redactions
require_route get /api/documents/{id}/redactions
require_route post /api/documents/{id}/redactions/{redaction_id}/confirm
require_route post /api/documents/{id}/annotations
require_route get /api/documents/{id}/annotations
require_route patch /api/annotations/{id}/disposition
require_route post /api/documents/{id}/bates
require_route get /api/audit-logs
require_route get /api/notifications
require_route post /api/notifications/{id}/read

echo "PASS Swagger route coverage"
