---
name: full-project-acceptance-hard-gates
description: Static hard-gated acceptance of complete software projects through source, test-definition, schema, configuration, documentation, workflow, deployment, and repository inspection without project execution or CI dependency.
---

# Full Project Static Acceptance Hard Gates

Use this Skill to accept or reject a repository, branch, commit, PR, generated project, source package, or ZIP by static inspection.

## Absolute static-only boundary

Acceptance is source inspection, not execution. The reviewer must not:

- run project code, tests, scripts, binaries, generators, linters, formatters, compilers, or package managers;
- build packages, applications, images, or containers;
- start services, databases, browsers, emulators, or deployments;
- trigger, retry, rerun, dispatch, approve, cancel, or wait for CI;
- require CI, build, deployment, test, screenshot, log, or external-validator results before continuing;
- ask the user to run commands as a condition for completing the review;
- stop after the first P0, failed gate, missing artifact, or absent external result.

Existing external results may be read as optional context. They are never prerequisites and never replace current-source inspection. Missing execution evidence alone must not cause `FAIL`, `CONDITIONAL`, or `NOT VERIFIED`. A green external result cannot override a static defect.

## No minimization and no early stop

Inspect the entire applicable project and enumerate every statically identifiable defect. Do not review only changed files when cross-cutting behavior is involved, stop after one blocker, reduce findings to the minimum needed for failure, defer known work to CI, or trust names and summaries without tracing implementation.

Every applicable Gate 0–27 must be completed even when earlier gates fail.

## Core rule

A project passes only when static inspection establishes that:

1. original requirements and user corrections form an atomic matrix;
2. every material requirement maps to current implementation and static evidence paths;
3. source, tests, schemas, migrations, configuration, manifests, workflows, deployment, docs, and comments agree;
4. positive, negative, boundary, failure, authorization, security, workflow, and recovery behavior exists in implementation and test definitions;
5. repository layout, naming, formats, packaging, and contamination are acceptable;
6. no required behavior is a stub, route-only shell, documentation-only promise, unsafe fallback, or undeclared external dependency;
7. all report tables are complete and consistent;
8. no blocking static Gate fails.

## Evidence order

1. Original requirements and explicit user corrections.
2. `AGENTS.md`, `AGENT.md`, and relevant Skills.
3. Current production source, schemas, migrations, configuration, manifests, deployment, and API contracts.
4. Current positive, negative, boundary, and failure-path test definitions.
5. Current static guards and workflow definitions.
6. Current documentation and comments.
7. Pre-existing external artifacts, optional and read-only.
8. User claims.
9. Reviewer summaries.

Reviewer reports are summaries, not project evidence.

## Status vocabulary

```text
PASS          Complete and consistent by static inspection.
CONDITIONAL   A material static ambiguity, partial implementation, indirect mapping, or non-blocking limitation remains.
FAIL          Required behavior is missing, contradicted, unsafe, malformed, misleading, non-portable, or incomplete.
NOT VERIFIED  Required source, package content, or rule material was inaccessible or not inspected. Never use this for missing execution.
N/A           Not required by the specification; reason required.
```

Final `PASS` requires every applicable Gate to be `PASS` or justified `N/A`.

## Gap severity

```text
P0 blocker       Cannot accept.
P1 conditional   Requires a static correction and caveat.
P2 quality       Non-blocking quality defect.
Source gap       Required source or rule material is inaccessible.
Spec gap         Requirement remains ambiguous after source reconciliation.
Packaging gap    Path, mode, format, or archive defect.
Doc-code gap     Documentation/comments contradict implementation.
Interaction gap  Static interaction, error, recovery, or state design is incomplete.
Code-quality gap Naming, structure, coupling, or complexity harms verification.
```

## Mandatory inventory

Record repository/package path, exact revision/hash, file count and root layout, source, tests, docs, scripts and modes, workflows, deployment, migrations/configuration, artifacts, generated/cache/secret/runtime files, and path-name hazards. For ZIPs compare archive entries and mode metadata without executing content. No inventory means no `PASS`.

