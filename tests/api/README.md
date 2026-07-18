# API Acceptance Tests

This directory contains stateful HTTP and browser acceptance flows. The scripts exercise local login, authentication state, RBAC denials, document lifecycle, PDF processing, audit records, notifications, backup/restore, and recovery paths.

Run the complete local entrypoint from the repository root:

```bash
bash run_tests.sh
```

For Docker-backed complete evidence, use:

```bash
bash ci/run_full_regression.sh artifacts/regression
```

Individual scripts require an explicit `BASE_URL` and the relevant execution-scoped credentials or tokens. They do not contain a fixed server address or credential.
