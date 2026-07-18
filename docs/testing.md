# Testing Guide

IronPage Vault separates source-level checks, real Docker/API acceptance, browser interaction evidence, and full-regression artifacts. A test claim is valid only for the exact revision named by the generated evidence.

## Test locations

```text
unit_tests/    repository and static contracts
API_tests/     real HTTP, PostgreSQL, filesystem, PDF, bootstrap, auth, and browser flows
ci/            Docker acceptance and full-regression orchestration
```

`run_tests.sh` is the local report entrypoint. `ci/run_full_regression.sh` is the complete regression entrypoint used by reusable CI workflows.

## Unit and static checks

Source-level checks include:

- Go package tests and race tests;
- role, workflow, access, PDF, backup, and protected-metadata contracts;
- migration and repository structure checks;
- generated Swagger contract and route coverage;
- shell syntax;
- documentation consistency and sensitive-value exposure checks.

These checks do not replace running the service against PostgreSQL.

## Docker and API acceptance

Run:

```bash
bash ci/docker_acceptance.sh
```

The Docker acceptance flow builds the single-container image and exercises real HTTP, PostgreSQL, filesystem, PDF, backup, and restore behavior.

Before the broader acceptance fixture flow, `API_tests/test_bootstrap_restart_docker.sh` must verify normal mode on a clean volume:

1. externally supplied bootstrap values create the first Admin;
2. the Admin can log in;
3. `/ui/` is not served;
4. the container is removed without deleting the volumes;
5. bootstrap values are removed;
6. restart succeeds against the same database;
7. the original Admin still logs in and is not duplicated.

`API_tests/test_auth_lockout_docker.sh` must then verify:

- attempts older than 15 minutes do not count in the current rolling window;
- the fifth fresh failure returns `423 ACCOUNT_LOCKED`;
- a correct password remains blocked during the active lock;
- login succeeds after lock expiry;
- successful login clears event rows and compatibility fields;
- failed-attempt, login-state, blacklist, replay, session, and logout database faults fail closed;
- a forced logout write failure rolls back both revocation changes;
- successful logout rejects later token reuse.

The broader API regression continues to cover RBAC, object access, workflow, finalization, redaction, Bates numbering, comparison, audit, notifications, pagination, errors, backup, restore, and strict dependency failure behavior.

## Acceptance browser interaction

The acceptance-only backend probe is served at:

```text
/ui/
```

It is not a product frontend. Normal mode must return 404 for this path.

`API_tests/test_ui_screenshot_acceptance.sh` proves rendering only. It captures a page screenshot and manifest but does not prove user interaction.

`API_tests/test_ui_interaction_acceptance.sh` uses Chrome/Chromium DevTools Protocol through the Python standard library to exercise the actual page. It must verify:

- missing-input validation and focus placement;
- visible invalid state;
- incorrect-credential API error output;
- successful mouse-click login;
- Tab order through username, password, and submit button;
- Enter-key form submission;
- network failure guidance;
- retry after network recovery;
- live status semantics;
- screenshot sequence and `interaction.json` evidence without recording the password.

The full-regression runner passes `IRONPAGE_UI_EVIDENCE_DIR` so these files are retained inside the regression artifact.

## Full regression

Run:

```bash
bash ci/run_full_regression.sh artifacts/regression
```

The generated evidence includes:

```text
artifacts/regression/results.tsv
artifacts/regression/summary.json
artifacts/regression/summary.md
artifacts/regression/logs/
artifacts/regression/ui-interaction/
```

`summary.json` must report `overall_status=passed`, and every defined stage must have status zero. The CI workflow publishes `summary.md` in the Actions job summary and retains the complete artifact; it does not push generated reports directly to protected `main`.

## Evidence boundary

A historical run proves only its tested SHA. A targeted PR job, screenshot, static guard, or reviewer report cannot be presented as a fresh current-HEAD full regression.

Before declaring the project accepted, record:

- tested commit SHA;
- workflow run and job IDs;
- generated summary result;
- artifact name, size, and digest;
- any difference between the tested revision and the current `main`;
- checks not executed and the reason.
