# Design Document

## Product boundary

IronPage Vault is an air-gapped legal PDF lifecycle backend. The product boundary is the Go/Echo API, PostgreSQL metadata and security state, and local filesystem PDF storage.

`public/index.html` is the only acceptance browser probe. It is not a production frontend and does not change the backend-only product scope.

## Single-container deployment

The standalone deployment packages PostgreSQL and the Go API in one container for an offline machine with no required external service. `scripts/deploy.sh` creates a protected installation file containing the complete local runtime configuration before the image is built.

Database identity, ports, host exposure, API binding, application asset roots, PostgreSQL data roots, product storage roots, credentials, signing material, encryption material, bootstrap identity, and acceptance fixtures are deployment-owned values. The image, Compose file, and Go application do not provide an alternative fixed local configuration.

Fresh configuration accepts only an IPv4 loopback host binding and selects a host port that is not currently accepting loopback TCP connections. Compose remains the final bind authority because port availability can change after the probe.

Compose uses the same generated database identity for PostgreSQL initialization and application access. The entrypoint checks that identity before startup.

## Data ownership

PostgreSQL is the source of truth for users, login attempts, sessions, replay state, documents, versions, workflow history, audit records, protected review metadata, Bates allocation, notifications, configuration, and backup jobs.

PDF binaries and transformed versions are stored on the generated local filesystem target. PostgreSQL records paths, hashes, sizes, page counts, versions, and related audit state.

## Package boundaries

`internal/core` owns deterministic domain policy such as roles, workflow transitions, object access, validation, mention parsing, and notification-cap calculations.

`internal/service` coordinates use cases that combine policy, persistence, and platform operations.

`internal/repository` and `internal/store` own persistence contracts and SQL-facing behavior. SQL fragments and schema knowledge do not belong in core policy.

`internal/platform` owns infrastructure adapters such as encryption, digesting, filesystem operations, strict PDF transforms, backup, and restore.

`internal/app` owns Echo routing, middleware, request/response mapping, runtime assembly, and narrow HTTP adapters.

## Roles and object access

The supported roles are Admin, Editor, and Reviewer. Admin manages identities and system configuration, Editor performs document operations, and Reviewer performs review operations. Admin does not automatically receive Editor document authority.

Role permission and object access are separate decisions. Object access depends on principal, document owner, document status, and requested operation. Backend policy enforces both decisions even when a route is role-gated.

## Document workflow

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Workflow policy is defined in core rules and enforced by mutation paths. Finalized is terminal; every document mutation must reject a Finalized record.

## Redaction, Bates, and comparison

Redaction is a two-phase operation. A proposal stores protected coordinate-bound metadata. Confirmation produces a new PDF version through strict raster burn-in so target content is not merely hidden by an overlay.

Bates numbering creates a page-visible version and allocates auditable sequence state across document sets.

Version comparison extracts structured text and location data to report added, removed, and modified blocks with page and bounding-box information.

## Authentication and lockout

Passwords are compared through bcrypt verifiers stored in protected form. Password inputs are validated against bcrypt's 72-byte limit before hashing. Failed logins are persisted as timestamped `login_attempts` events.

For one user, failed-attempt processing uses a transaction and row lock. Events older than the 15-minute rolling cutoff are removed, the new event is inserted, the current window is counted, and compatibility fields are updated atomically. The fifth in-window failure sets a 15-minute lock. Successful login clears attempt events and lock fields while creating the server-side session in one transaction.

## JWT, session, freshness, and replay

JWTs are locally signed and contain `sub`, role, username, `jti`, issued-at time, and expiration. PostgreSQL session state remains required for inactivity expiration and immediate logout.

Authenticated requests require a fresh `X-Request-Timestamp` and unique `X-Request-ID`. Replay records are bound to the JWT `jti`. Blacklist reads, replay writes, session updates, login-state changes, and logout writes fail closed on database errors. Logout inserts the blacklist record and revokes the session in one transaction.

## Protected metadata

AES-256-GCM columns hold sensitive source values. Deterministic lookup keys, blank compatibility values, or documented legacy migration values may remain in compatibility columns, but protected plaintext is not the source of truth and is not returned through ordinary API responses.

## Audit and notifications

Material mutations create audit records with actor, target, action, request ID, timestamp, and structured metadata. Notifications are local PostgreSQL records. Workflow transitions and annotation mentions create in-app notifications; read acknowledgement and the unread cap require no external service.

## Backup and recovery

Backup success requires both a PostgreSQL custom dump and a local filesystem archive. Restore validates and applies both supplied local artifacts before reporting success. Automated physical WAL-based PITR orchestration is not claimed; the supported boundary is documented in `docs/pitr.md`.

## Verification architecture

Go tests remain colocated with their packages. Stateful acceptance is under `tests/api/`; repository and generated-contract checks are under `tests/contracts/`.

`.github/workflows/ci.yml` is the sole GitHub workflow and performs static acceptance only. A checkout-free admission job collapses active duplicates, applies the target cooldown, evaluates the failed-revision latch from paginated history, and consumes exact one-time unlocks. A later sequential job runs only static syntax, formatting, inventory, documentation, and contract gates. The successful source inventory is retained after all static gates pass.

`ci/run_full_regression.sh`, Docker acceptance, API flows, browser interaction, databases, and deployments remain separate manual or normal-lifecycle execution paths. The static workflow does not call or claim them.

A static reviewer reads source and pre-existing evidence only and must not run project code or CI to fill gaps. A route, screenshot, static contract, reviewer report, or historical artifact is not by itself current runtime acceptance.

GitHub creates a workflow-run object before repository YAML runs. The repository design provides pre-checkout admission and active-run collapse, not platform-level pre-dispatch prevention.
