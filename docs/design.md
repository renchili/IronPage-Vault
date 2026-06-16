# Design Document

## Why IronPage Vault is a pure backend project

The prompt describes IronPage Vault as the backbone API for legal professionals, paralegals, and compliance teams. The system boundary is therefore the backend API, not a formal web application.

The repository may include a small browser page for manual testing, but that page is only a backend test aid. It is not a production frontend, not a fullstack requirement, and not part of the product deliverable.

This design keeps the project aligned with the required stack: Go, Echo, sqlx, PostgreSQL, and local filesystem storage.

## Why the system is single-container

The prompt requires standalone deployment on an air-gapped machine. The design packages PostgreSQL and the Go API into one container so an evaluator can start the backend system with one command and without any external services.

This is not the usual production recommendation for large distributed deployments, but it fits the stated air-gapped standalone requirement. Persistent data is kept in Docker volumes:

- PostgreSQL data volume
- PDF storage volume
- backup volume

## Why PostgreSQL stores metadata but not PDF binaries

PostgreSQL stores users, sessions, replay guard records, document metadata, version metadata, audit logs, redaction metadata, annotations, Bates jobs, notifications, configuration entries, and backup jobs.

PDF binaries are stored on the local filesystem because the prompt explicitly separates binary assets from database metadata. The database stores file pointers, hashes, size, page count, and version numbers.

This keeps large binary files out of ordinary relational queries while preserving traceability through metadata and audit logs.

## Why roles are strict and discrete

The prompt defines three roles only:

- Admin
- Editor
- Reviewer

The design does not make Admin a super-editor by default. Admin manages the system. Editor manipulates documents. Reviewer reviews and annotates documents.

This prevents accidental privilege expansion.

## Why business rules live outside route declarations

Route middleware provides a first permission boundary, but sensitive rules must also live in handlers or service logic. For example, Finalized document immutability must be enforced by the document operation itself, not only by route grouping.

This prevents accidental bypass when a handler is reused or a new route is added later.

The implementation is being refactored so pure domain rules live in `internal/core`, not in the API package. `internal/app` should translate HTTP requests into application calls and API responses; it should not own role rules, workflow rules, object-access policy, PDF helpers, crypto helpers, or SQL repositories.

## Why object access policy is a core rule

Object access decides whether a principal can read, edit, review, or transition a specific document. That decision is domain policy, not HTTP behavior.

The policy depends only on principal user ID, principal role, document owner ID, and document status. It does not need Echo, sqlx, PostgreSQL, request headers, or response formatting. For that reason the access policy belongs in `internal/core`.

The SQL list filter remains outside `internal/core`. A WHERE clause is persistence/query-adapter logic, not domain policy. It should eventually move to `internal/store`, not to `internal/core`.

## Why workflow chain rules are core rules

The workflow status chain is domain policy. It decides the valid next legal document status and does not require Echo, SQL, filesystem access, or response formatting.

For that reason `NextWorkflowStatus` and `WorkflowStatusChain` belong in `internal/core`. `internal/app` may temporarily keep a wrapper for handler compatibility, but the real rule should not live in the API adapter layer.

## Why crypto and digest helpers are platform code

Encryption and digesting are implementation adapters. They do not decide whether a user may perform an action and they do not map HTTP requests. They provide low-level capabilities used by higher-level workflows.

For that reason AES-GCM string encryption and SHA-256 reader digesting belong in `internal/platform` instead of `internal/app`.

The migration keeps temporary app wrappers so existing handlers can continue calling the old helper names while follow-up PRs move callers directly to `internal/platform`. The target direction is:

```text
service/app caller -> internal/platform crypto/digest adapter
```

not:

```text
service/app caller -> internal/app helper hidden inside API package
```


## Why PDF helpers are platform code

PDF inspection and append-only PDF transform helpers are infrastructure adapters. They perform filesystem reads and writes, PDF header validation, page-marker counting, and digest calculation.

They do not belong in the API adapter layer. `internal/app` may temporarily expose wrapper functions for compatibility, but the real implementation belongs in `internal/platform`.

The current append-only transform remains a placeholder-style implementation and is not forensic redaction or page-visible Bates rendering.

## Why request timestamp and request ID are required

The prompt requires anti-replay behavior with timestamp validation. The system uses:

- `X-Request-Timestamp` to reject stale requests
- `X-Request-ID` to reject duplicate request IDs for the same token
- JWT `jti` to bind replay records to a token/session

This provides deterministic local replay protection without requiring any external identity or security service.

## Why JWT still has server-side session state

A pure stateless JWT cannot enforce inactivity expiration or immediate logout reliably. The design therefore stores session state in PostgreSQL:

- token `jti`
- user ID
- last seen timestamp
