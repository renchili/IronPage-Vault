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

## Why document list filters are store logic

The document list filter turns a principal into a SQL WHERE clause and query arguments. That output contains database column names and placeholder syntax, so it is persistence adapter logic.

It does not belong in `internal/core`, because core must not emit SQL fragments or know storage schema details. It also should not remain owned by `internal/app`, because HTTP handlers should not own query construction.

The current migration places `DocumentListWhereClause` in `internal/store`. `internal/app` keeps a small wrapper only to preserve current handler calls while larger repository methods are extracted later.

## Why workflow chain rules are core rules

The workflow status chain is domain policy. It decides the valid next legal document status and does not require Echo, SQL, filesystem access, or response formatting.

For that reason `NextWorkflowStatus` and `WorkflowStatusChain` belong in `internal/core`. `internal/app` may temporarily keep a wrapper for handler compatibility, but the real rule should not live in the API adapter layer.

## Why text token parsing is a core rule

Mention parsing decides which local usernames are referenced by an annotation comment. The parser is deterministic text policy and does not require Echo, SQL, or notification persistence.

For that reason `ExtractMentionUsernames` belongs in `internal/core`. The database lookup and notification creation remain outside core because they are persistence and side-effect behavior.

## Why notification cap policy is a core rule

The unread notification cap decides how many old unread records must be marked read before inserting a new notification. The calculation itself is deterministic domain policy and does not require Echo or SQL.

For that reason `NotificationTrimCount` and `MaxUnreadNotifications` belong in `internal/core`. The SQL update that marks old rows as read remains outside core because it is persistence adapter behavior.

## Why crypto and digest helpers are platform code

Encryption and digesting are implementation adapters. They do not decide whether a user may perform an action and they do not map HTTP requests. They provide low-level capabilities used by higher-level workflows.

For that reason AES-GCM string encryption and SHA-256 reader digesting belong in `internal/platform` instead of `internal/app`.

The migration keeps temporary app wrappers so existing handlers can continue calling the old helper names while follow-up PRs move callers directly to `internal/platform`.

## Why PDF helpers are platform code

PDF inspection and strict PDF transforms are infrastructure adapters. They perform filesystem reads and writes, PDF validation, page inspection, digest calculation, rasterized redaction burn-in, and page-visible Bates overlays.

The strict redaction path rewrites page content so sensitive text is not merely hidden by an annotation layer. The strict Bates path renders the page-visible label into a new version. Both transforms are validated by content-level acceptance tests.

These responsibilities do not belong in the API adapter layer. `internal/app` may expose compatibility wrappers, but the underlying implementation belongs in `internal/platform` and `internal/service`.

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
