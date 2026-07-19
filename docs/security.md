# Security Notes

IronPage Vault is designed for a local air-gapped legal document environment. The security model uses local identities, explicit role boundaries, server-side session state, protected metadata, audit records, and immutable finalized documents.

## Installation configuration

`scripts/deploy.sh` generates the complete local runtime configuration into a mode-`0600` file excluded from Git and the Docker build context. Database identity, ports, filesystem targets, credentials, JWT material, AES material, and the initial Admin pair are installation-specific. Product code, image configuration, Compose, documentation, and browser assets do not contain fixed runtime credentials or cryptographic keys or provide alternative local runtime defaults.

## Local identity

Normal mode does not create reusable fixture users. On an empty database it creates one initial Admin only from the generated `BOOTSTRAP_ADMIN_USERNAME` and `BOOTSTRAP_ADMIN_PASSWORD`. Once users exist, restart does not overwrite them and bootstrap values are no longer required.

Acceptance fixtures are separate from normal deployment. They are created only when `ACCEPTANCE_MODE=true`, all fixture values are execution-scoped, and bootstrap values are absent. The sole `/ui/` browser probe is mounted only in that mode.

Passwords are bcrypt verifiers sealed before database storage. Inputs intended for bcrypt hashing are rejected when they exceed bcrypt's 72-byte limit.

## Rolling failed-login lockout

Each failed login is stored as a timestamped event. Only events within the preceding 15 minutes count. The fifth in-window failure locks the account for 15 minutes, old events are discarded from the count, an active lock rejects the correct password, and a successful login after expiry clears the event and compatibility state.

Updates for one account are serialized and committed transactionally so concurrent failures cannot bypass the threshold.

## Sessions and logout

JWT tokens include a `jti`, issued-at time, and expiration. PostgreSQL stores server-side session, last activity, absolute expiration, and revocation state.

Authenticated requests require a fresh `X-Request-Timestamp` and unique request ID. The request ID is persisted with the token identifier to reject replay. Session activity is updated only when the session is active and inside its inactivity and absolute-expiration limits.

Blacklist lookup, replay persistence, session activity, successful-login state, and logout revocation fail closed on database errors. Logout writes blacklist and session revocation in one transaction; it does not report `logged_out` after a partial write.

## Protected metadata

Sensitive values use AES-256-GCM protected columns as their source of truth. Compatibility or lookup columns may contain deterministic lookup keys, blanks, or documented migration values, but do not expose protected plaintext through the API.

Role-contextual masking and object-level authorization are enforced by backend policy and service paths, not by the browser probe.

## Roles

The only roles are Admin, Editor, and Reviewer.

- Admin manages local users and system configuration.
- Editor manages document operations.
- Reviewer manages review and annotation operations.

Admin does not automatically inherit Editor document rights.

## Finalized records

Finalized documents are terminal legal records. Replacement upload, rollback, redaction, annotation mutation, Bates numbering, workflow transition, and metadata mutation must be rejected after finalization.

## Error contract

Security and business failures use the standard JSON error envelope with code, message, details, request ID, and timestamp. Authentication-state failures must not become successful responses or ad hoc strings.

## Acceptance evidence

Security acceptance requires executed evidence for:

- generated normal-mode empty-volume Admin initialization and restart without bootstrap values;
- acceptance fixture isolation and normal-mode `/ui/` absence;
- rolling-window expiry and fifth-attempt lock;
- lock expiry and successful-login reset;
- database fault injection for lockout, login state, blacklist, replay, session activity, and logout;
- request timestamp expiry and duplicate request-ID rejection;
- successful logout followed by rejected token reuse;
- role-denial and object-access negative paths; and
- finalized-document immutability and corresponding audit evidence.

Static source inspection alone does not establish these runtime results.
