---
name: full-project-acceptance-hard-gates
description: Read-only, hard-gated methodology for accepting or rejecting a complete software project from current source and pre-existing evidence without reviewer execution.
---

# Full Project Acceptance Hard Gates

Use this Skill to decide whether a complete software project should be accepted or rejected. The target may be a repository, branch, commit, pull request, generated package, or archive.

Keep this Skill reusable. Project names, repository identifiers, pull request numbers, revisions, and findings belong in the acceptance report, not in this file.

## Mandatory acceptance mode: static and read-only

Acceptance review is static and read-only unless the user explicitly authorizes a different mode in the current request.

During static acceptance, the reviewer must not:

- run project code;
- run project scripts, test entrypoints, linters, formatters, generators, or migrations;
- build binaries, images, packages, or documentation;
- start containers, services, databases, browsers, or deployments;
- issue live API or UI interactions;
- trigger, retry, rerun, create, approve, or wait for CI execution;
- create evidence intended to fill a missing runtime or interaction gap.

The reviewer may inspect current source, configuration, documentation, existing completed CI metadata, and existing retained artifacts read-only.

Missing execution evidence must be recorded as `NOT VERIFIED`. The reviewer must never execute work merely to turn `NOT VERIFIED` into `PASS`.

A reviewer-authored report is a review summary only. It is never implementation, test, runtime, interaction, deployment, or CI evidence.

## Core acceptance rule

A project does not pass because a route exists, a test file exists, a report exists, a screenshot exists, CI is green, or a pull request was merged.

A project passes only when:

1. the original requirement is reconstructed into an atomic requirement matrix;
2. every material requirement maps to current implementation and current evidence paths;
3. repository structure, naming, formats, comments, packaging, and contamination are checked;
4. documentation agrees with source and evidence;
5. security, authorization, workflow, persistence, and negative paths are independently assessed;
6. runtime and interaction claims are supported by pre-existing evidence for the inspected revision;
7. static checks, local probes, full regression, CI, deployment, and reviewer reports are not confused;
8. every required report section is produced;
9. no blocking gate fails.

## Evidence hierarchy

Prefer evidence in this order:

1. Pre-existing end-to-end interaction evidence tied to the inspected revision.
2. Pre-existing executed test logs and retained artifacts tied to the inspected revision.
3. Pre-existing completed CI runs and downloadable artifacts tied to the inspected revision.
4. Generated summaries produced by those completed executions.
5. Current source, migrations, configuration, manifests, scripts, and comments.
6. Static and contract guards.
7. Current documentation.
8. User claims.
9. Reviewer summaries.

Do not reuse evidence from another revision without an explicit source-tree equivalence proof and a caveat describing what was and was not revalidated.

## Status vocabulary

Use only:

```text
PASS          Implemented and directly supported by current evidence.
CONDITIONAL   Mostly acceptable, but evidence is incomplete, indirect, local-only, probe-only, or environment-limited.
FAIL          Required behavior is missing, contradicted, misleading, malformed, non-portable, or unsafe.
NOT VERIFIED  Required evidence was unavailable or was not checked.
N/A           Not required by the original specification; reason is mandatory.
```

Final `PASS` is allowed only when every required gate is `PASS` or justified `N/A`. Any required `CONDITIONAL`, `FAIL`, or `NOT VERIFIED` prevents final `PASS`.

## Gap severity

```text
P0 blocker       Cannot accept.
P1 conditional   Acceptance requires an explicit caveat and follow-up evidence.
P2 quality       Non-blocking maintainability or presentation issue.
Evidence gap     Implementation may exist, but proof is missing.
Spec gap         Original requirement is ambiguous.
Packaging gap    Path, file, permission, format, or archive issue affects reproducibility.
Doc-code gap     Documentation or comments contradict implementation or evidence.
Interaction gap  Real user or operator behavior is missing or only simulated.
Code-quality gap Naming, structure, language idiom, or comments harm maintainability.
```

## Mandatory preflight inventory

Before judging implementation, record:

```text
repository or package path
branch, commit, tag, PR head, merge-test commit, or archive SHA256
file count and root layout
source directories
test directories
documentation files
scripts and executable modes
workflow files
deployment files
migrations and configuration
artifact directories
binary, large, generated, cache, secret, and runtime files
loaded rule and Skill paths with stable identifiers
reviewer execution performed: none, unless explicitly authorized
```

For archives, compare original entries and modes with extracted entries and modes. No inventory means no final `PASS`.

# Hard gates

## Gate 0: Evidence provenance

Confirm the exact target revision or package hash, evidence revision, merge-test revision where applicable, source-tree equivalence, and existence of every cited path. Separate reviewer reports from generated execution artifacts.

FAIL for invented paths, stale revisions, hidden differences, or reused conclusions without current inspection.

## Gate 1: Requirement coverage

