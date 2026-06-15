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

PostgreSQL stores:

- users
- sessions
- replay guard records
- document metadata
- version metadata
- audit logs
- redaction metadata
- annotations
- Bates jobs
- notifications
- configuration entries
- backup jobs

PDF binaries are stored on the local filesystem because the prompt explicitly separates binary assets from database metadata. The database stores file pointers, hashes, size, page count, and version numbers.

This keeps large binary files out of ordinary relational queries while preserving traceability through metadata and audit logs.

## Why roles are strict and discrete

The prompt defines three roles only:

- Admin
- Editor
- Reviewer

The design does not make Admin a super-editor by default. Admin manages the system. Editor manipulates documents. Reviewer reviews and annotates documents.

This prevents accidental privilege expansion. It also makes role-based tests clearer:

- Editor cannot manage users.
- Reviewer cannot upload documents.
- Admin does not automatically bypass document workflow.

## Why business rules live outside route declarations

Route middleware provides a first permission boundary, but sensitive rules must also live in handlers or service logic. For example, Finalized document immutability must be enforced by the document operation itself, not only by route grouping.

This prevents accidental bypass when a handler is reused or a new route is added later.

The implementation is being refactored so pure domain rules live in `internal/core`, not in the API package. `internal/app` should translate HTTP requests into application calls and API responses; it should not own role rules, workflow rules, object-access policy, PDF helpers, crypto helpers, or SQL repositories.

## Why object access policy is a core rule

Object access decides whether a principal can read, edit, review, or transition a specific document. That decision is domain policy, not HTTP behavior.

The policy depends only on small inputs:

- principal user ID
- principal role
- document owner ID
- document status

It does not need Echo, sqlx, PostgreSQL, request headers, or response formatting. For that reason the access policy belongs in `internal/core`.

`internal/app` may still adapt API-owned structs into core policy inputs while the migration is in progress. This compatibility wrapper is temporary. It prevents a large risky rewrite while still moving the real policy implementation out of the API layer.

The SQL list filter remains outside `internal/core`. A WHERE clause is persistence/query-adapter logic, not domain policy. It should eventually move to `internal/store`, not to `internal/core`.

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
