# Swagger / OpenAPI

IronPage Vault supports Swaggo-based API documentation generation.

## Dependencies

The Go module includes:

```text
github.com/swaggo/swag
github.com/swaggo/echo-swagger
```

## Generate Swagger files

From the repository root:

```bash
bash scripts/generate_swagger.sh
```

The generation script uses `cmd/server/main.go` as the Swaggo entrypoint and writes output under `docs/swagger/`.

## Expected output

Swaggo generates:

```text
docs/swagger/docs.go
docs/swagger/swagger.json
docs/swagger/swagger.yaml
```

## Source of truth

Route-level Swaggo annotations in the Go source are the generated contract source. The hand-written API usage reference is:

```text
docs/api-spec.md
```

Generated Swagger output and the hand-written reference must agree with the current routes, request fields, response fields, authentication behavior, and error envelopes.

## Served routes

Swagger UI is mounted at:

```text
/swagger/index.html
```

The acceptance-only backend probe UI is separate and is served only when acceptance mode is explicitly enabled:

```text
/ui/
```

The acceptance UI is not a product frontend. Rendering or screenshot evidence alone does not prove its login, error-recovery, keyboard, focus, or retry interactions.