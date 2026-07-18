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
| Read and update system configuration | Yes | No | No |
| Manage workflow status definitions | Yes | No | No |
| Manage notification templates | Yes | No | No |
| Run, list, and restore backups | Yes | No | No |
| Query audit logs | Yes | No | No |
| Read document metadata and files | Oversight access | Owned documents | Non-Draft documents |
| Upload and batch-upload PDFs | No | Yes | No |
| List document versions | Oversight access | Owned documents | Non-Draft documents |
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

Admin manages identities, configuration dictionaries, workflow definitions, notification templates, and backup operations. Editor manipulates legal documents. Keeping those roles separate prevents infrastructure administration from silently granting document-edit authority.

Admin retains read-only oversight access to document objects but cannot upload, roll back, redact, annotate, apply Bates numbering, transition, or finalize them.

## Enforcement layers

Authorization is enforced at multiple backend layers:

1. route middleware provides the coarse role boundary;
2. core object policy evaluates principal, owner, status, and operation;
3. service and mutation paths enforce workflow and Finalized-state rules; and
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

The workflow state machine independently rejects skipped or invalid transitions.

## Required positive evidence

Acceptance must prove at least:

- Admin can manage users and configuration;
- Admin can query audit records and run backup operations;
- Admin can read document objects for oversight without mutating them;
- owning Editor can upload and perform permitted document operations;
- Reviewer can read and review a non-Draft document; and
- each role can query and acknowledge only its own notifications.

## Required negative evidence

Acceptance must prove at least:

- Reviewer cannot upload, redact, apply Bates numbering, or finalize;
- Reviewer cannot access Draft documents;
- Editor cannot manage users, configuration, audit logs, or backups;
- non-owning Editor cannot access another Editor's document;
- Admin cannot perform Editor or Reviewer document mutations;
- invalid workflow transitions are rejected; and
- every role is denied every mutation after Finalized.

Static policy inspection identifies the intended rules. A runtime RBAC acceptance claim requires executed positive and negative flows tied to the tested revision.
