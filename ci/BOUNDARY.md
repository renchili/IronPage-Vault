# CI Boundary

This repository separates the CI control plane from product code and project-owned tests.

## Pre-merge pull request checks

`pull_request` workflows are static and build-oriented gates. They analyze the changed surface and verify that new or changed capability has matching test or contract updates.

Allowed pre-merge checks include:

- changed-file impact analysis
- gofmt
- go vet
- targeted `go test` for affected packages
- generated Swagger/static contract checks
- shell syntax checks with `bash -n`
- workflow lint
- Docker image build

Pre-merge pull request checks must not use `run_tests.sh` as the pass/fail source.

## Merge candidate regression

Full runtime/API regression belongs on a real merge candidate, not on an arbitrary feature branch checkout. Use the `merge_group` workflow for merge-queue regression so the tested tree is the temporary merge result produced from current `main` plus queued changes.

## Post-merge evidence

`push` to `main` may replay full regression and retain logs, JSON summaries, Markdown summaries, and artifacts under `reports/regression/**`.

## CI control plane

CI-owned control scripts live under `ci/`.

Examples:

- `ci/change_impact_check.sh`
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
4. Runtime/API regression runs on `merge_group` merge candidates or post-merge `main`.
5. `ci/docker_acceptance.sh` owns Docker runtime regression orchestration and does not delegate to `run_tests.sh`.
