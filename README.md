# IronPage Vault

IronPage Vault is an offline legal PDF lifecycle management **backend API** for air-gapped legal, compliance, and regulated document environments. It provides local identity, strict role boundaries, PDF intake, versioned document records, redaction workflows, reviewer annotations, Bates numbering, workflow transitions, audit trails, notifications, configuration, and local backup metadata.

This is a **pure backend project**. The browser UI under `public/` is only a testing aid for manual backend acceptance. It is not a formal frontend deliverable, not part of the product scope, and should not be evaluated as a fullstack application.

## Project Goals

IronPage Vault is built to satisfy these backend product goals:

- provide a REST API for legal PDF lifecycle management
- keep sensitive legal PDFs inside a standalone local environment
- avoid external identity providers, cloud services, remote PDF processors, or external notification providers
- enforce Admin, Editor, and Reviewer role boundaries
- preserve a complete document lifecycle from Draft to Finalized
- make Finalized documents immutable
- keep PDF binaries on the local filesystem while storing metadata in PostgreSQL
- record important mutating actions in audit logs
- provide local notification records for workflow and review activity
- support acceptance with real local PDF and CSV test fixtures

## Implementation Summary

The backend is implemented in Go with Echo and sqlx. PostgreSQL is the only persistence layer for metadata, sessions, audit records, workflow records, configuration, notifications, and backup job metadata. PDF files are stored in a local filesystem volume and referenced by database records.

The project includes:

```text
cmd/server/          backend application entrypoint
internal/app/        Echo routes, handlers, database access, auth, document logic
migrations/          PostgreSQL schema
public/              backend test UI only, not formal frontend scope
testdata/            local PDF and CSV fixtures
unit_tests/          unit and structure validation scripts
API_tests/           API acceptance scripts
docs/                API, design, RBAC, security, usage, testing, and requirement docs
Dockerfile           Docker builder plus runtime image definition
docker-compose.yml   one-command local startup
```

## Core Backend Modules

| Module | Responsibility |
|---|---|
| Authentication | local username/password login, bcrypt hashes, JWT issuance, sessions |
| RBAC | Admin, Editor, Reviewer capability boundaries |
| Documents | PDF upload, metadata, local storage, current version tracking |
| Versions | document version metadata, rollback, and revision ceiling support |
| Workflow | Draft, Under Review, Redaction Pending, Approved, Finalized |
| Redaction | staged redaction regions and Editor confirmation flow |
| Annotations | Reviewer notes, highlights, strikethroughs, text stamps, dispositions |
| Bates | Bates job metadata with prefix, suffix, padding, and start number |
| Audit | structured records for mutating operations |
| Notifications | local in-app notification records |
| Configuration | Admin-managed system entries and workflow definitions |
| Backup | local backup job metadata and recovery documentation |
| Testing | Docker-based acceptance, API scripts, unit checks, local PDF/CSV fixtures |

## Roles

IronPage Vault supports only three roles:

| Role | Responsibility |
|---|---|
| Admin | user management, configuration, workflow definitions, templates, backup metadata |
| Editor | PDF upload, version actions, redaction confirmation, Bates numbering, finalization |
| Reviewer | document retrieval, annotations, annotation dispositions, review workflow movement |

Admin is intentionally not treated as a document editor. This keeps system administration separate from legal document manipulation.

## Document Lifecycle

Documents follow this required chain:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Finalized documents are treated as closed legal records. Mutation APIs should reject changes after Finalized status.

## Storage Model

IronPage Vault separates binary and relational data:

- PostgreSQL stores users, sessions, documents, versions, audit logs, notifications, workflow history, redaction metadata, annotations, Bates jobs, config, and backup records.
- The local filesystem stores PDF binaries.
- Database version rows point to PDF files by path and include file hash, size, page count, and version number.

This design keeps large binary assets out of ordinary relational queries while preserving traceability.

## Offline Deployment Model

The project is packaged for local standalone execution. The Compose setup uses one application container that includes PostgreSQL and the Go API process. Persistent data is stored through Docker volumes.

The intended build path is **Docker builder**, not a local Go toolchain. The project does not require a local Go installation or a committed `go.sum` for acceptance.

See:

```text
docs/docker-builder.md
docs/usage.md
scripts/docker_acceptance.sh
```

## Backend Test UI

A lightweight browser-based test UI is included only for manual backend verification. It is not a production frontend and not a fullstack requirement.

Source location:

```text
public/manual-test.html
```

Runtime route:

```text
http://localhost:8080/ui/manual-test.html
```

The test UI exists only to manually exercise backend flows such as login, upload, workflow, annotations, redaction, Bates, audit, notifications, backup metadata, and Swagger YAML retrieval during acceptance.

## Swagger / OpenAPI

The project includes Swaggo dependencies. The intended generation flow is documented in `docs/api-spec.md` and `docs/swagger/README.md`:

```bash
swag init -g cmd/server/main.go -o docs/swagger
```

Generated Swagger output should mirror the Markdown API specification. A static YAML copy is also exposed for manual backend testing at:

```text
/swagger/swagger.yaml
```

## Test Data

The project includes local acceptance fixtures:

```text
testdata/pdfs/sample_contract.pdf
testdata/pdfs/sample_regulatory_filing.pdf
testdata/csv/batch_import_manifest.csv
```

These files allow offline backend testing without downloading external documents.

## Testing Position

Acceptance should use Docker, not the local Go environment.

Main test entrypoints:

```text
run_tests.sh
scripts/docker_acceptance.sh
unit_tests/test_rules.sh
API_tests/test_api_flow.sh
```

The API test flow is expected to log in with seeded Admin, Editor, and Reviewer users, export the three tokens, and then run role-specific API suites.

No tests were executed during generation.

## Documentation Map

| Document | Purpose |
|---|---|
| `AGENT.md` | single source of implementation and acceptance rules |
| `CLAUDE.md` | pointer to `AGENT.md` to avoid duplicated rules |
| `PLAN.md` | implementation plan and module breakdown |
| `metadata.json` | project metadata and full prompt |
| `docs/api-spec.md` | backend API interface reference and Swaggo notes |
| `docs/design.md` | backend design rationale and architecture decisions |
| `docs/requirement-check.md` | prompt-to-implementation completion review |
| `docs/questions.md` | project Q&A and decision reasoning |
| `docs/rbac.md` | role and capability matrix |
| `docs/security.md` | local security model and acceptance checks |
| `docs/usage.md` | startup, manual backend testing, and operational commands |
| `docs/testing.md` | backend testing strategy and acceptance flow |
| `docs/docker-builder.md` | Docker builder workflow and no-local-Go build policy |
| `docs/backup-recovery.md` | local database and PDF storage recovery guidance |
| `docs/pitr.md` | point-in-time recovery model |
| `docs/deployment-offline.md` | standalone offline deployment guidance |
| `docs/swagger/` | Swaggo/OpenAPI generation notes and initial spec |
| `docs/test-effectiveness-followup.md` | notes on API test effectiveness improvements |

## Acceptance Position

This repository is a backend API prototype plus backend acceptance documentation, local fixtures, and a small manual testing UI. `docs/requirement-check.md` records which requirements are complete, partial, or planned so reviewers can evaluate implementation status honestly.
