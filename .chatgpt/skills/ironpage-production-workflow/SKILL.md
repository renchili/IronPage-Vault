# Prompt-Driven Project Generation Workflow

Use this skill to generate or repair a software project from user-provided input. The skill is variable-driven. Do not store conversation text in this file.

## Input slots

- `{{PROJECT_PROMPT}}`: required. The original user-provided prompt, uploaded metadata, issue text, README text, or equivalent requirement source.
- `{{PROJECT_NAME}}`: optional. Use when provided; otherwise infer from `{{PROJECT_PROMPT}}` and confirm only if blocking.
- `{{TARGET_REPO}}`: optional. Repository to create or modify.
- `{{USER_GOAL}}`: required. What the user wants done now: generate, repair, validate, package, or review.
- `{{CONSTRAINTS}}`: optional. Language, framework, database, CI, deployment, security, and non-goal constraints.
- `{{USER_FEEDBACK}}`: optional. Latest correction or requested change from the user.

## Working state slots

Maintain these internal artifacts while working:

- `{{REQUIREMENT_LEDGER}}`: prompt requirements mapped to plan, code, tests, CI, artifacts, and status.
- `{{DELIVERY_PLAN}}`: implementation plan derived from `{{PROJECT_PROMPT}}`.
- `{{ASSUMPTIONS}}`: assumptions made because the prompt does not specify something.
- `{{OPEN_QUESTIONS}}`: only questions that block correct implementation.
- `{{CHANGE_SET}}`: files and behavior changed in the current iteration.
- `{{EVIDENCE_MAP}}`: proof for each requirement, including code, tests, CI, artifacts, and reports.

## Interaction contract

If a required slot is missing and cannot be inferred safely, ask the user for that slot. Ask only the smallest blocking question.

If `{{USER_FEEDBACK}}` contradicts `{{DELIVERY_PLAN}}` or `{{REQUIREMENT_LEDGER}}`, update the plan and ledger first, then modify code or docs.

Do not continue from an outdated interpretation after the user corrects the requirement.

## Generation algorithm

1. Read `{{PROJECT_PROMPT}}`.
2. Extract requirements into `{{REQUIREMENT_LEDGER}}`.
3. Build `{{DELIVERY_PLAN}}` with product goal, roles, workflows, data model, API surface, security rules, tests, CI, artifacts, and acceptance checklist.
4. Generate or modify the repository according to `{{DELIVERY_PLAN}}`.
5. Add tests and contract checks for the ledger items.
6. Add CI and durable artifacts when evidence is required.
7. Compare code and evidence back to `{{REQUIREMENT_LEDGER}}`.
8. Apply `{{USER_FEEDBACK}}` as new or corrected ledger items.
9. Repeat until all required items are verified or explicitly marked pending.

## Requirement ledger schema

Use this schema:

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

## Evidence rules

Never mark a requirement as `verified` only because code exists. Verification needs the evidence required by the prompt.

Distinguish:

- code exists;
- test exists;
- test ran;
- CI passed for the exact commit;
- report exists inside a CI workspace;
- report was uploaded or committed;
- full acceptance suite ran.

## Final response contract

Report using current slot values:

Conclusion: `<verified | partially_verified | implemented_but_evidence_missing | not_fixed>`

Ledger summary:
1. `{{REQUIREMENT}}`: `{{STATUS}}`. Evidence: `{{EVIDENCE}}`.
2. `{{REQUIREMENT}}`: `{{STATUS}}`. Evidence: `{{EVIDENCE}}`.
3. `{{REQUIREMENT}}`: `{{STATUS}}`. Evidence: `{{EVIDENCE}}`.

Still pending:
- `{{PENDING_ITEM}}`

Do not claim yet:
- `{{UNPROVEN_CLAIM}}`
