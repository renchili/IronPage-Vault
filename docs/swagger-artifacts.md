# Swagger Generated Artifacts Policy

## Source of truth

Route-level Swaggo annotations under `internal/app/swagger_*.go` are the API contract source of truth.

Generated files under `docs/swagger/` are build/test artifacts. They may be created locally or in CI by `scripts/generate_swagger.sh`.

## Why generation is required before tests

The application imports the generated package:

```go
_ "ironpage-vault/docs/swagger"
```

A fresh checkout can therefore fail to compile if `docs/swagger/docs.go` does not exist. Every entrypoint that compiles the application must prepare generated Swagger files first.

## Required behavior by entrypoint

| Entrypoint | Required Swagger behavior |
|---|---|
| `run_tests.sh` | Create `docs/swagger/docs.go` stub and run `scripts/generate_swagger.sh` before `go test` |
| Docker build | Create `docs/swagger/docs.go` stub and generate Swagger before building the server |
| PR CI app tests | Run `scripts/generate_swagger.sh` before app package tests |
| Full regression | Prepare generated Swagger before full gofmt/vet/race tests |
| Local contract probe | Remove `docs/swagger`, run `run_tests.sh` probe mode, and verify regenerated files exist |

## CI guards

- `ci/run_tests_contract_check.sh` verifies local entrypoint behavior from a clean `docs/swagger` state.
- `ci/swagger_contract_check.sh` validates generated `docs/swagger/swagger.yaml` against every route-level `@Router` annotation.
- `Pull Request CI` runs the local entrypoint contract when `run_tests.sh`, Swagger generation, or Swagger artifacts change.
- `Pull Request CI` runs the generated Swagger contract when route annotations, Swagger generation, or the contract checker changes.

## Do not rely on stale generated files

Do not treat an existing `docs/swagger/swagger.yaml` from a prior run as sufficient evidence. CI and local entrypoints must regenerate Swagger as part of the relevant workflow.
