# Swagger Generated Artifacts Policy

## Source of truth

Route-level Swaggo annotations under `internal/app/swagger_*.go` are the API contract source of truth. Generated files under `docs/swagger/` are local or lifecycle build artifacts created by `scripts/generate_swagger.sh`.

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
| `ci/run_full_regression.sh` | Generate Swagger before generated-contract, vet, race-test, and Docker stages |
| `ci/run_tests_contract_check.sh` | Start from a clean generated-artifact state and verify the local entrypoint recreates the files |
| `.github/workflows/ci.yml` | Perform static route-annotation coverage only; do not generate Swagger or compile the application |

## Contract guards

- `ci/swagger_contract_check.sh` compares generated routes with route-level `@Router` annotations when a supported execution entrypoint generated the files.
- `tests/contracts/swagger_route_coverage.sh` checks required route annotations directly and is safe for the repository static workflow.
- `.github/workflows/ci.yml` is the sole GitHub workflow and treats route coverage as static evidence only.

## Static reviewer boundary

A static reviewer must not run `scripts/generate_swagger.sh`, compilation, tests, or CI to fill a missing artifact. The reviewer may inspect existing generated files and existing completed artifacts read-only.

Missing current generated Swagger evidence is `NOT VERIFIED` for generated-contract claims. It does not authorize reviewer execution.

## Evidence boundary

Generated Swagger files from an earlier revision are not proof for current source. An API-contract claim must identify the revision and either:

- a pre-existing generated contract tied to that revision; or
- source-tree equivalence plus an explicit caveat.

Static route-annotation coverage proves only that required annotations are present. It does not prove runtime routing, response behavior, authentication behavior, or successful generation.
