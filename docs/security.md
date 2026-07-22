# Security Notes

IronPage Vault is designed for a local air-gapped legal document environment. The security model uses installation-specific local identity, explicit roles/object policy, server-side sessions, protected metadata, mandatory audit side effects, immutable Finalized documents, and code-enforced backup/restore operation boundaries.

## Installation configuration

`scripts/deploy.sh` generates complete runtime configuration into a mode-`0600` file excluded from Git and the Docker context. Database identity, ports, filesystem targets, credentials, JWT/AES material and initial Admin values are installation-specific. The schema does not seed a machine-specific backup path; startup persists the actual generated `BACKUP_DIR`.

Normal mode creates one initial Admin from bootstrap values only while the user table is empty. Acceptance fixtures exist only under explicit acceptance mode and are not embedded in product/browser source. Password verifiers use bcrypt and are sealed before storage; inputs above bcrypt's 72-byte limit are rejected.

## Authentication state

Only failed attempts in the preceding 15 minutes count. The fifth applies a 15-minute lock. Failed-attempt insert/count/lock and `LOGIN_FAILED` audit commit under one user row lock. Successful attempt reset, session creation and `LOGIN` audit commit together. Logout blacklist, session revocation and `LOGOUT` audit commit together. Any required database or audit error fails the request.

Authenticated requests require a fresh timestamp and unique request ID. Blacklist lookup, replay persistence and session activity fail closed. Restore maintenance starts in global middleware before authentication work, so authentication-state writes cannot occur concurrently with database replacement.

## Protected metadata and lookup

Sensitive source values use AES-256-GCM ciphertext columns. Deterministic keys are limited to equality lookup. Audit source IP has a deterministic lookup column plus ciphertext; startup backfills the lookup from existing ciphertext or legacy plaintext. Audit metadata remains encrypted JSON. The Admin route decrypts both fields after a typed query and does not return lookup values or blank compatibility fields as user data.

Annotation comments, notification messages, document titles, identities and redaction geometry/reasons use protected storage paths. Mention lookup uses the encrypted username lookup key rather than plaintext.

Restore lifecycle journals are stored outside `STORAGE_DIR` as AES-256-GCM envelopes in a mode-`0700` directory with mode-`0600` files. Plaintext envelopes, malformed identities and undecryptable records are rejected.

## Mandatory acting-user audit

A successful material database mutation cannot silently lose its audit. The main state change and audit share one transaction. Workflow/finalization also include status history and owner notification; annotation creation includes mention notifications; Admin user/config/template/workflow changes and notification acknowledgement include audit.

The common audit helper rejects an empty actor and no longer converts an empty value to `NULL`. Scheduled backup and startup restore reconciliation use a protected system scheduler principal. That principal has a random, unretained bcrypt verifier and is hidden from the Admin user collection; it exists so automated mutations still have an explicit acting identity and foreign-key integrity.

File-producing redaction and Bates operations keep their database transaction open through verified file generation, remove generated files on failed persistence, and commit version/document/audit state together. Bates sequence reservation is part of that transaction.

## Backup and restore isolation

Unsafe API mutations acquire a shared PostgreSQL advisory lock. Manual and scheduled backup acquire the matching exclusive lock across metadata collection, `pg_dump`, filesystem tar creation, and job/audit persistence. This is the application mutation barrier for the supported single-container deployment.

Restore additionally owns a local maintenance gate before authentication and handler work. New requests fail with `MAINTENANCE_MODE`, active requests drain, concurrent restore is rejected, and the exclusive advisory lock remains held through filesystem replacement, `pg_restore`, lifecycle persistence, and response.

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
