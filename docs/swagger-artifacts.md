# Swagger Generated Artifacts Policy

## Source of truth

Route-level Swaggo annotations under `internal/app/swagger_*.go` are the API contract source of truth. Generated files under `docs/swagger/` are local or CI build artifacts created by `scripts/generate_swagger.sh`.

## Generation before compilation

The application imports the generated package:

```go
_ "ironpage-vault/docs/swagger"
```

Every supported entrypoint that compiles the application must therefore create the package and generate the Swagger artifacts first.

## Entrypoint behavior

| Entrypoint | Required behavior |
|---|---|
| `run_tests.sh` | Prepare `docs/swagger/docs.go`, generate Swagger, then run selected checks |
| Docker builder | Generate Swagger in the builder stage before compiling the server |
| `ci/run_full_regression.sh` | Generate Swagger before route coverage, vet, race tests, and Docker validation |
| `ci/run_tests_contract_check.sh` | Start from a clean generated-artifact state and verify the local entrypoint recreates the files |

## Contract guards

- `ci/swagger_contract_check.sh` compares generated routes with route-level `@Router` annotations.
- `tests/contracts/swagger_route_coverage.sh` checks route coverage in the generated contract.
- `.github/workflows/ci.yml` is the sole workflow and reaches these checks through the sequential complete-regression entrypoint.

## Evidence boundary

Generated Swagger files from an earlier revision are not proof for the current source. An API-contract claim must identify the revision and the generated contract or successful complete-regression artifact that belongs to it.
