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
check "pii helper provides deterministic lookup and sealed storage" "grep -q 'piiLookupKey' internal/app/pii_storage.go && grep -q 'lookup:v1:' internal/app/pii_storage.go && grep -q 'sealPII' internal/app/pii_storage.go"
check "username uses lookup key plus ciphertext" "grep -q 'username_ciphertext' migrations/001_schema.sql && grep -q 'piiLookupKey(a.cfg.AESKey, req.Username)' internal/app/admin.go && grep -q 'username_ciphertext' internal/app/auth.go"
check "display name uses ciphertext storage" "grep -q 'display_name_ciphertext' migrations/001_schema.sql && grep -q 'displayCipher' internal/app/admin.go && grep -q 'openUserPII' internal/app/admin.go"
check "document title uses ciphertext storage" "grep -q 'title_ciphertext' migrations/001_schema.sql && grep -q 'titleCipher' internal/app/documents.go && grep -q 'openDocumentPII' internal/app/documents.go"
check "notification message uses ciphertext storage" "grep -q 'message_ciphertext' migrations/001_schema.sql && grep -q 'messageCipher' internal/app/notifications.go && grep -q 'openNotificationPII' internal/app/admin.go"
check "audit source ip uses ciphertext storage" "grep -q 'source_ip_ciphertext' migrations/001_schema.sql && grep -q 'sealAuditSourceIP' internal/app/pii_storage.go && grep -q \"source_ip,source_ip_ciphertext\" internal/app/domain_events.go"
check "audit metadata uses ciphertext storage" "grep -q 'metadata_ciphertext' migrations/001_schema.sql && grep -q 'sealAuditMetadata' internal/app/pii_storage.go && grep -q 'openAuditPII' internal/app/admin.go"
check "redaction reason uses sealed storage" "grep -q 'req.Reason' internal/app/review.go && grep -q 'reason, err := encryptString' internal/app/review.go"
check "redaction geometry uses sealed storage" "grep -q 'xCipher, err := encryptString' internal/app/review.go && grep -q 'heightCipher, err := encryptString' internal/app/review.go"
check "redaction legacy numeric geometry is zeroed" "grep -Fq 'VALUES(\$1,\$2,\$3,0,0,0,0,\$4,\$5,\$6,\$7,\$8' internal/app/review.go"
check "redaction list excludes detailed fields" "grep -q 'SELECT id,document_id,page,status,created_by' internal/app/review.go && ! grep -q 'SELECT id,document_id,page,x,y,width,height,reason' internal/app/review.go"
check "annotation comment uses sealed storage" "grep -q 'req.Comment' internal/app/review.go && grep -q 'comment, err := encryptString' internal/app/review.go"
check "annotation mention path uses local request copy" "grep -q 'plainComment := req.Comment' internal/app/review.go && grep -q 'notifyMentionedUsers(c, plainComment' internal/app/review.go"
check "password hash uses sealed storage on user create" "grep -q 'sealPasswordHash(a.cfg.AESKey, hash)' internal/app/admin.go"
check "password hash uses sealed storage on seed users" "grep -q 'sealPasswordHash(cfg.AESKey, hash)' internal/app/database.go"
check "login opens sealed password hash before bcrypt compare" "grep -q 'openPasswordHash(a.cfg.AESKey, u.PasswordHash)' internal/app/auth.go && grep -q 'CompareHashAndPassword' internal/app/auth.go"
check "password hash sealed storage test exists" "grep -q 'TestPasswordHashIsSealedAndBcryptCompatible' internal/app/credential_storage_test.go"
check "metadata matrix covers protected pii fields" "grep -q 'User identity' docs/metadata-security.md && grep -q 'Document title' docs/metadata-security.md && grep -q 'Notification message' docs/metadata-security.md && grep -q 'Audit source IP' docs/metadata-security.md && grep -q 'Audit metadata' docs/metadata-security.md"
check "role visibility matrix covers all roles" "test -f docs/role-field-visibility.md && grep -q 'Admin' docs/role-field-visibility.md && grep -q 'Editor' docs/role-field-visibility.md && grep -q 'Reviewer' docs/role-field-visibility.md"
check "metadata matrix document exists" "test -f docs/metadata-security.md"
check "metadata matrix covers redaction" "grep -q 'Redaction reason' docs/metadata-security.md"
check "metadata matrix covers annotation" "grep -q 'Annotation comment' docs/metadata-security.md"

TOTAL=$((PASS+FAIL))
echo "METADATA STORAGE SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
