# Swagger / OpenAPI

IronPage Vault supports Swaggo-based API documentation generation.

## Dependencies

The Go module includes:

```text
github.com/swaggo/swag
github.com/swaggo/echo-swagger
```

## Generate Swagger Files

From the repository root:

```bash
swag init -g cmd/server/main.go -o docs/swagger
```

## Expected Output

Swaggo normally generates:

```text
docs/swagger/docs.go
docs/swagger/swagger.json
docs/swagger/swagger.yaml
```

## Source of Truth

The hand-written API reference is:

```text
docs/api-spec.md
```

Generated Swagger output should mirror that API document.

## UI Route

When Swagger UI is wired into the Echo server, the intended route is:

```text
/swagger/*
```

The manual acceptance UI is separate and is served from:

```text
/ui/manual-test.html
```