Reconstruct the specification into atomic requirements covering architecture, deployment, persistence, storage, API, workflow, security, authentication, protected data, domain behavior, audit, notifications, backup, interactions, documentation, tests, CI, artifacts, repository hygiene, and maintainability.

Required table:

```text
ID | Requirement | Category | Implementation path | Evidence path | Status | Gap
```

## Gate 2: Architecture and deployment model

Check required language, framework, persistence, storage, dependency posture, startup assets, build/package files, and deployment boundary. Static review may prove structure and configuration ownership; actual build/start behavior remains `NOT VERIFIED` without pre-existing current evidence.

## Gate 3: Data model and persistence

Check required entities, schema constraints, source-of-truth fields, migrations, transactions, histories, read/write alignment, and restart persistence. Do not infer runtime persistence from schema presence alone.

## Gate 4: Authentication, session, freshness, and replay

Check credential protection, lockout, session/token creation, expiration, revocation, freshness, replay protection, fail-closed persistence, and positive/negative evidence. Missing stateful evidence is `NOT VERIFIED`.

## Gate 5: Protected data and sensitive fields

For every sensitive field, map classification, schema, write protection, read/decrypt path, key source, protected format, API exposure, masking, and guards. FAIL when protected plaintext remains the source of truth.

## Gate 6: RBAC, object access, and field visibility

Required table:

```text
Role | Allowed | Forbidden | Object scope | Visible fields | Hidden fields | Positive evidence | Negative evidence
```

FAIL when forbidden paths or object scope are absent or contradicted. Test definitions without execution prove intent only.

## Gate 7: Workflow and terminal immutability

Check states, allowed and rejected transitions, terminal-state immutability across every mutator, history, audit, notifications, and evidence.

## Gate 8: Domain feature completeness

For each core feature map endpoint or command, service/domain logic, persistence, side effects, errors, positive and negative tests, realistic artifact, and implementation path. Route names and stubs are insufficient.

## Gate 9: Audit, notifications, and configuration side effects

Check every required mutation, audit record, notification, protected metadata, acknowledgement behavior, editable configuration/templates, and authorization.

## Gate 10: API contract, errors, and pagination

Check generated contract sources, route coverage, request/response schemas, authentication annotations, status codes, error envelopes, pagination defaults and ceilings, and contract consistency. Runtime response behavior is not proven by static route coverage.

## Gate 11: Backup, restore, and operations

Check scheduling, dump and filesystem artifact paths, restore mechanism, recovery documentation, configured paths, and pre-existing executed evidence. Documentation or static calls alone do not prove recoverability.

## Gate 12: Local test entrypoints

For each entrypoint record command, probe mode, dependencies, stages, outputs, summaries, logs, exit behavior, modes, and nested invocations. A probe proves only what it actually records.

## Gate 13: Full regression

Check that a separate full-regression entrypoint exists and that any claimed pass has a retained summary and artifact tied to the target revision. Static acceptance must not execute it. Without current evidence, use `NOT VERIFIED`, not `FAIL`, unless the entrypoint is statically broken.

## Gate 14: CI admission, execution safety, and artifacts

Inspect workflow source and pre-existing run metadata read-only.

Required controls:

- one explicit target namespace shared by all relevant events;
- at most one active run per target through platform concurrency;
- duplicate events collapse rather than accumulate queued work;
- admission occurs before checkout or any repository-controlled code;
- admission rejects rather than sleeps inside a runner;
- failure history is fully paginated or stored in a dedicated durable control record;
- a failed target/revision remains latched until a new revision or a bounded reviewed unlock;
- unlock identifies the exact target, revision, failed run, reviewer reason, and one-time consumption record;
- ordinary rerun buttons do not bypass the latch;
- later jobs and artifact uploads do not start after a failed gate;
- retained artifacts identify the exact revision and command scope;
- static inventory evidence is retained only after all static gates succeed.

GitHub Actions workflow YAML cannot prove that no workflow-run object or runner is ever created. When the specification literally requires pre-dispatch prevention, acceptance additionally requires repository/platform ruleset, app, or external admission evidence. Without that evidence, mark the requirement `NOT VERIFIED` or `FAIL`; never overstate repository YAML.

## Gate 15: Manual UI and smoke surfaces

Check whether the surface exists, is mode-gated, included in deployable assets, uniquely named, and correctly described. A static screenshot or source file is not interaction evidence.

## Gate 16: Documentation and implementation consistency

Compare setup, commands, paths, environment variables, dependencies, routes, DTOs, statuses, security, storage, workflow, backup, deployment, verification scope, evidence status, and limitations.

Required table:

```text
Doc claim | Document path | Implementation path | Evidence path | Match? | Severity | Required correction
```

## Gate 17: Repository and package layout

Check expected roots, stable source/test/docs/scripts/migrations/deploy paths, duplicate or conflicting files, misplaced implementation, and hidden generated copies.

