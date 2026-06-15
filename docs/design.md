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
- expiry timestamp
- revoked timestamp

Each authenticated request updates session activity. Logout records the token in the blacklist and revokes the session.

## Why the workflow is a fixed chain first

The prompt defines a mandatory status chain:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

The implementation starts with this fixed chain because it is the safest and most testable interpretation. Admin-managed workflow definitions are stored in the database so future work can allow configured transitions without changing code.

## Why Finalized is immutable

Finalized means the legal record is closed. The design rejects all mutation attempts after Finalized, including:

- rollback
- redaction proposal
- redaction confirmation
- annotation creation
- annotation disposition update
- Bates numbering
- workflow transition
- metadata mutation

This is a central legal/compliance requirement, not a UI preference.

## Why redaction is two-phase

The prompt distinguishes staged redaction metadata from burn-in confirmation. The design therefore separates:

1. Proposed region metadata.
2. Editor confirmation that creates a new document version.

A visual overlay alone is not treated as final redaction. The prototype records the burn-in as a new PDF version and audit event. A production-grade PDF engine can later replace the local PDF transformation implementation while keeping the same service contract.

## Why audit logs are stored in PostgreSQL

Audit logs need indefinite retention, filtering, and correlation with users, documents, and request IDs. PostgreSQL provides reliable local persistence and query support without requiring external logging systems.

Every mutating business flow should create an audit row with:

- actor
- document when applicable
- action type
- request ID
- source IP
- timestamp
- structured metadata

## Why notifications are in-app only

The prompt requires compliance-grade notification delivery in an air-gapped environment. External email, SMS, Slack, or SaaS delivery would violate the offline constraint.

Therefore the design uses PostgreSQL-backed in-app notification rows. Users query their own notifications and explicitly acknowledge reads.

## Why there is a test UI

The user clarified that this is a pure backend project and the UI is only for testing.

The test UI stays under `public/` and is served from `/ui/manual-test.html` only to help acceptance testers manually call and observe backend flows. It must not introduce a frontend framework, build step, routing layer, or product UI scope.

API tests and backend behavior remain the source of truth.

## Why test data is local

Acceptance must work offline. Sample PDFs and CSV manifests live in `testdata/` so tests can run without downloading files.

## Why Swaggo is supported

The user requested `api-spec.md` and Swaggo support. The project includes Swaggo dependencies and documents the generation command:

```bash
swag init -g cmd/server/main.go -o docs/swagger
```

The Markdown API spec remains the human-readable acceptance document. Generated OpenAPI files should mirror it.
