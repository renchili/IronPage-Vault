#!/usr/bin/env bash
set -u -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAIL=0

check_absent() {
  local name="$1" pattern="$2" file="$3"
  local target="$ROOT_DIR/$file"
  if grep -q "$pattern" "$target"; then
    echo "FAIL api: $name"
    grep -n "$pattern" "$target" || true
    FAIL=$((FAIL+1))
  else
    echo "PASS api: $name"
  fi
}

check_present() {
  local name="$1" pattern="$2" file="$3"
  local target="$ROOT_DIR/$file"
  if grep -q "$pattern" "$target"; then
    echo "PASS api: $name"
  else
    echo "FAIL api: $name"
    FAIL=$((FAIL+1))
  fi
}

check_present "redaction service uses strict entrypoint" "RewritePDFWithRedactionsStrict" internal/service/pdf.go
check_present "bates service uses strict entrypoint" "RewritePDFWithBatesStrict" internal/service/pdf.go
check_present "strict redaction requires pdftoppm" "strict redaction requires pdftoppm" internal/platform/pdf_strict.go
check_present "strict redaction requires python deps" "strict redaction requires python PIL and reportlab" internal/platform/pdf_strict.go
check_present "strict bates requires python deps" "strict Bates numbering requires python pypdf and reportlab" internal/platform/pdf_strict.go
check_present "strict backup rejects non pg_dump_custom" "strict backup requires successful pg_dump_custom" internal/platform/backup_strict.go
check_present "strict backup rejects non tar" "strict backup requires successful tar snapshot" internal/platform/backup_strict.go
check_present "strict restore requires artifact paths" "database_dump_path and file_snapshot_path are required" internal/platform/backup_strict.go
check_absent "service layer must not call legacy redaction fallback" "RewritePDFWithRedactions(input" internal/service/pdf.go
check_absent "service layer must not call legacy bates fallback" "RewritePDFWithBates(input" internal/service/pdf.go

exit "$FAIL"
