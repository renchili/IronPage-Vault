# Prompt-Driven Project Generation Workflow

Use this workflow to generate or repair a software project from an original product prompt. The file is English-only and contains process rules, not conversation history.

## Main loop

1. Use the user-provided original prompt input, such as uploaded metadata, issue text, README requirements, or a supplied prompt.
2. Extract the prompt into requirements.
3. Convert the requirements into a delivery plan.
4. Generate the project structure from the plan.
5. Generate runtime code, schema, tests, CI, and reports.
6. Compare the generated project back to the prompt.
7. Turn review feedback into new ledger items.
8. Patch the plan and code until every required item is verified.

## Requirement ledger

Track every prompt item with this table:

| Requirement | Plan | Code evidence | Test evidence | CI evidence | Artifact evidence | Status |
|---|---|---|---|---|---|---|

Statuses:

- Unknown
- Planned
- Implemented
- Partial
- Missing
- CI pending
- Artifact missing
- Verified

## Plan generation

Before coding, produce a plan with these sections:

1. Product goal.
2. Actors and roles.
3. Core workflows.
4. Data model and migrations.
5. API endpoints and error model.
6. Security, privacy, and access-control rules.
7. UI or evidence UI requirements.
8. Test and regression strategy.
9. CI and artifact strategy.
10. Acceptance checklist.

The plan is the bridge between the prompt and generated code. If the plan misses a prompt requirement, the generated project will be incomplete.

## Project generation

Generate a complete repository, not isolated snippets. A backend project normally needs:

- module and dependency files;
- server entrypoint;
- router and handlers;
- domain models;
- database migrations;
- auth and authorization middleware;
- validation and error handling;
- storage and security helpers;
- unit tests;
- API tests;
- contract checks;
- local test runner;
- Docker files;
- CI workflows;
- acceptance reports or artifacts.

For this repository, regenerate the concrete product by using the user-provided original prompt input first, then deriving the document workflow, roles, APIs, storage model, security requirements, tests, and CI evidence from that prompt.

## Self-check

After generation, inspect the repository again. For every ledger row, check:

- the code path exists;
- the schema supports the behavior;
- all roles and edge cases are covered;
- tests or contracts prove the behavior;
- CI runs the relevant checks for the exact commit;
- reports or artifacts are durable when evidence is required.

A generated file inside a CI workspace is not durable evidence unless uploaded or committed.

## Review iteration

When review feedback finds a mismatch, update the ledger and plan. Then patch code, tests, docs, CI, or artifact handling. Do not mark the item verified until evidence exists.

## Final answer

Report completion like this:

Conclusion: <Verified / Partially verified / Implemented but evidence missing / Not fixed>

Acceptance ledger:
1. <Requirement>: <Status>. Evidence: <code/test/CI/artifact>.
2. <Requirement>: <Status>. Evidence: <code/test/CI/artifact>.
3. <Requirement>: <Status>. Evidence: <code/test/CI/artifact>.

Still pending:
- <missing code, CI, artifact, post-merge, or full-suite proof>

Do not claim yet:
- <anything not proven by current evidence>

## Done definition

The work is done only when the prompt ledger, plan, generated code, tests, CI, artifacts, and final answer are aligned.
