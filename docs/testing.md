# Testing Guide

IronPage Vault separates static repository acceptance, Go unit tests, repository contracts, stateful API/browser acceptance, local report generation, and complete retained regression evidence. Every result applies only to the exact revision that produced it.

## Test locations

```text
internal/**/*_test.go  colocated Go unit and package tests
tests/contracts/       repository, structure, and generated-contract checks
tests/api/             real HTTP, PostgreSQL, filesystem, PDF, bootstrap, auth, and browser flows
ci/                    static workflow contracts and manual full-regression helpers
```

`run_tests.sh` is the local report entrypoint. `ci/run_full_regression.sh` is the complete regression entrypoint. Neither is called by the GitHub static-acceptance workflow.

## Static reviewer boundary

A static acceptance reviewer reads source and pre-existing evidence only. The reviewer must not run tests, scripts, generators, builds, containers, databases, browsers, deployments, or CI to fill evidence gaps.

When no current execution artifact exists, the corresponding runtime or interaction requirement is `NOT VERIFIED`. Static source presence does not become runtime evidence.

## Local report entrypoint

The normal local command is:

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

The generated report records only the stages actually present in `results.tsv`. A lightweight entrypoint probe may list only Swagger preparation and the probe stage. It must not claim RBAC, PDF, backup, browser, Docker, or full-regression coverage unless those rows were executed.

Generated local files are written under:

```text
artifacts/local-acceptance/
```

## Static GitHub acceptance

`.github/workflows/ci.yml` is the sole GitHub workflow and is limited to static acceptance.

Admission behavior:

- one target key is shared by pull request, merge-group, `main` push, and manual events;
- target concurrency uses `cancel-in-progress: true` to collapse superseded active runs;
- admission runs before checkout and repository-controlled code;
- denied admission is cancelled immediately rather than sleeping in a runner;
- workflow history is fully paginated;
- completed non-cancelled runs enforce a ten-minute target cooldown;
- failed target/revision pairs remain latched;
- ordinary rerun attempts are denied;
- a manual unlock must name the exact target and failed run, include a reviewed reason, and is consumed once.

Static gates then check workflow syntax, shell syntax, Python syntax, Go formatting, source inventory, documentation consistency, repository/structure contracts, backup contracts, metadata contracts, and Swagger route coverage.

The successful source manifest is written to:

```text
artifacts/static-acceptance/source-inventory.json
```

It is uploaded only after all static gates succeed. It records the commit, every tracked file, mode, size, SHA-256, contamination findings, path-hygiene findings, and explicit naming exceptions.

GitHub creates a workflow-run object before repository YAML can execute. The repository proves pre-checkout admission and active-run collapse, not literal pre-dispatch prevention. A stricter requirement needs separate platform-level evidence.

## Source and repository contracts

Static and full-regression helpers cover:

- Go formatting and package-oriented source checks;
- repository and structure contracts under `tests/contracts/`;
- generated Swagger route coverage;
- migration, protected-metadata, PDF, and backup contracts;
- shell syntax parsing;
- documentation consistency and fixed-sensitive-value scanning;
- tracked source contamination and path naming.

Static checks establish source properties but do not replace real PostgreSQL/API interaction.

## Docker and API acceptance

```bash
bash ci/docker_acceptance.sh
```

The acceptance orchestrator creates independent generated runtime files. It does not depend on a fixed host port, database identity, path, credential, or image-local runtime fallback.

`tests/api/test_bootstrap_restart_docker.sh` defines normal-mode clean-storage coverage:

1. generate the complete installation configuration;
2. build and start through `scripts/deploy.sh`;
3. log in with the generated initial Admin;
4. verify `/ui/` is absent;
5. rerun without rotating `.env`;
6. restart after removing bootstrap values; and
7. verify the original Admin remains unique and usable.

`tests/api/test_auth_lockout_docker.sh` defines the rolling login window and fail-closed authentication persistence flows. The broader API suite defines RBAC, object access, workflow, finalization, redaction, Bates numbering, comparison, audit, notifications, pagination, error envelopes, backup, restore, and strict dependency flows.

The definitions prove intended coverage only. A PASS requires a retained executed artifact for the exact revision.

## Browser acceptance

The only browser asset is `public/index.html`, served at `/ui/` only in acceptance mode.

- `tests/api/test_ui_screenshot_acceptance.sh` defines rendering evidence and a screenshot manifest.
- `tests/api/test_ui_interaction_acceptance.sh` defines validation, keyboard focus, incorrect credentials, successful login, network failure guidance, recovery, retry, and evidence capture through Chrome DevTools Protocol.

A static screenshot or script definition is not interaction evidence.

## Complete regression

```bash
bash ci/run_full_regression.sh artifacts/regression
```

The runner is designed to be sequential and fail-fast. On the first failed stage it records that stage, writes a failed summary, and exits before later validation starts. A successful run is expected to write:

```text
artifacts/regression/results.tsv
artifacts/regression/summary.json
artifacts/regression/summary.md
artifacts/regression/source-inventory.json
artifacts/regression/logs/
artifacts/regression/ui-interaction/
```

A complete PASS requires pre-existing `summary.json` with `overall_status=passed`, every recorded stage status equal to zero, and a clean source inventory tied to the exact inspected revision.

## Evidence boundary

A reviewer-written report is a static summary, not test evidence. A historical run proves only its tested SHA. Before declaring runtime acceptance, record:

- tested commit SHA;
- workflow run and job IDs;
- generated summary status;
- artifact name, size, and digest;
- any difference between the evidence revision and the inspected revision; and
- checks that were not executed and why.
