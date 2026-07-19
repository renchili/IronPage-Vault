# CI Boundary

IronPage Vault uses one GitHub Actions workflow: `.github/workflows/ci.yml`.

## Execution model

The same workflow handles pull requests, merge groups, pushes to `main`, and reviewed manual replays. One repository-and-target concurrency group serializes these events. `ci/ci_execution_guard.py` enforces a ten-minute same-revision cooldown and uses GitHub Actions history as an auditable failed-revision latch.

A failed revision may proceed only after either:

- a new reviewed commit changes the SHA; or
- a deliberate `workflow_dispatch` replay sets the explicit unlock input.

## Fail-fast rule

The workflow has one job and runs validation in a fixed sequence. `ci/run_full_regression.sh` records the first failed stage, writes its local failure summary, and exits immediately. GitHub therefore starts no later build, test, artifact upload, or summary-publication step after a failure.

Artifacts are retained only after a complete successful regression.

## CI-owned control plane

CI orchestration and contracts live under `ci/`:

- `ci/ci_execution_guard.py`
- `ci/regression_contract_check.sh`
- `ci/run_full_regression.sh`
- `ci/docker_acceptance.sh`
- `ci/run_project_api_regression.sh`
- `ci/run_tests_contract_check.sh`
- `ci/docs_consistency_check.sh`
- `ci/shell_syntax_check.sh`
- `ci/Dockerfile.acceptance`

## Product and test boundaries

Product code and runtime assets are under `cmd/`, `internal/`, `migrations/`, `public/`, `scripts/`, `Dockerfile`, and `docker-compose.yml`.

Stateful HTTP and browser acceptance flows are under `tests/api/`. Repository and generated-contract checks are under `tests/contracts/`. Go unit tests remain colocated with their packages, following Go conventions. Fixtures remain under `testdata/`.

`run_tests.sh` is a local/manual entrypoint. Its report describes only stage rows actually executed. Complete retained evidence comes from the serialized full-regression workflow, not from the lightweight entrypoint contract probe.