Path audit must generically detect:

- non-ASCII path names unless explicitly required;
- whitespace and control characters;
- case-only collisions;
- near-duplicate sibling names;
- mixed naming conventions inside one path segment;
- legacy and canonical directory pairs;
- explicitly documented exceptions with distinct roles.

## Gate 18: File format, encoding, and content hygiene

Check UTF-8 or declared encoding, JSON/YAML/TOML structure, Markdown, SQL/shell syntax, shebangs, line endings, extension/content agreement, placeholders, and bundled runtime files. During static review, inspect existing parser results rather than running parsers.

## Gate 19: Source and evidence path validation

Required table:

```text
Claim | Implementation path | Test path | Artifact/log path | Exists? | Current revision? | Notes
```

## Gate 20: Idiomatic naming and readable code

Check language/framework conventions for files, modules, packages, types, functions, variables, configuration, routes, DTOs, and tests. Critical behavior must be decomposed and named by domain intent.

## Gate 21: Comment quality and consistency

Enumerate TODO/FIXME/HACK/XXX markers, generated-code notices, stale comments, and claims about security, workflow, storage, CI, cooldown, latches, and unlock behavior. Comments must match implementation and documentation.

## Gate 22: Script permissions and portable execution

Inspect modes, shebangs, interpreter assumptions, nested paths, working-directory assumptions, dependency checks, and packaging preservation. Do not execute scripts during static acceptance.

## Gate 23: Source-package contamination

Scan tracked inventory and source packages for caches, compiled objects, databases, coverage, logs, temporary files, secrets, real environment files, editor files, stale reports, and generated output. The complete manifest must be retained when the project claims a retained static acceptance artifact.

## Gate 24: Report schema and rendering

Validate that all required sections and tables exist, statuses use the approved vocabulary, every non-PASS row includes its gap, links and paths exist, and the verdict matches the rows. Do not run rendering tools during static acceptance.

## Gate 25: Documentation pollution and invented obligations

Scan README, design, architecture, question/FAQ, planning, prompt, requirement, evidence, and review documents for assistant residue, stale process narratives, unsupported roadmap items, invented requirements, and claims that exceed evidence.

## Gate 26: Real interaction and reliable guidance

For every critical user/operator/API/browser/CLI flow, inspect pre-existing realistic evidence for positive, negative, recovery, accessibility, and resulting-state verification. Unit tests, mocks, direct service calls, route checks, and static screenshots are not real-interaction evidence. Static acceptance must not create missing interaction evidence.

Required tables:

```text
Flow | Role | Realistic input | Expected interaction | Actual result | State verification | Evidence artifact | Status | Gap
Negative flow | Trigger | Expected recovery | Actual result | Evidence | Status
Evidence item | Real or mock | What it proves | What it does not prove
```

## Gate 27: Final verdict

Verdict rules:

```text
PASS          Every required gate is PASS or justified N/A.
CONDITIONAL   No P0 exists, but at least one required gate is CONDITIONAL or NOT VERIFIED.
FAIL          Any P0 exists or a required core behavior is missing or contradicted.
```

# Required report template

```markdown
# Full Project Acceptance Report

## Scope
## Executive verdict
## Repository/package inventory
## Requirement matrix
## Hard gate results
## Source and evidence path validation
## Documentation-code consistency
## Code readability and naming
## Comment consistency
## Documentation pollution scan
## README/design/question consistency
## Invented obligations and roadmap validation
## Real interaction flows
## Prompt/copy reliability
## Error and recovery interactions
## State verification and mock-vs-real classification
## Test and artifact provenance
## CI execution safety
## Acceptance Skill audit
## Runtime configuration audit
## Repository/package hygiene
## Gaps
## Final decision
```

# Anti-false-acceptance checklist

```text
[ ] Original requirements reconstructed
[ ] Exact source revision or package hash confirmed
[ ] Rule and Skill metadata recorded
[ ] Reviewer execution recorded as none unless explicitly authorized
[ ] No code, scripts, tests, builds, containers, browsers, databases, deployments, or CI were triggered by the reviewer
[ ] Repository/package inventory produced
[ ] Requirement matrix complete
[ ] Security and RBAC negative paths assessed
[ ] Workflow invalid paths and terminal immutability assessed
[ ] Documentation checked against source and evidence
[ ] Naming, path collisions, formats, comments, and contamination checked
[ ] Probe, static checks, full regression, CI, deployment, and reviewer reports distinguished
[ ] CI admission, cooldown, latch, pagination, duplicate collapse, and one-time unlock inspected
[ ] Existing interaction evidence classified as real or mock
[ ] Missing execution or interaction evidence marked NOT VERIFIED rather than generated
[ ] Every cited path exists in the inspected revision
[ ] Every caveat includes the required correction or evidence
[ ] Final verdict matches all gate rows
```

If any required item is unchecked, final `PASS` is prohibited.
