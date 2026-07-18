# Test Effectiveness

This document records the current test-effectiveness boundary. It describes what the repository's checks prove and what still requires executed runtime evidence. It is not a work log or implementation chronology.

## API acceptance

The API acceptance suite must exercise real HTTP requests against the running service and verify resulting PostgreSQL and filesystem state where applicable.

Required coverage includes:

- local login and supported roles;
- positive and negative RBAC paths;
- document upload, retrieval, versions, and workflow transitions;
- terminal Finalized immutability;
- redaction, Bates numbering, and comparison outputs;
- audit records and notification side effects;
- backup and restore artifacts;
- request timestamp, replay, session, and logout behavior;
- rolling failed-login lockout and authentication persistence failures.

Static scripts or route-name checks do not substitute for these runtime flows.

## Acceptance UI

The backend probe UI is served at:

```text
/ui/
```

It is an acceptance-only test aid, not a product frontend. Screenshot evidence proves rendering only. Interaction acceptance must separately verify form submission, invalid-login output, recovery after failure, keyboard submission, focus behavior, and retry behavior.

## Swagger and API documentation

Swagger UI is served at:

```text
/swagger/index.html
```

Generated Swagger artifacts come from route-level Go annotations and are written under `docs/swagger/`. Generated output, `docs/api-spec.md`, and the implemented routes must remain consistent.

## Evidence boundary

A test claim is valid only for the exact revision that produced the test summary, logs, screenshots, traces, and retained artifact. A historical passing run must not be reported as current-HEAD validation.

A complete acceptance result requires executed evidence. Source inspection and static contracts can prevent regressions, but they cannot by themselves prove runtime interaction or state changes.