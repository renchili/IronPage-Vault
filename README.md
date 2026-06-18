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

The backend is being split into explicit packages so API handlers do not own domain policy or infrastructure helpers:

```text
cmd/server/          backend application entrypoint
internal/app/        Echo routes, middleware, request binding, API response mapping
internal/core/       domain constants, role rules, workflow rules, access policy, validation rules
internal/service/    use-case orchestration and transaction-level application flows
internal/store/      database repositories and SQL-facing persistence code
internal/platform/   filesystem, PDF, digest, crypto, and backup adapters
migrations/          PostgreSQL schema
public/              backend test UI only, not formal frontend scope
testdata/            local PDF and CSV fixtures
unit_tests/          unit and structure validation scripts
API_tests/           API acceptance scripts
docs/                API, design, RBAC, security, usage, testing, and requirement docs
Dockerfile           Docker builder plus runtime image definition
docker-compose.yml   one-command local startup
```

Current migration status:

- domain validation rules have moved from `internal/app` to `internal/core`
- workflow chain rules have moved from `internal/app` to `internal/core`
- notification unread-cap policy has moved from `internal/app` to `internal/core`
- text token and mention parsing policy has moved from `internal/app` to `internal/core`
- object-level access policy has moved from `internal/app` to `internal/core`
- document list SQL filter construction has moved from `internal/app` to `internal/store`
- crypto, digest, and PDF helper implementations have moved from `internal/app` to `internal/platform`
- `internal/app` keeps temporary compatibility wrappers while handlers are migrated in small PRs
- remaining SQL-heavy code and backup adapters still need follow-up migrations

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

## Testing Entry Point

`run_tests.sh` is the unified local verification entry point. It runs:

```text
unit_tests/test_rules.sh
go test -mod=mod ./...
API_tests/test_api_flow.sh
API_tests/test_static_review_reject_flows.sh
API_tests/test_acceptance_denials.sh
API_tests/test_compare_acceptance.sh
API_tests/test_finalized_immutability.sh
API_tests/test_redaction_coordinate_ciphertext.sh
API_tests/test_pdf_content_acceptance.sh
API_tests/test_notification_mention_side_effect.sh
unit_tests/test_structure_rules.sh
API_tests/test_strict_dependency_failures.sh
API_tests/test_bates_sequence_multi_doc.sh
```

CI additionally runs `gofmt`, `go vet`, Swaggo generation, Docker build, and Docker acceptance.

## Generated API Documentation

OpenAPI documentation is generated from Swaggo annotations in Go source code. Do not manually edit Swagger YAML as the source of truth.

```bash
bash scripts/generate_swagger.sh
```

Generated files live under:

```text
docs/swagger/
```

The Echo server mounts the generated Swagger UI at:

```text
/swagger/index.html
```

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
```

## AWS Deployment Options

AWS deployment guidance is available for both serverless-container and EKS targets:

```text
docs/aws-deployment.md
deploy/aws/serverless/
deploy/aws/eks/
```

Without an AWS account, use Docker/SAM validation for the serverless artifacts and kind or minikube for the EKS manifests.
