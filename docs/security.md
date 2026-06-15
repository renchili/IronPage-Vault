# Security Notes

IronPage Vault is designed for a local air-gapped legal document environment. The security model is based on local accounts, role boundaries, local sessions, audit records, and immutable finalized documents.

## Local Identity

The service uses local username and password login. User metadata is stored in PostgreSQL. Passwords are represented with bcrypt hashes.

The seeded users are intended for local acceptance only and should be changed for any non-demo deployment.

## Sessions

JWT tokens include a token identifier. The database stores session activity so the application can track inactivity and local logout state.

## Request Metadata

Authenticated API calls include a request timestamp and request identifier. These values support request freshness checks, replay detection, and audit correlation.

## Roles

The only roles are Admin, Editor, and Reviewer.

Admin manages users and configuration. Editor manages document operations. Reviewer manages review annotations. Admin does not automatically inherit Editor document rights.

## Finalized Records

Finalized documents represent closed legal records. After finalization, document-changing APIs should reject further changes.

## Error Shape

Business errors should return a consistent JSON response with code, message, details, request ID, and timestamp.

## Local-only Operation

The runtime design avoids external identity providers, remote PDF processors, cloud storage, and external notification providers.

## Acceptance Items

Acceptance should verify:

- successful login
- failed login
- role denial paths
- upload as Editor
- annotation as Reviewer
- Admin configuration access
- request timestamp behavior
- request ID reuse behavior
- logout behavior
- finalized document immutability
- audit rows for mutating actions
