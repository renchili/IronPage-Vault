# CI Boundary

This repository separates CI control-plane code from project-owned code and project-owned tests.

## CI control plane

CI workflows may call only CI-owned control scripts under `ci/` for pre-merge conclusions.

Examples:

- `.github/workflows/**`
- `ci/run_full_regression.sh`
- `ci/docker_acceptance.sh`
- `ci/api_smoke.sh`
- `ci/shell_syntax_check.sh`
- `ci/Dockerfile.acceptance`

CI control scripts may run standard tools such as `go test`, `go vet`, `gofmt`, `docker build`, `docker compose`, `bash -n`, and `actionlint`.

## Project-owned code

Project code is the object being tested. It must not decide pre-merge pass/fail by invoking its own aggregate test entrypoint.

Examples:

- `cmd/`
- `internal/`
- `migrations/`
- `public/`
- `Dockerfile`
- `docker-compose.yml`
- product build/runtime helpers under `scripts/`

## Project-owned tests and local tools

These are useful for local development, manual replay, and post-merge evidence, but they are not neutral pre-merge CI gates.

Examples:

- `run_tests.sh`
- `API_tests/`
- `unit_tests/`
- `testdata/`

Pre-merge CI may parse these scripts with `bash -n` or lint them, but it must not use `run_tests.sh` as the required CI conclusion.

## Hard rules

1. Pre-merge workflows call `ci/**`, not `run_tests.sh`.
2. `ci/**` must not call `run_tests.sh`.
3. `ci/docker_acceptance.sh` owns Docker runtime acceptance and must not delegate to project-owned aggregate test scripts.
4. Project-owned tests may be executed post-merge or manually as regression evidence.
5. If CI control-plane files change, workflow lint and shell syntax checks must run.
