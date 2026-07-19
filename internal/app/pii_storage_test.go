package app

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPIIStorageSealsPlaintextAndSupportsLookup(t *testing.T) {
	secret := "test-aes-key"
	plain := "Case-001 Privileged Memo"
	sealed, err := sealPII(secret, plain)
	if err != nil {
		t.Fatalf("seal pii: %v", err)
	}
	if !strings.HasPrefix(sealed, encryptedPrefix) {
		t.Fatalf("sealed pii must use encrypted prefix, got %q", sealed)
	}
	if strings.Contains(sealed, plain) {
		t.Fatalf("sealed pii must not contain plaintext")
	}
	opened, err := openPII(secret, sealed, "")
	if err != nil {
		t.Fatalf("open pii: %v", err)
	}
	if opened != plain {
		t.Fatalf("opened pii mismatch")
	}
	left := piiLookupKey(secret, "Admin")
	right := piiLookupKey(secret, " admin ")
	if left != right {
		t.Fatalf("lookup key must be normalized and deterministic")
	}
	if strings.Contains(left, "Admin") || !strings.HasPrefix(left, piiLookupPrefix) {
		t.Fatalf("lookup key must be opaque and version-prefixed, got %q", left)
	}
}

func TestAuditMetadataIsSealed(t *testing.T) {
	sealed, err := sealAuditMetadata("test-aes-key", map[string]interface{}{"username": "admin", "source_ip": "10.0.0.1"})
	if err != nil {
		t.Fatalf("seal audit metadata: %v", err)
	}
	if !strings.HasPrefix(sealed, encryptedPrefix) {
		t.Fatalf("sealed audit metadata must use encrypted prefix, got %q", sealed)
	}
	if strings.Contains(sealed, "admin") || strings.Contains(sealed, "10.0.0.1") {
		t.Fatalf("sealed audit metadata must not expose plaintext")
	}
}

func TestOpenAuditPIIRestoresStructuredFields(t *testing.T) {
	secret := "test-aes-key"
	sourceCipher, err := sealAuditSourceIP(secret, "127.0.0.1")
	if err != nil {
		t.Fatalf("seal source ip: %v", err)
	}
	metadataCipher, err := sealAuditMetadata(secret, map[string]interface{}{"version": 1, "kind": "upload"})
	if err != nil {
		t.Fatalf("seal metadata: %v", err)
	}
	row := auditLogResponse{SourceIPCiphertext: sourceCipher, Metadata: json.RawMessage(`{}`), MetadataCiphertext: metadataCipher}
	if err := openAuditPII(secret, &row); err != nil {
		t.Fatalf("open audit pii: %v", err)
	}
	if row.SourceIP != "127.0.0.1" {
		t.Fatalf("source ip mismatch: %q", row.SourceIP)
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
		t.Fatalf("decode opened metadata: %v", err)
	}
	if metadata["kind"] != "upload" || metadata["version"] != float64(1) {
		t.Fatalf("opened metadata mismatch: %#v", metadata)
	}
}

func TestOpenAuditPIIAcceptsLegacyPlainFields(t *testing.T) {
	row := auditLogResponse{SourceIP: "127.0.0.1", Metadata: json.RawMessage(`{"legacy":true}`)}
	if err := openAuditPII("test-aes-key", &row); err != nil {
		t.Fatalf("open legacy audit pii: %v", err)
	}
	if row.SourceIP != "127.0.0.1" || string(row.Metadata) != `{"legacy":true}` {
		t.Fatalf("legacy audit fields changed unexpectedly")
	}
}
