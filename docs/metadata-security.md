# Metadata Security Matrix

Protected values use AES-256-GCM envelope strings with the `enc:v1:` prefix. Deterministic `lookup:v1:` values are equality indexes only and are never returned as plaintext data. Compatibility columns may remain for migration but are not the source of truth for new writes.

## Protected storage

| Area | Field | Storage behavior | API behavior |
|---|---|---|---|
| Password verifier | `users.password_hash` | bcrypt verifier is sealed before insert | login opens it before bcrypt comparison |
| User identity | `users.username_ciphertext` | ciphertext is source of truth; `users.username` is deterministic lookup | authorized responses open ciphertext |
| User display name | `users.display_name_ciphertext` | ciphertext source; blank compatibility field | authorized responses open ciphertext |
| Document title | `documents.title_ciphertext` | ciphertext source; blank compatibility field | readable document responses open ciphertext |
| Redaction geometry | coordinate ciphertext columns | numeric compatibility fields are zeroed | list responses omit protected geometry |
| Redaction reason | `redaction_proposals.reason` | encrypted before insert | list responses omit reason |
| Annotation comment | `annotations.comment` | encrypted before insert; plaintext exists only in request memory for mention parsing | readable annotation responses decrypt the comment |
| Notification message | `notifications.message_ciphertext` | ciphertext source; blank compatibility field | recipient response opens ciphertext |
| Audit source IP | `audit_logs.source_ip_ciphertext`, `source_ip_lookup` | source IP is encrypted and a deterministic equality key is stored separately; migration/startup backfills lookup for existing rows | Admin filter hashes input; response opens ciphertext or legacy plaintext |
| Audit metadata | `audit_logs.metadata_ciphertext` | structured JSON is encrypted; compatibility JSON remains `{}` for new writes | Admin response decrypts and validates JSON before serialization |

## Plain operational metadata

Object/document IDs, page number, workflow status, annotation type/disposition, timestamps, actor IDs, backup/restore job status, audit action type/request ID, notification template key, version numbers, hashes, sizes and local file paths remain plain operational metadata.

## Transaction rule

Protected metadata insertion and its parent material mutation share the caller transaction where the database can provide one boundary. Annotation comment/audit/mention notifications, workflow history/audit/notification, audit source/metadata, and notification acknowledgement do not commit independently.

## Runtime rule

Plain request text may exist only in local request memory for validation, deterministic lookup derivation, mention extraction, PDF processing, or encryption. It must not be inserted as protected source-of-truth data.

`ci/metadata_storage_check.sh` defines static storage and API exposure contracts for these rules.