# Hard Gates

## Gate 0: Source and rule provenance

Confirm exact target revision/package, latest relevant changes, loaded rules with identifiers, and existence of every cited path. FAIL for stale targets, invented paths, skipped rules, or reused conclusions without re-reading current source.

## Gate 1: Atomic requirement coverage

Build requirements for architecture, deployment, persistence, storage, API, workflow, security, auth/session, protected data, domain features, audit, notifications, backup, interaction, docs, tests/static guards, CI definitions, hygiene, and maintainability.

Required table:

```text
ID | Requirement | Category | Implementation path | Test/static path | Status | Gap
```

FAIL if a major requirement is omitted.

## Gate 2: Architecture and deployment

Inspect language, framework, persistence, storage, dependencies, package files, startup logic, assets, entrypoints, manifests, and deployment definitions. Trace configuration through application, image, Compose/manifests, entrypoint, and deploy scripts. Reject undeclared services and hard-coded deployment identity, credentials, hosts, ports, or machine paths. Do not build or deploy.

## Gate 3: Data model and persistence

Inspect entities, keys, constraints, source-of-truth fields, history/version/audit structures, migrations, read/write alignment, transactions, restart semantics, and immutability. FAIL when schema and code disagree or required state/constraints are absent.

## Gate 4: Authentication, sessions, freshness, replay

Trace credential storage/comparison, rolling lockout, session/token creation, inactivity expiry, revocation/logout, blacklist, request freshness, replay protection, fail-closed dependency errors, and positive/negative/boundary tests. FAIL for happy-path-only coverage or ignored security-state errors.

## Gate 5: Protected data

For each sensitive field map schema, encryption/write path, decryption/read path, key source, protected format, masking, logging exclusions, and tests/guards. FAIL for plaintext source of truth, unsafe key fallback, frontend-only masking, or missing exposure boundaries.

## Gate 6: RBAC and object access

Trace every role through middleware, service/domain rules, object checks, repository queries, serialization, and tests.

```text
Role | Allowed | Forbidden | Object scope | Visible fields | Hidden fields | Positive test | Negative test
```

FAIL when forbidden, object-level, state-dependent, or field-visibility paths are absent.

## Gate 7: Workflow and terminal immutability

Inspect states, valid/invalid transitions, permissions, terminal immutability across every mutation, history, audit, notifications, and tests. FAIL if any mutation bypasses terminal rules or invalid transitions are accepted.

## Gate 8: Domain features

For every feature trace endpoint/command, validation, domain/service logic, storage mutation, side effects, errors, positive/negative tests, and output contract. FAIL for documentation-only, route-only, type-only, or stub implementations.

## Gate 9: Audit, notifications, configuration side effects

Inspect every mutation for audit metadata, notifications, acknowledgement, limits, templates/configuration, authorization, transactions, and error handling. FAIL for missing or silently ignored side effects.

## Gate 10: API contract, errors, pagination

Compare routes, handlers, request/response types, auth annotations, OpenAPI source, statuses, uniform errors, pagination, upload limits, and tests. Do not regenerate schemas. FAIL for disagreement or missing contracts.

## Gate 11: Backup, restore, operations

Trace backup schedule/configuration, command construction, database dump, filesystem snapshot, metadata, retention, restore mapping, paths, failure handling, authorization, and docs. Do not execute operations. FAIL for documentation-only or unsafe/incomplete implementation.

## Gate 12: Test and script entrypoints

Statically inspect documented commands, probe modes, services, stages, nested paths, outputs, summaries, exit propagation, working-directory assumptions, shebangs, and modes. PASS does not require execution. FAIL for missing paths, swallowed failures, or reports that can count skipped work as pass.

## Gate 13: Full-regression definition

