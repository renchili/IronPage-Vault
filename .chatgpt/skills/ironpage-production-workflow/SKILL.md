# Prompt-Driven Project Generation Workflow

Use this skill to generate or repair a software project from user-provided input. The skill is variable-driven. It describes how the assistant should transform prompt input into a plan, code, checks, evidence, and iteration.

## User input slots

These slots are supplied by the user or by user-uploaded material:

- `{{PROJECT_PROMPT}}`: required. Original prompt, uploaded metadata, issue text, README text, or equivalent requirement source.
- `{{PROJECT_NAME}}`: optional. Project name when supplied.
- `{{TARGET_REPO}}`: optional. Repository to create or modify.
- `{{USER_GOAL}}`: required. Current goal: generate, repair, validate, package, or review.
- `{{CONSTRAINTS}}`: optional. Language, framework, database, CI, deployment, security, and non-goal constraints.
- `{{USER_FEEDBACK}}`: optional. Latest correction or requested change.

## Assistant working slots

These slots are produced by the assistant while working. They are not submitted by the user:

- `{{REQUIREMENT_LEDGER}}`: requirements extracted from `{{PROJECT_PROMPT}}`.
- `{{DELIVERY_PLAN}}`: implementation plan generated from `{{PROJECT_PROMPT}}` and `{{CONSTRAINTS}}`.
- `{{ASSUMPTIONS}}`: assumptions made when the prompt is incomplete.
- `{{OPEN_QUESTIONS}}`: blocking questions only.
- `{{CHANGE_SET}}`: files and behavior changed in the current iteration.
- `{{EVIDENCE_MAP}}`: evidence collected by the assistant from repository files, tests, CI runs, artifacts, and reports.

## Interaction contract

If a required user input slot is missing and cannot be safely inferred, ask for the smallest blocking input.

When `{{USER_FEEDBACK}}` changes the requirement, update `{{REQUIREMENT_LEDGER}}` and `{{DELIVERY_PLAN}}` before changing code.

`{{REQUIREMENT_LEDGER}}`, `{{DELIVERY_PLAN}}`, `{{CHANGE_SET}}`, and `{{EVIDENCE_MAP}}` are assistant-owned working state. Build them from the prompt, repository, and CI evidence.

## Generation algorithm

1. Read `{{PROJECT_PROMPT}}`.
2. Extract requirements into `{{REQUIREMENT_LEDGER}}`.
3. Build `{{DELIVERY_PLAN}}`: product goal, roles, workflows, data model, API surface, security rules, tests, CI, artifacts, and acceptance checklist.
4. Generate or modify the repository according to `{{DELIVERY_PLAN}}`.
5. Add tests and contract checks for ledger items.
6. Add CI and durable artifacts when evidence is required.
7. Populate `{{EVIDENCE_MAP}}` from code, tests, CI runs, artifacts, and reports.
8. Compare `{{EVIDENCE_MAP}}` back to `{{REQUIREMENT_LEDGER}}`.
9. Apply `{{USER_FEEDBACK}}` as new or corrected ledger items.
10. Repeat until required items are verified or explicitly pending.

## Requirement ledger schema

| Requirement | Source | Plan | Code evidence | Test evidence | CI evidence | Artifact evidence | Status |
|---|---|---|---|---|---|---|---|

Status values:

- `unknown`
- `planned`
- `implemented`
- `partial`
- `missing`
- `ci_pending`
- `artifact_missing`
- `verified`

## Implementation acceptance rules

Do not replace a required backend with a demo or proof-only implementation.

If `{{PROJECT_PROMPT}}` requires a DB-backed runtime, the generated project must include persistent storage, schema or migrations, a repository/store layer, and tests that prove persisted state. In-memory state is only acceptable when the prompt explicitly asks for a prototype.

If `{{PROJECT_PROMPT}}` requires a complete backend API, enumerate the required API groups in `{{REQUIREMENT_LEDGER}}` and implement each group. A small demo API must be marked `partial`.

If `{{PROJECT_PROMPT}}` requires RBAC, test capability rules, workflow-state rules, object-level access, and field visibility. Capability-only tests are insufficient.

If `{{PROJECT_PROMPT}}` requires document workflows, include upload/read/version behavior plus redaction, annotation, notification, audit, and workflow-transition APIs when those are in scope.

If `{{PROJECT_PROMPT}}` requires CI proof, distinguish static workflow files from an actual run. A workflow file alone is not CI execution evidence.

If `{{PROJECT_PROMPT}}` requires acceptance reports, generate per-stage logs as well as summary files.

## Evidence rules

Do not mark a requirement as `verified` only because code exists. Verification requires the evidence requested by the prompt.

Distinguish code existence, test existence, test execution, CI success for the exact commit, CI workspace reports, uploaded or committed reports, per-stage logs, and full-suite execution.

## Final response contract

Conclusion: `<verified | partially_verified | implemented_but_evidence_missing | not_fixed>`

Ledger summary:
1. `{{REQUIREMENT}}`: `{{STATUS}}`. Evidence: `{{EVIDENCE}}`.
2. `{{REQUIREMENT}}`: `{{STATUS}}`. Evidence: `{{EVIDENCE}}`.
3. `{{REQUIREMENT}}`: `{{STATUS}}`. Evidence: `{{EVIDENCE}}`.

Still pending:
- `{{PENDING_ITEM}}`

Do not claim yet:
- `{{UNPROVEN_CLAIM}}`
