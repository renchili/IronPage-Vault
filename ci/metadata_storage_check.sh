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
check "audit source ip uses ciphertext plus lookup" "grep -q 'source_ip_lookup' migrations/001_schema.sql && grep -q 'source_ip_ciphertext' migrations/001_schema.sql && grep -q 'sourceIPLookup := piiLookupKey' internal/app/domain_events.go && grep -q 'EnsureAuditSourceIPLookups' internal/app/audit_lookup_backfill.go"
check "audit metadata opens structured ciphertext" "grep -q 'metadata_ciphertext' migrations/001_schema.sql && grep -q 'sealAuditMetadata' internal/app/pii_storage.go && grep -q 'openAuditPII' internal/app/audit_filters.go && grep -q 'MetadataCiphertext' internal/repository/audit.go"
check "document diff uses protected result storage" "grep -q 'result_ciphertext TEXT NOT NULL' migrations/004_required_entities_and_backup_schedule.sql && grep -q 'sealAuditMetadata(a.cfg.AESKey, result)' internal/app/workflows.go && grep -q 'DOCUMENT_DIFF_CREATE' internal/app/workflows.go"
check "document diff ciphertext is not exposed as an API field" "! grep -R -q 'json:\"result_ciphertext\"' internal/app internal/repository"
check "redaction reason uses sealed storage" "grep -q 'req.Reason' internal/app/redactions.go && grep -q 'reason, err := encryptString' internal/app/redactions.go"
check "redaction geometry uses sealed storage" "grep -q 'xCipher, err := encryptString' internal/app/redactions.go && grep -q 'heightCipher, err := encryptString' internal/app/redactions.go"
check "redaction legacy numeric geometry is zeroed" "grep -Fq 'VALUES(\$1,\$2,\$3,0,0,0,0,\$4,\$5,\$6,\$7,\$8' internal/app/redactions.go"
check "redaction list excludes detailed fields" "grep -q 'SELECT id,document_id,page,status,created_by' internal/app/redactions.go && ! grep -q 'SELECT id,document_id,page,x,y,width,height,reason' internal/app/redactions.go"
check "annotation comment uses sealed storage" "grep -q 'plainComment := req.Comment' internal/app/annotations.go && grep -q 'commentCipher, err := encryptString' internal/app/annotations.go && grep -q 'decryptString(a.cfg.AESKey, rows\[index\].Comment)' internal/app/annotations.go"
check "annotation mention notifications share the transaction" "grep -q 'notifyMentionedUsersWithExecutor' internal/app/annotations.go && grep -q 'piiLookupKey(a.cfg.AESKey, username)' internal/app/mentions.go && grep -q 'createNotificationWithExecutor' internal/app/mentions.go"
check "password hash uses sealed storage on user create" "grep -q 'sealPasswordHash(a.cfg.AESKey, hash)' internal/app/admin.go"
check "password hash uses sealed storage on seed users" "grep -q 'sealPasswordHash(cfg.AESKey, hash)' internal/app/database.go"
check "login opens sealed password hash before bcrypt compare" "grep -q 'openPasswordHash(a.cfg.AESKey, u.PasswordHash)' internal/app/auth.go && grep -q 'CompareHashAndPassword' internal/app/auth.go"
check "password hash sealed storage test exists" "grep -q 'TestPasswordHashIsSealedAndBcryptCompatible' internal/app/credential_storage_test.go"
check "security document covers protected pii fields" "grep -q 'User identity' docs/security.md && grep -q 'Document title' docs/security.md && grep -q 'Notification message' docs/security.md && grep -q 'Audit source IP' docs/security.md && grep -q 'Audit metadata' docs/security.md"
check "security document covers protected document diffs" "grep -q 'Document comparison' docs/security.md && grep -q 'document_diffs.result_ciphertext' docs/security.md"
check "rbac document covers role field visibility" "grep -q 'Contextual field visibility' docs/rbac.md && grep -q 'Admin' docs/rbac.md && grep -q 'Editor' docs/rbac.md && grep -q 'Reviewer' docs/rbac.md"
check "canonical security document exists" "test -f docs/security.md"
check "security document covers redaction" "grep -q 'Redaction reason' docs/security.md"
check "security document covers annotation" "grep -q 'Annotation comment' docs/security.md"

TOTAL=$((PASS+FAIL))
echo "METADATA STORAGE SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
