# Contextual Role-Based Field Visibility Matrix

This matrix defines which API roles may see decrypted fields after column-level AES storage is opened by the server. Raw ciphertext and security-control fields are never serialized to any role.

| Field group | Admin | Editor | Reviewer | Masking rule |
|---|---:|---:|---:|---|
| `users.password_hash` | Hidden | Hidden | Hidden | Go JSON tag is `json:"-"`; only authentication code may open the verifier for bcrypt comparison. |
| `users.username` | Visible on user admin endpoints and current principal responses | Visible only for the current principal/JWT context | Visible only for the current principal/JWT context | Stored as `username_ciphertext`; `username` DB column is a deterministic lookup key, not plaintext. |
| `users.display_name` | Visible on user admin endpoints and login/current principal responses | Visible only where the API returns the current principal | Visible only where the API returns the current principal | Stored as `display_name_ciphertext`; compatibility column stays blank for new writes. |
| `documents.title` | Not exposed through Admin-only system endpoints unless Admin is the owner/reader | Visible for owned/editor-readable documents | Visible for reviewer-readable documents | Stored as `title_ciphertext`; opened only after object-level access checks. |
| Redaction coordinates/reason | Hidden from list responses | Hidden from list responses | Hidden from list responses | Coordinates are stored in ciphertext columns and list queries select only id/document/page/status/actor/timestamp. |
| Annotation comment | Not exposed through Admin-only system endpoints | Visible only on document endpoints the Editor can read | Visible only on document endpoints the Reviewer can read | Stored encrypted; mention notification uses a local plaintext request copy only. |
| Notification message | Hidden unless recipient | Visible only for recipient's own notification list | Visible only for recipient's own notification list | Stored as `message_ciphertext`; opened only in `/notifications` for the authenticated user. |
| Audit source IP and metadata | Visible on Admin audit endpoint | Hidden | Hidden | Stored as `source_ip_ciphertext` and `metadata_ciphertext`; opened only in Admin audit-log response path. |
| Ciphertext companion columns | Hidden | Hidden | Hidden | All companion fields use `json:"-"` or map-local response decoding and are never API fields. |
