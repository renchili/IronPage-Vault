# Backend Architecture Boundaries

IronPage Vault backend code must be split by responsibility. The API package is not a dumping ground for domain rules, utility helpers, storage logic, or infrastructure behavior.

## Target layers

```text
cmd/server             process entrypoint
internal/app           HTTP routing, Echo handlers, auth middleware, request/response mapping
internal/core          domain constants, workflow rules, permission rules, validation rules
internal/service       use-case orchestration, transactions across domain operations
internal/store         database repositories, SQL, persistence adapters
internal/platform      filesystem, PDF inspection/transform adapters, crypto, digest, backup adapters
```

## Package rules

### internal/app

Allowed:

- Echo routes and middleware
- request binding and response formatting
- translating service/domain errors into API errors

Not allowed:

- domain rules such as role permissions or workflow transitions
- SQL query ownership beyond temporary migration code
- PDF/file transformation logic
- crypto/digest implementation
- backup implementation details

### internal/core

Allowed:

- role constants and permission rules
- workflow status constants and transition rules
- validation rules for annotations, redactions, Bates numbering, batch size, lockout, request freshness

Not allowed:

- Echo context
- SQL or sqlx
- filesystem access
- environment variables
- encryption implementation

### internal/service

Allowed:

- use-case orchestration, such as upload, transition, finalize, compare, redact, annotate, backup
- transaction boundaries when multiple repositories must be updated together

Not allowed:

- HTTP request/response structs coupled to Echo
- raw SQL embedded directly in handlers

### internal/store

Allowed:

- SQL queries
- repositories
- transaction helper methods

Not allowed:

- Echo handlers
- API response bodies
- domain policy decisions that should live in core

### internal/platform

Allowed:

- PDF inspection and transformation adapters
- crypto helpers
- digest helpers
- filesystem backup adapters

Not allowed:

- HTTP handlers
- business permission checks
- request principal logic

## Refactor sequence

1. Create empty package boundaries.
2. Move pure domain rules from `internal/app/rules.go` into `internal/core`.
3. Move object access rules from `internal/app/access.go` into `internal/core` using core-owned domain structs or small policy inputs.
4. Move PDF, digest, and crypto helpers into `internal/platform`.
5. Move SQL-heavy handler logic into `internal/store` repositories and `internal/service` use cases.
6. Keep `internal/app` as a thin API adapter.

Each step should be a small PR with no behavior change unless the PR title explicitly says it changes behavior.