Inspect the full-regression entrypoint, complete stages, dependency contract, failure propagation, summary model, artifacts, revision fields, and cleanup. Executed summaries are optional. FAIL when required areas are omitted, skipped stages can pass, or paths are missing.

## Gate 14: CI and workflow definitions

Inspect workflows only; never trigger or wait for CI. Check triggers, filters, permissions, shared concurrency, queue/cancellation, rate limiting, duplicate collapse, first-error propagation, failed-revision latches, unlock scope, dependency graph, `always()`/`continue-on-error`, paths, artifact claims, retention, secrets, and action dependencies.

CI repetition controls must be revision-aware:

- concurrency may serialize or collapse active work by target;
- cooldown must reject duplicate events for the same target and revision, not a different corrective revision;
- a failed revision remains latched against automatic repetition;
- a new revision must be admissible immediately so source fixes can be checked;
- an explicit unlock may authorize one exact failed revision once and must not become a general bypass.

FAIL for target-wide blocking of new revisions, parallel/conflicting work, duplicate queues, post-failure continuation, unbounded retry, false pass claims, missing paths, or acceptance dependency on CI.

## Gate 15: UI, CLI, manual surfaces

Inspect inclusion, mode gating, real route/command connection, scope description, primary action, loading, success, recovery, empty states, validation, destructive confirmation, terminology, focus/keyboard behavior, and specific errors. No screenshot or execution required. FAIL for disconnected mocks, fixture credential exposure, false production claims, or missing recovery.

## Gate 16: Documentation consistency

Compare setup, commands, paths, variables, dependencies, routes, DTOs, statuses, security, storage, workflow, backup, deployment, tests, evidence claims, and limitations with current source.

```text
Doc claim | Document path | Implementation path | Test/static path | Match? | Severity | Correction
```

FAIL for overstatement, nonexistent paths, stale process history, contradictions, or claims that future CI establishes current implementation.

## Gate 17: Repository/package layout

Inspect root files, source/test/docs/scripts/migrations/deployment paths, package root, duplicates, generated output, and ownership. FAIL for missing required files, duplicate roots, implementation hidden in artifacts, or ambiguous near-duplicates.

## Gate 18: File formats and hygiene

Inspect encoding, JSON/YAML/TOML/SQL/Markdown structure, shell/Python/Go syntax by reading, shebangs, line endings, extension/content agreement, placeholders, merge markers, binaries, and malformed generated source. Do not run parsers or compilers. FAIL for malformed or misleading files.

## Gate 19: Source/evidence paths

```text
Claim | Implementation path | Test path | Static artifact/doc path | Exists? | Current revision? | Notes
```

FAIL for stale, missing, unresolved, or invented paths.

## Gate 20: Naming and readability

Inspect conventions for files, packages, types, functions, variables, configuration, environment variables, routes, DTOs, tests, and scripts. Scan non-ASCII/locale-dependent paths, spaces/control characters, case-only collisions, singular/plural near-duplicates, mixed conventions, vague critical names, and oversized functions.

```text
Area | Source path | Finding | Convention | Status | Correction
```

## Gate 21: Comments and consistency

Inspect comments for intent, invariants, security, data contracts, state transitions, generated markers, and errors. Enumerate TODO/FIXME/HACK/XXX.

```text
Comment/topic | Source path | Related doc | Related implementation | Consistent? | Severity | Correction
```

FAIL for misleading, stale, contradictory, or opaque critical logic.

## Gate 22: Script portability

Inspect modes, shebangs, interpreters, nested invocation, relative paths, working-directory assumptions, quoting, dependencies, and documented commands. Do not run scripts. FAIL for machine paths, implicit modes, unsafe interpolation, missing interpreters, or missing nested files.

## Gate 23: Source contamination

Scan tracked/package paths for caches, `node_modules`, compiled output, databases, coverage, logs, temp files, secrets, real `.env`, editor/system files, stale reports, runtime storage, and binaries. FAIL when contamination can alter interpretation or conceal missing source.

