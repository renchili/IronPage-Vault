#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

check() {
  local name="$1"
  local cmd="$2"
  if bash -lc "$cmd" >/tmp/ironpage_metadata_storage.out 2>&1; then
    echo "PASS metadata-storage: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL metadata-storage: $name"
    cat /tmp/ironpage_metadata_storage.out
    FAIL=$((FAIL+1))
  fi
}

check "sealed string helper exists" "grep -q 'cipher.NewGCM' internal/platform/crypto.go && grep -q 'enc:v1:' internal/platform/crypto.go"
check "redaction reason uses sealed storage" "grep -q 'req.Reason' internal/app/review.go && grep -q 'reason, err := encryptString' internal/app/review.go"
check "redaction geometry uses sealed storage" "grep -q 'xCipher, err := encryptString' internal/app/review.go && grep -q 'heightCipher, err := encryptString' internal/app/review.go"
check "redaction legacy numeric geometry is zeroed" "grep -q 'VALUES($1,$2,$3,0,0,0,0,$4,$5,$6,$7,$8' internal/app/review.go"
check "redaction list excludes detailed fields" "grep -q 'SELECT id,document_id,page,status,created_by' internal/app/review.go && ! grep -q 'SELECT id,document_id,page,x,y,width,height,reason' internal/app/review.go"
check "annotation comment uses sealed storage" "grep -q 'req.Comment' internal/app/review.go && grep -q 'comment, err := encryptString' internal/app/review.go"
check "annotation mention path uses local request copy" "grep -q 'plainComment := req.Comment' internal/app/review.go && grep -q 'notifyMentionedUsers(c, plainComment' internal/app/review.go"
check "metadata matrix document exists" "test -f docs/metadata-security.md"
check "metadata matrix covers redaction" "grep -q 'Redaction reason' docs/metadata-security.md"
check "metadata matrix covers annotation" "grep -q 'Annotation comment' docs/metadata-security.md"

TOTAL=$((PASS+FAIL))
echo "METADATA STORAGE SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
