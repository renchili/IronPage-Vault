# RBAC Documentation

IronPage Vault supports exactly three roles:

- Admin
- Editor
- Reviewer

The roles are intentionally separate. Admin is a system-management and oversight role, not an automatic document editor.

## Capability matrix

| Capability | Admin | Editor | Reviewer |
|---|---:|---:|---:|
| Login and view own principal | Yes | Yes | Yes |
| Create and list users | Yes | No | No |
| Read and update pagination configuration | Yes | No | No |
| Enable/disable scheduled backup and set interval | Yes | No | No |
| Change deployment-owned backup path | No | No | No |
| Manage workflow status definitions | Yes | No | No |
| Manage notification templates | Yes | No | No |
| Run, list, and restore backups | Yes | No | No |
| Query audit logs | Yes | No | No |
| Read document metadata and files | Oversight access | Owned documents | Non-Draft documents |
| Upload and batch-upload PDFs | No | Yes | No |
| List document versions | Oversight access | Owned documents | Non-Draft documents |
| Create persisted comparison for readable non-Finalized versions | Oversight access | Owned/readable versions | Non-Draft readable versions |
| Roll back versions | No | Owning Editor only | No |
| Propose and confirm redaction | No | Owning Editor only | No |
| Apply Bates numbering | No | Owning Editor only | No |
| Create and disposition annotations | No | No | Reviewable documents only |
| Transition workflow | No | Owned non-Finalized documents | Non-Draft, non-Finalized documents |
| Finalize documents | No | Owning Editor after approval | No |
| Query notifications | Own records | Own records | Own records |
| Acknowledge notifications | Own records | Own records | Own records |

Every document mutation is additionally rejected after Finalized, regardless of role.

## Why Admin is not Editor

Admin manages identities, pagination and backup schedule configuration, workflow definitions, notification templates, and backup operations. Editor manipulates legal documents. Keeping those roles separate prevents infrastructure administration from silently granting document-edit authority.

Admin retains read-only oversight access to document objects and may create a persisted comparison only while both readable source documents are non-Finalized, but cannot upload, roll back, redact, annotate, apply Bates numbering, transition, or finalize them.

`backup.local_volume` is deployment-owned and no role can change it through the API. Admin may update only `backup.schedule_enabled`, `backup.interval`, and the two pagination keys. An Editor or Reviewer receives `403 FORBIDDEN` for backup schedule changes.

## Enforcement layers

Authorization is enforced at multiple backend layers:

1. route middleware provides the coarse role boundary;
2. core object policy evaluates principal, owner, status, and operation;
3. service and mutation paths enforce workflow, version ceiling, and Finalized-state rules; and
4. persistence queries and response mapping preserve object scope and field visibility.

Frontend filtering and route grouping are never sufficient authorization controls.

## Object-level policy

The current object policy is defined in `internal/core/access.go`.

### Read

- Admin may read documents for oversight.
- An owning Editor may read the Editor's own document.
- Reviewer may read a document only after it leaves Draft.
- A non-owning Editor cannot read another Editor's document.

### Edit

Only the owning Editor may mutate document content, and never after Finalized.

### Review

Reviewer may add review activity only after Draft and before Finalized.

### Transition

- Owning Editor may transition a non-Finalized document where the workflow permits.
- Reviewer may transition a non-Draft, non-Finalized document where the workflow permits.
- Admin has no document-transition authority.

The workflow state machine independently rejects skipped or invalid transitions. Finalized status is evaluated before a role/object denial on document mutation handlers, so those routes report `409 DOCUMENT_FINALIZED` for the terminal state.

## Finalized mutation inventory

The backend exposes these existing document mutations after creation:

- rollback;
- redaction proposal;
- redaction confirmation;
- annotation creation;
- annotation disposition;
- Bates numbering;
- persisted comparison creation;
- workflow transition; and
- finalization.

Every one is denied after Finalized. The backend does not expose a document replacement route or a document metadata mutation route, so acceptance must not invent those endpoints. Finalized denials must leave versions/files, redactions, annotations, persisted diffs, history, audit, and notifications unchanged.

## Contextual field visibility

The server opens encrypted fields only after role and object-access checks. Raw ciphertext, deterministic lookup values, and security-control fields are never serialized to any role.

| Field group | Admin | Editor | Reviewer | Masking rule |
|---|---:|---:|---:|---|
| `users.password_hash` | Hidden | Hidden | Hidden | Go JSON tag is `json:"-"`; only authentication code may open the verifier for bcrypt comparison. |
| User identity | Visible on user-admin endpoints and current-principal responses | Visible only for the current principal/JWT context | Visible only for the current principal/JWT context | Stored as `username_ciphertext`; the compatibility username column is a deterministic lookup key, not plaintext. |
| User display name | Visible on user-admin endpoints and current-principal responses | Visible only where the API returns the current principal | Visible only where the API returns the current principal | Stored as `display_name_ciphertext`; the compatibility field stays blank for new writes. |
| Document title | Only when Admin has oversight read access to the document | Visible for owned/readable documents | Visible for reviewer-readable documents | Stored as `title_ciphertext`; opened only after object-level access checks. |
| Document file metadata | Visible on readable version responses | Visible for owned/readable versions | Visible for reviewer-readable versions | Operational path/hash/size/page count are sourced from `document_files`; access follows the parent document. |
| Redaction coordinates and reason | Hidden from list responses | Hidden from list responses | Hidden from list responses | Protected geometry and reason fields are not selected by list queries. |
| Annotation comment | Only when Admin has document read access | Visible only on readable document endpoints | Visible only on readable document endpoints | Stored encrypted; mention parsing uses only the request-local plaintext copy. |
| Notification message | Visible only for the recipient | Visible only for the recipient | Visible only for the recipient | Stored as `message_ciphertext`; opened only for the authenticated recipient. |
| Persisted diff result | Returned only for an authorized compare request | Returned only when both versions are readable | Returned only when both versions are readable | Stored as `document_diffs.result_ciphertext`; no generic diff-list endpoint exposes ciphertext. |
| Audit source IP and metadata | Visible on the Admin audit endpoint | Hidden | Hidden | Stored as ciphertext and opened only in the Admin audit response path. |
| Ciphertext and lookup companion columns | Hidden | Hidden | Hidden | Companion fields use `json:"-"` or response-local decoding and are never API fields. |

## Required positive evidence

Acceptance must prove at least:

- Admin can manage users, pagination, and backup schedule configuration;
- Admin can query audit records and run backup operations;
- Admin can read document objects for oversight without mutating them;
- owning Editor can upload and perform permitted document operations;
- Reviewer can read and review a non-Draft document; and
- each role can query and acknowledge only its own notifications.

## Required negative evidence

Acceptance must prove at least:

- Reviewer cannot upload, redact, apply Bates numbering, or finalize;
- Reviewer cannot access Draft documents;
- Editor cannot manage users, configuration, audit logs, backup schedule, or backups;
- non-owning Editor cannot access another Editor's document;
- Admin cannot perform Editor or Reviewer document mutations;
- invalid workflow transitions are rejected; and
- every role is denied every existing mutation after Finalized.

Static policy inspection identifies the intended rules. A runtime RBAC acceptance claim requires executed positive and negative flows tied to the tested revision.
