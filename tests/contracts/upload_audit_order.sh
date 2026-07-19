#!/usr/bin/env bash
set -euo pipefail

audit_line=$(grep -n 'auditWithExecutor(c, tx, p.UserID, "DOCUMENT_UPLOAD"' internal/app/documents.go | head -1 | cut -d: -f1)
commit_line=$(grep -n 'tx.Commit()' internal/app/documents.go | head -1 | cut -d: -f1)
cleanup_line=$(grep -n 'os.RemoveAll(dir)' internal/app/documents.go | head -1 | cut -d: -f1)

if [ -z "$audit_line" ] || [ -z "$commit_line" ] || [ -z "$cleanup_line" ]; then
  echo "FAIL contract: upload must define transactional audit, commit, and orphan-directory cleanup"
  exit 1
fi

if [ "$audit_line" -ge "$commit_line" ]; then
  echo "FAIL contract: DOCUMENT_UPLOAD audit must be inserted before the shared transaction commit"
  exit 1
fi

if grep -q 'a.audit(c, p.UserID, "DOCUMENT_UPLOAD"' internal/app/documents.go; then
  echo "FAIL contract: upload must not use a post-commit best-effort audit helper"
  exit 1
fi

echo "PASS contract: upload document/version/audit share one transaction and failed persistence removes the file directory"
