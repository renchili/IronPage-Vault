# CI Boundary

This repository separates the CI control plane from product code and project-owned tests.

## Pre-merge pull request checks

`pull_request` workflows are change-driven gates. They analyze the changed surface and execute the checks for that surface.

Allowed pre-merge checks include:

- changed-file impact analysis
- gofmt
- go vet
- targeted `go test` for affected packages
- generated Swagger/static contract checks
- shell syntax checks with `bash -n`
- workflow lint
- Docker image build
- CI-flow contract probes when CI workflow or CI runner logic changes

Pre-merge pull request checks must not use `run_tests.sh` as the pass/fail source.

## Flow-specific rule

Changing CI flow code must execute a CI-owned contract for that flow.

Examples:

- Changes to `ci/run_full_regression.sh` or full-regression workflows run `ci/regression_contract_check.sh`.
- Changes to product runtime code run targeted package tests and Docker build.
- Changes to API surface run Swagger/static contract checks.
- Changes to project-owned shell tests run syntax checks and impact analysis, but not `run_tests.sh` as the PR conclusion.

## Merge candidate regression

Full runtime/API regression belongs on a real merge candidate, not on an arbitrary feature branch checkout. Use the `merge_group` workflow for merge-queue regression so the tested tree is the temporary merge result produced from current `main` plus queued changes.

## Post-merge evidence

`push` to `main` may replay full regression and retain logs, JSON summaries, Markdown summaries, and artifacts under `reports/regression/**` when product/runtime/regression-impacting paths changed.

## CI control plane

CI-owned control scripts live under `ci/`.

Examples:

- `ci/change_impact_check.sh`
- `ci/regression_contract_check.sh`
- `ci/run_full_regression.sh`
- `ci/docker_acceptance.sh`
- `ci/run_project_api_regression.sh`
- `ci/shell_syntax_check.sh`
- `ci/Dockerfile.acceptance`

## Product code

Product code is the object being tested.

Examples:

- `cmd/`
- `internal/`
- `migrations/`
- `public/`
- `Dockerfile`
- `docker-compose.yml`
- product build/runtime helpers under `scripts/`

## Project-owned tests and local tools

These are useful for local development, manual replay, merge-candidate regression, and post-merge evidence. They are not the neutral pull-request static gate.

Examples:

- `run_tests.sh`
- `API_tests/`
- `unit_tests/`
- `testdata/`

## Hard rules

1. Pull-request workflows call CI-owned checks under `ci/**` and standard tools.
2. Pull-request workflows do not use `run_tests.sh` as a pass/fail source.
3. Pull-request workflows must analyze changed files and require tests/contracts for new capability.
4. If CI flow code changes, pull-request workflows must execute a CI-owned contract probe for that flow.
5. Runtime/API regression runs on `merge_group` merge candidates or post-merge `main`.
6. `ci/docker_acceptance.sh` owns Docker runtime regression orchestration and does not delegate to `run_tests.sh`.
