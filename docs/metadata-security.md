# Metadata Security Matrix

This matrix records which stored metadata is protected by encrypted storage and which fields remain plain operational metadata. Protected values use AES-256-GCM envelope strings with the `enc:v1:` prefix. Compatibility/plain columns may remain for lookup keys, routing, or legacy migration fallback, but protected plaintext must be written to ciphertext columns as the source of truth.

## Encrypted storage

| Area | Field | Storage behavior | API behavior |
|---|---|---|---|
| Password hash verifier | `users.password_hash` | bcrypt verifier is AES-256-GCM sealed before insert using `enc:v1:` storage | login opens the sealed verifier before bcrypt comparison; legacy unsealed bcrypt rows remain readable for migration compatibility |
| User identity | `users.username_ciphertext` | username plaintext is AES-256-GCM sealed; `users.username` stores only a deterministic `lookup:v1:` key for login lookup and uniqueness | authorized user/admin responses open the sealed username before JSON serialization |
| User display name | `users.display_name_ciphertext` | display name plaintext is AES-256-GCM sealed; `users.display_name` remains a blank compatibility placeholder for new writes | authorized user/admin responses open the sealed display name before JSON serialization |
| Document title | `documents.title_ciphertext` | title plaintext is AES-256-GCM sealed; `documents.title` remains a blank compatibility placeholder for new writes | authorized document responses open the sealed title before JSON serialization |
| Redaction geometry | `x_ciphertext`, `y_ciphertext`, `width_ciphertext`, `height_ciphertext` | numeric columns are compatibility placeholders; ciphertext columns hold the source-of-truth values | list responses omit geometry |
| Redaction reason | `reason` | encrypted before insert | list responses omit reason |
| Annotation comment | `comment` | encrypted before insert | stored value is never request plaintext; mention extraction uses a local request copy only |
| Notification message | `notifications.message_ciphertext` | notification message plaintext is AES-256-GCM sealed; `notifications.message` remains a blank compatibility placeholder for new writes | authorized recipient responses open the sealed message before JSON serialization |
| Audit source IP | `audit_logs.source_ip_ciphertext` | request source IP is AES-256-GCM sealed; `audit_logs.source_ip` remains a blank compatibility placeholder for new writes | admin audit responses open the sealed IP before JSON serialization |
| Audit metadata | `audit_logs.metadata_ciphertext` | structured metadata JSON is AES-256-GCM sealed; `audit_logs.metadata` remains `{}` as a compatibility placeholder for new writes | admin audit responses open the sealed metadata before JSON serialization |

## Plain operational metadata

The following remain plain because they are identifiers, routing fields, workflow state, or audit control data rather than user/content PII: object IDs, document IDs, page number, workflow status, annotation type, annotation disposition, timestamps, actor IDs, backup job status, audit action type, request ID, notification template key, version numbers, file hashes, file sizes, and local file paths.

## Runtime rule

Plain request text may exist only in local request memory for validation, lookup-key derivation, mention extraction, PDF processing, or encryption. It must not be inserted as the source-of-truth value for protected metadata.

## Contract

`ci/metadata_storage_check.sh` validates the storage and API exposure rules above.
