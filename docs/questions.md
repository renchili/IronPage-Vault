# Questions and Implementation Answers

This document is a current project clarification-answer record. It explains how ambiguous project behavior should be understood from the current repository rules, implementation, and available validation evidence.

It is not a work log, not an agent-process diary, not a future cleanup checklist, and not a substitute for tests or CI evidence.

## Source inputs

- `AGENT.md`: project-specific product and documentation rules.
- `README.md`: current backend scope and package map.
- `docs/requirement-check.md`: current requirement status summary.
- `docs/api-spec.md`: API surface documentation.
- Current source paths under `cmd/`, `internal/`, `migrations/`, `API_tests/`, and `unit_tests/`.
- Full project re-audit report uploaded on 2026-07-13.

## Evidence boundary

The strongest historical runtime evidence is the full regression run recorded in the re-audit report: GitHub Actions run `28109623265` at SHA `3522a9dfed8e38556c0cc4d4e147be14fe405a95`.

That evidence supports the behavior captured by that run, but it is not the same as a fresh full-regression run at the current `main` HEAD. Documentation must state exact SHAs and run IDs when claiming validation evidence.

## Current project scope

IronPage Vault is a backend-first legal PDF lifecycle API. The deliverable is a local Go/Echo service with PostgreSQL metadata, filesystem PDF storage, workflow, versioning, annotations, redaction metadata, Bates processing records, audit logs, notifications, and backup records.

The browser UI under `public/` is a backend testing aid for manual API probing and screenshot evidence. It is not a production frontend and must not expand the acceptance scope into a fullstack product.

## Docker delivery model

The intended local acceptance path is Docker-based. The repository provides a Docker/Compose path so acceptance does not depend on a developer's local Go installation.

The re-audit report records successful Docker build and stateful API acceptance evidence for a historical behavior-equivalent SHA. Current-HEAD acceptance still requires an exact run tied to the current default-branch SHA before documentation may claim fresh current-HEAD full regression.

## Role access and object access

Role access decides whether a role may call a category of endpoint. Object access decides whether that user may access a specific document.

Admin, Editor, and Reviewer are the only supported roles. Admin is not automatically a document editor. Editors own document mutation flows such as upload, redaction confirmation, Bates numbering, finalization, and version actions. Reviewers retrieve non-Draft records, annotate, set dispositions, and move workflow where allowed.

Documentation must not collapse role access and object access into one generic RBAC claim.

## Document lifecycle

The required document lifecycle is:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Finalized is the terminal immutable state. Mutating operations after Finalized must be rejected, including replacement upload, rollback, redaction, annotation changes, Bates numbering, workflow transition, and metadata mutation.

## Batch upload

Batch upload is part of the document intake model. It must use the same persistence path as single upload: document record, version record, stored PDF file, and audit output per accepted file.

Documentation must not describe batch upload as a placeholder if current source and acceptance evidence prove persisted document/version/file side effects.

## Protected data and deployment status

The repository contains encrypted metadata paths for sensitive annotation and redaction fields. This implementation detail is not the same as secure deployment acceptance.

The re-audit report still blocks acceptance on unsafe runtime defaults. Documentation must not describe the project as production-ready or security-accepted until those defaults are removed or gated behind an explicit acceptance-only mode.

## Redaction status

Redaction is a two-phase workflow: proposed coordinate-bound metadata is staged, then an authorized Editor confirms burn-in.

Current documentation must not claim marker-only redaction when the current strict PDF path and recorded Docker acceptance evidence show strict burn-in behavior and target-text removal. If a document still states marker-only behavior, that document is stale and must be rewritten or removed.

## Bates status

Bates processing is not metadata-only in the current implementation record. Current documentation must describe visible Bates processing and sequence allocation only when backed by strict adapter code and runtime evidence.

Documentation must not keep obsolete statements that Bates numbers are not visible if current strict Bates evidence says otherwise.

## Document comparison status

Comparison must be described from current service behavior and evidence. Documentation must not call comparison binary-only if current code and tests expose structured text and coordinate-aware comparison behavior.

When static source assembly prevents a complete route or field inventory, documentation must say what is directly verified and point to the source or API documentation home.

## Audit and notifications

Audit logs are required for material mutations and must support filtering by user, document, action type, and date range.

Notifications are local in-app records. Workflow updates and annotation mentions can create notifications. Documentation must distinguish notification persistence and read acknowledgement from external delivery, which is out of scope.

## Backup and restore status

Backup is not metadata-only in the current implementation record. The Admin backup path calls strict artifact creation and requires a PostgreSQL custom dump and tar snapshot before it reports restore support.

Documentation must not say that a future worker is still needed for full local backup if the current implementation already runs strict local artifacts. It may still document operational prerequisites, local backup volume access, and consistent database/filesystem recovery boundaries.

## PITR status

The project requires point-in-time recovery documentation, but full automated WAL archiving and restore orchestration are not proven as current implementation.

PITR documentation must state the supported scope clearly: local recovery strategy and required consistency model are documented; automated physical PITR orchestration must not be claimed unless implemented and validated.

## API and acceptance evidence

The repository contains broad API acceptance coverage, including authentication, RBAC denials, workflow/finalization, redaction, Bates, comparison, audit, notifications, backup, restore, and UI screenshot evidence.

Documentation must still distinguish historical successful runs from exact current-HEAD runs. A passing historical artifact is evidence, but it is not fresh verification for a later commit.

## Manual backend UI scope

The manual UI is served at `/ui/` and is only a backend testing aid.

Current screenshot acceptance proves page load and screenshot generation. It does not prove login clicks, retry behavior, accessibility, keyboard focus, or full operator recovery flow. Documentation must not describe screenshot evidence as full UI E2E coverage.

## Documentation status

The current re-audit verdict is `FAIL` until the blocking P0 issues are fixed:

- unsafe runtime defaults.
- official documentation contradictions.

This document removes stale contradictions from the clarification record, but it does not fix runtime configuration safety and does not replace a current-HEAD full regression.
