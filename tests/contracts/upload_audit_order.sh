#!/usr/bin/env bash
set -euo pipefail

if grep -n 'insertAuditRecord(c.Request().Context(), p.UserID, "DOCUMENT_UPLOAD"' internal/app/documents.go; then
  echo "FAIL contract: DOCUMENT_UPLOAD audit must not be written through the raw DB helper inside the upload transaction"
  exit 1
fi

audit_line=$(grep -n 'a.audit(c, p.UserID, "DOCUMENT_UPLOAD"' internal/app/documents.go | head -1 | cut -d: -f1)
commit_line=$(grep -n 'tx.Commit()' internal/app/documents.go | head -1 | cut -d: -f1)

if [ -z "$audit_line" ] || [ -z "$commit_line" ]; then
  echo "FAIL contract: missing upload audit or tx commit line"
  exit 1
fi

if [ "$audit_line" -le "$commit_line" ]; then
  echo "FAIL contract: DOCUMENT_UPLOAD audit must run after document transaction commit"
  exit 1
fi

echo "PASS contract: DOCUMENT_UPLOAD audit runs after document commit"
