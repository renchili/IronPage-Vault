# Design Document

## Product boundary

IronPage Vault is an air-gapped legal PDF lifecycle backend. The product boundary is the Go/Echo API, PostgreSQL metadata and security state, and local filesystem PDF storage.

Files under `public/` provide an acceptance-only browser probe. They are not a production frontend and do not change the backend-only product scope.

## Single-container deployment

The standalone deployment packages PostgreSQL and the Go API in one container because the target environment is one offline machine without required external services. Persistent state is separated into PostgreSQL data, PDF storage, and backup volumes.

Runtime credentials, signing material, encryption material, bootstrap identity values, and acceptance fixture values are supplied by the deployment environment. They are not image defaults.

## Data ownership

PostgreSQL is the source of truth for users, login attempts, sessions, replay state, documents, versions, workflow history, audit records, protected review metadata, Bates allocation, notifications, configuration, and backup jobs.

PDF binaries and transformed versions are stored on the local filesystem. PostgreSQL records their paths, hashes, sizes, page counts, versions, and related audit state.

## Package boundaries

`internal/core` owns deterministic domain policy such as roles, workflow transitions, object access, validation, mention parsing, and notification-cap calculations.

`internal/service` coordinates use cases that combine policy, persistence, and platform operations.

`internal/repository` and `internal/store` own persistence contracts and SQL-facing behavior. SQL fragments and schema knowledge do not belong in core policy.

`internal/platform` owns infrastructure adapters such as encryption, digesting, filesystem operations, strict PDF transforms, backup, and restore.

`internal/app` owns Echo routing, middleware, request/response mapping, runtime assembly, and narrow adapter functions required by handlers. A compatibility wrapper in `internal/app` does not move ownership of the underlying domain or platform rule back into the HTTP layer.

## Roles and object access

The supported roles are Admin, Editor, and Reviewer. Admin manages identities and system configuration, Editor performs document operations, and Reviewer performs review operations. Admin does not automatically receive Editor document authority.

Role permission and object access are separate decisions. Object access depends on the principal, document owner, document status, and requested operation. Backend policy enforces both decisions even when a route is already role-gated.

## Document workflow

The default lifecycle is:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Workflow policy is defined in core rules and enforced by mutation paths. Finalized is terminal; every document mutation must reject a Finalized record.

## Redaction, Bates, and comparison

Redaction is a two-phase operation. A proposal stores protected coordinate-bound metadata. Confirmation produces a new PDF version through strict raster burn-in so the underlying target content is not merely hidden by an overlay.

Bates numbering creates a new page-visible version and allocates auditable sequence state across document sets.

Version comparison extracts structured text and location data to report added, removed, and modified blocks with page and bounding-box information.

## Authentication and lockout

Passwords are compared through bcrypt verifiers stored in protected form. Failed logins are persisted as timestamped `login_attempts` events.

For one user, failed-attempt processing uses a transaction and row lock. Events older than the 15-minute rolling cutoff are removed, the new event is inserted, the current window is counted, and compatibility fields on `users` are updated atomically. The fifth in-window failure sets a 15-minute lock. Successful login clears attempt events and lock fields while creating the server-side session in one transaction.

## JWT, session, freshness, and replay

JWTs are locally signed and contain `sub`, role, username, `jti`, issued-at time, and expiration. PostgreSQL session state is still required for inactivity expiration and immediate logout.

Authenticated requests must provide a fresh `X-Request-Timestamp` and unique `X-Request-ID`. Replay records are bound to the JWT `jti`. Session activity updates require an active, unrevoked, unexpired session.

Blacklist reads, replay writes, session updates, successful-login state, and logout writes fail closed on database errors. Logout inserts the blacklist record and revokes the session in one transaction.

## Protected metadata

AES-256-GCM protected columns hold sensitive source values. Deterministic lookup keys, blank compatibility values, or documented legacy migration values may remain in compatibility columns, but protected plaintext is not the source of truth and is not returned through ordinary API responses.

## Audit and notifications

Material mutations create audit records with actor, target, action, request ID, timestamp, and structured metadata. Audit queries support user, document, action, and date filtering.

Notifications are local database records. Workflow transitions and annotation mentions create in-app notifications; read acknowledgement and the unread cap are enforced without external delivery services.

## Backup and recovery

Backup success requires both a PostgreSQL custom dump and a local filesystem archive. Restore validates and applies the supplied local artifacts before reporting success. Automated physical WAL-based PITR orchestration is not claimed; the supported recovery boundary is documented separately in `docs/pitr.md`.

## Acceptance evidence boundary

A route, screenshot, static guard, or historical artifact is not by itself current project acceptance. Product claims must cite implementation paths and executed evidence tied to the tested revision.

The full-regression workflow publishes its generated summary in the Actions job summary and retains the complete artifact. It does not write generated reports directly to the protected `main` branch.
