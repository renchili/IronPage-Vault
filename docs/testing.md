# Testing Guide

IronPage Vault separates Go unit tests, repository contracts, stateful API/browser acceptance, local report generation, and complete retained regression evidence. Every result applies only to the exact revision that produced it.

## Test locations

```text
internal/**/*_test.go  colocated Go unit and package tests
tests/contracts/       repository, structure, and generated-contract checks
tests/api/             real HTTP, PostgreSQL, filesystem, PDF, bootstrap, auth, and browser flows
ci/                    serialized verification and full-regression orchestration
```

`run_tests.sh` is the local report entrypoint. `ci/run_full_regression.sh` is the complete regression entrypoint.

## Local report entrypoint

Source-only local checks can be started with:

```bash
bash run_tests.sh
```

Stateful API and browser stages additionally require an already running isolated acceptance service and explicit execution-scoped values:

```text
BASE_URL
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

When those values are absent, `run_tests.sh` records the affected stages as `SKIP`, marks the report `INCOMPLETE`, and exits with status `2`. A skipped stage can never contribute to a local PASS.

The generated report records only the stages actually present in `results.tsv`. A lightweight entrypoint probe may therefore list only Swagger preparation and the probe stage. It must not claim RBAC, PDF, backup, browser, Docker, or full-regression coverage unless those rows were executed.

Generated local files are written under:

```text
artifacts/local-acceptance/
```

## Source and repository contracts

The full regression includes:

- Go package tests with race detection;
- `go vet` and formatting checks;
- repository and structure contracts under `tests/contracts/`;
- generated Swagger and route coverage;
- migration, protected-metadata, PDF, and backup contracts;
- shell syntax parsing;
- documentation consistency and fixed-sensitive-value scanning.

Static checks establish source properties but do not replace real PostgreSQL/API interaction.

## Docker and API acceptance

```bash
bash ci/docker_acceptance.sh
```

The acceptance orchestrator creates independent generated runtime files. It does not depend on a fixed host port, database identity, path, credential, or image-local runtime fallback.

`tests/api/test_bootstrap_restart_docker.sh` covers normal mode on clean generated storage:

1. generate the complete installation configuration;
2. build and start through `scripts/deploy.sh`;
3. log in with the generated initial Admin;
4. verify `/ui/` is absent;
5. rerun without rotating `.env`;
6. restart after removing bootstrap values; and
7. verify the original Admin remains unique and usable.

`tests/api/test_auth_lockout_docker.sh` covers the rolling login window and fail-closed authentication persistence paths. The broader API suite covers RBAC, object access, workflow, finalization, redaction, Bates numbering, comparison, audit, notifications, pagination, error envelopes, backup, restore, and strict dependency failures.

## Browser acceptance

The only browser asset is `public/index.html`, served at `/ui/` only in acceptance mode.

- `tests/api/test_ui_screenshot_acceptance.sh` proves rendering and writes a screenshot manifest.
- `tests/api/test_ui_interaction_acceptance.sh` drives the actual page through Chrome DevTools Protocol and verifies validation, keyboard focus, incorrect credentials, successful login, network failure guidance, recovery, retry, and evidence capture.

A static screenshot is not interaction evidence.

## Complete regression

```bash
bash ci/run_full_regression.sh artifacts/regression
```

The runner is sequential and fail-fast. On the first failed stage it records that stage, writes a failed summary, and exits before later validation starts. A successful run writes:

```text
artifacts/regression/results.tsv
artifacts/regression/summary.json
artifacts/regression/summary.md
artifacts/regression/source-inventory.json
artifacts/regression/logs/
artifacts/regression/ui-interaction/
```

A complete PASS requires `summary.json` to report `overall_status=passed`, every recorded stage to have status zero, and the source inventory to contain no contamination finding.

## GitHub verification safety

`.github/workflows/ci.yml` is the sole workflow. It uses one repository-and-target concurrency group, waits out any remaining ten-minute target cooldown before validation, enforces an auditable failed-revision latch, runs one sequential job, and has no `if: always()` post-failure steps. Successful evidence is uploaded only after the complete regression returns success.

A failed SHA is not automatically replayed. Verification proceeds after a new reviewed commit or an explicit reviewed manual unlock.

## Evidence boundary

A reviewer-written report is a static summary, not test evidence. A historical run proves only its tested SHA. Before declaring runtime acceptance, record:

- tested commit SHA;
- workflow run and job IDs;
- generated summary status;
- artifact name, size, and digest;
- any difference between the evidence revision and the inspected revision; and
- checks that were not executed and why.