## Gate 24: Report integrity

Inspect required sections/tables, Markdown/HTML structure, code fences, links, statuses, non-PASS corrections, verdict consistency, and evidence wording. FAIL if execution is falsely claimed, external results are treated as blockers, or the report is malformed/incomplete.

## Gate 25: Documentation pollution

Scan all docs for process residue including `next step`, `follow-up`, `future work`, `wait for CI`, `pending CI`, `run tests and report back`, `minimum fix`, recommendations, assistant notes, and correction chronology. Each occurrence must map to the specification, current implementation, a user-requested tracked issue, or explicit out-of-scope text.

Required tables:

```text
Doc path | Pattern | Excerpt | Justification | Classification | Action
Doc path | Claimed obligation | Requirement source | Implementation/test path | Valid? | Gap
Topic | README | Design | Questions/FAQ | Implementation | Consistent?
Roadmap item | Requested/tracked? | Implemented? | Acceptance relevance | Status
```

## Gate 26: Static interaction and recovery

For each critical role/flow inspect input, validation, permission, success, state mutation, subsequent visibility, audit/notification effects, invalid input, denial, expired/revoked session, replay, large/unsupported input, dependency failure, cancellation, terminal state, empty results, and retry/recovery.

```text
Flow | Role | Input contract | Expected interaction | State path | Test definition | Status | Gap
Prompt/copy path | Intended user | Required input | Expected output | Recovery guidance | Source | Status
Negative flow | Trigger | Expected recovery | Implementation path | Test path | Status
```

Do not execute interactions. FAIL for vague errors, missing recovery, incomplete state paths, or mocks presented as production behavior.

## Gate 27: Final verdict

Before the verdict produce scope/revision, loaded rules, inventory, requirement matrix, Gate table, source-path table, doc consistency, naming, comments, pollution, static interactions, test/workflow provenance, hygiene, gaps, and final decision.

```text
PASS          Every applicable static Gate is PASS or justified N/A.
CONDITIONAL   No P0, but at least one Gate is CONDITIONAL or NOT VERIFIED.
FAIL          Any P0 or required core behavior is missing/contradicted.
```

Missing test execution, CI, build, deployment, runtime logs, screenshots, or external artifacts does not alter the verdict.

# Required report

```markdown
# Full Project Static Acceptance Report
## Scope
## Executive verdict
## Loaded rule sources
## Repository/package inventory
## Requirement matrix
## Hard Gate results
## Source and evidence path validation
## Documentation-code consistency
## Code readability and naming
## Comment consistency
## Documentation pollution scan
## README/design/questions consistency
## Invented obligations and roadmap validation
## Static interaction flows
## Prompt/copy reliability
## Error and recovery paths
## State-path and mock-vs-production classification
## Test-definition and workflow provenance
## Repository/package hygiene
## Gaps
## Final decision
```

# Anti-false-acceptance checklist

```text
[ ] Requirements and user corrections reconstructed
[ ] Exact current source confirmed
[ ] Rule files loaded with identifiers
[ ] Complete inventory produced
[ ] Requirement matrix complete
[ ] Entire applicable source surface inspected
[ ] Security/RBAC positive and negative paths traced
[ ] Workflow and terminal paths traced
[ ] Protected fields mapped end to end
[ ] Docs checked against source
[ ] Naming and path hazards checked
[ ] Comments and TODO/FIXME/HACK/XXX checked
[ ] Formats and source paths inspected
[ ] Script metadata inspected without execution
[ ] Contamination checked
[ ] Tests, regression, CI, and deployment definitions distinguished
[ ] Static interaction/error/recovery/state paths inspected
[ ] External artifacts treated as optional read-only context
[ ] No project or CI execution performed
[ ] Review continued after every blocker through Gate 27
[ ] Report structure and links inspected
[ ] Every caveat has a static correction
```

If any applicable item is unchecked, final `PASS` is prohibited.
