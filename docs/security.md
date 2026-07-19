# Security Notes

IronPage Vault is designed for a local air-gapped legal document environment. The security model uses installation-specific local identity, explicit roles/object policy, server-side sessions, protected metadata, mandatory audit side effects, and immutable Finalized documents.

## Installation configuration

`scripts/deploy.sh` generates complete runtime configuration into a mode-`0600` file excluded from Git and the Docker context. Database identity, ports, filesystem targets, credentials, JWT/AES material and initial Admin values are installation-specific. The schema does not seed a machine-specific backup path; startup persists the actual generated `BACKUP_DIR`.

Normal mode creates one initial Admin from bootstrap values only while the user table is empty. Acceptance fixtures exist only under explicit acceptance mode and are not embedded in product/browser source. Password verifiers use bcrypt and are sealed before storage; inputs above bcrypt's 72-byte limit are rejected.

## Authentication state

Only failed attempts in the preceding 15 minutes count. The fifth applies a 15-minute lock. Failed-attempt insert/count/lock and `LOGIN_FAILED` audit commit under one user row lock. Successful attempt reset, session creation and `LOGIN` audit commit together. Logout blacklist, session revocation and `LOGOUT` audit commit together. Any required database or audit error fails the request.

Authenticated requests require a fresh timestamp and unique request ID. Blacklist lookup, replay persistence and session activity fail closed.

## Protected metadata and lookup

Sensitive source values use AES-256-GCM ciphertext columns. Deterministic keys are limited to equality lookup. Audit source IP has a deterministic lookup column plus ciphertext; startup backfills the lookup from existing ciphertext or legacy plaintext. Audit metadata remains encrypted JSON. The Admin route decrypts both fields after a typed query and does not return lookup values or blank compatibility fields as user data.

Annotation comments, notification messages, document titles, identities and redaction geometry/reasons use protected storage paths. Mention lookup uses the encrypted username lookup key rather than plaintext.

## Mandatory side effects

A successful material database mutation cannot silently lose its audit. The main state change and audit share one transaction. Workflow/finalization also include status history and owner notification; annotation creation includes mention notifications; Admin user/config/template/workflow changes and notification acknowledgement include audit.

File-producing redaction and Bates operations keep their database transaction open through verified file generation, remove generated files on failed persistence, and commit version/document/audit state together. Bates sequence reservation is part of that transaction.

Backup job/audit persistence failure removes generated dump, tar, manifest and metadata. Restore rejects unsafe archive paths and links, stages files, retains a rollback directory, and uses PostgreSQL single-transaction restore. A success response requires Completed restore state and audit.

## Roles and Finalized records

The only roles are Admin, Editor and Reviewer. Admin is system management/oversight and does not inherit Editor document mutation rights. Object-level policy remains required after route-level checks.

Finalized documents reject replacement, rollback, redaction, annotation mutation, Bates, workflow movement and metadata mutation regardless of role.

## Error contract

Security and business failures use the standard JSON error envelope. Authentication, audit, history, notification, sequence, version, document and restore-state errors cannot become successful responses.

## Static and runtime evidence

Static acceptance judges whether the source, migrations, routes, tests and documentation define these controls without executing them. Missing execution does not alter that static verdict. Existing runtime evidence may demonstrate behavior only for its exact revision and inputs; a static reviewer must not run tests, containers, databases, browsers, deployments or CI to create it.
