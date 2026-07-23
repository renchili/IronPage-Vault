# Security Notes

IronPage Vault is designed for a local air-gapped legal document environment. The security model uses installation-specific local identity, explicit roles/object policy, server-side sessions, protected metadata, mandatory audit side effects, immutable Finalized documents, and code-enforced backup/restore operation boundaries.

## Installation configuration

`scripts/deploy.sh` generates complete runtime configuration into a mode-`0600` file excluded from Git and the Docker context. Database identity, ports, filesystem targets, credentials, JWT/AES material and initial Admin values are installation-specific. The schema does not seed a machine-specific backup path; startup persists the actual generated `BACKUP_DIR`.

Normal mode creates one initial Admin from bootstrap values only while the user table is empty. Acceptance fixtures exist only under explicit acceptance mode and are not embedded in product/browser source. Password verifiers use bcrypt and are sealed before storage; inputs above bcrypt's 72-byte limit are rejected.

## Authentication state

Only failed attempts in the preceding 15 minutes count. The fifth applies a 15-minute lock. Failed-attempt insert/count/lock and `LOGIN_FAILED` audit commit under one user row lock. Successful attempt reset, session creation and `LOGIN` audit commit together. Logout blacklist, session revocation and `LOGOUT` audit commit together. Any required database or audit error fails the request.

Authenticated requests require a fresh timestamp and unique request ID. Blacklist lookup, replay persistence and session activity fail closed. A restore request must complete normal authentication and Admin role validation before its route middleware can activate maintenance. A non-blocking admission mutex prevents concurrent restore authentication; invalid callers release admission without changing service mode.

## Protected metadata and lookup

Protected values use AES-256-GCM envelope strings with the `enc:v1:` prefix. Deterministic `lookup:v1:` values are equality indexes only and are never returned as plaintext data. Compatibility columns may remain for migration but are not the source of truth for new writes.

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

Object/document IDs, page number, workflow status, annotation type/disposition, timestamps, actor IDs, backup/restore job status, audit action type/request ID, notification template key, version numbers, hashes, sizes and local file paths remain plain operational metadata.

Protected metadata insertion and its parent material mutation share the caller transaction where the database can provide one boundary. Annotation comment/audit/mention notifications, workflow history/audit/notification, audit source/metadata, and notification acknowledgement do not commit independently.

Plain request text may exist only in local request memory for validation, deterministic lookup derivation, mention extraction, PDF processing, or encryption. It must not be inserted as protected source-of-truth data. `ci/metadata_storage_check.sh` defines the static storage and API exposure contract.

Restore lifecycle journals are stored outside `STORAGE_DIR` as AES-256-GCM envelopes in a mode-`0700` directory with mode-`0600` files. Plaintext envelopes, malformed identities and undecryptable records are rejected.

## Mandatory acting-user audit

A successful material database mutation cannot silently lose its audit. The main state change and audit share one transaction. Workflow/finalization also include status history and owner notification; annotation creation includes mention notifications; Admin user/config/template/workflow changes and notification acknowledgement include audit.

The common audit helper rejects an empty actor and no longer converts an empty value to `NULL`. Scheduled backup and startup restore reconciliation use a protected system scheduler principal. That principal has a random, unretained bcrypt verifier and is hidden from the Admin user collection; it exists so automated mutations still have an explicit acting identity and foreign-key integrity.

File-producing redaction and Bates operations keep their database transaction open through verified file generation, remove generated files on failed persistence, and commit version/document/audit state together. Bates sequence reservation is part of that transaction.

## Backup and restore isolation

Unsafe API mutations acquire a shared PostgreSQL advisory lock. Manual and scheduled backup acquire the matching exclusive lock across metadata collection, `pg_dump`, filesystem tar creation, and job/audit persistence. This is the application mutation barrier for the supported single-container deployment.

Restore admission first prevents a second restore request from authenticating concurrently. After authentication and Admin role validation, route middleware marks maintenance active, rejects new ordinary requests, drains active requests, and obtains the exclusive advisory lock before filesystem replacement, `pg_restore`, lifecycle persistence, and response. An unauthenticated request cannot activate maintenance.

A crash before a durable restore result creates `Interrupted` with `outcome=unknown`; it is not mislabeled Failed. Resolution requires an Admin acting user and a non-empty verification note.

## PostgreSQL command credentials

`Config.DSN()` remains an in-process database-driver connection string. It is not passed to `pg_dump` or `pg_restore`. PostgreSQL subprocess arguments contain only non-secret port, username, database, paths and operation flags.

For each command the platform adapter creates a short-lived `PGPASSFILE` under `BACKUP_DIR`, sets mode `0600`, escapes password-file delimiters, removes inherited `PGPASSWORD`/`PGPASSFILE`, supplies only the scoped file path to the child environment, and removes the file after command completion. Database passwords therefore do not appear in subprocess argv.

## Configuration integrity

`backup.local_volume` is deployment-owned and cannot be patched through the API. Unknown generic config keys are rejected. The two Admin-managed pagination values are locked and validated as one pair before persistence; they must satisfy `1 <= default <= max <= 100`. Extremely large page numbers are clamped before offset multiplication to avoid integer overflow.

## Roles and Finalized records

The only roles are Admin, Editor and Reviewer. Admin is system management/oversight and does not inherit Editor document mutation rights. Object-level policy remains required after route-level checks.

Finalized documents reject replacement, rollback, redaction, annotation mutation, Bates, workflow movement and metadata mutation regardless of role.

## Error contract

Security and business failures use the standard JSON error envelope. Authentication, audit, history, notification, sequence, version, document, configuration, barrier and restore-state errors cannot become successful responses.

## Static and runtime evidence

Static acceptance judges whether the source, migrations, routes, tests and documentation define these controls without executing them. Missing execution does not alter that static verdict. Existing runtime evidence may demonstrate behavior only for its exact revision and inputs; a static reviewer must not run tests, containers, databases, browsers, deployments or CI to create it.
