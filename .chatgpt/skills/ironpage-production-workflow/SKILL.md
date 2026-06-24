# IronPage Production Iteration Workflow Skill

Use this skill when working on IronPage-Vault or a similar generated-code project where the task is not only to write code, but to drive the whole production loop: understand the original prompt, generate or modify code, prove the implementation, reconcile gaps against the prompt, respond to user corrections, and iterate until the acceptance evidence is complete.

## Core rule

Never treat a PR title, previous assistant summary, or green CI badge as sufficient proof. The source of truth is the combination of:

1. the original prompt / requirement text,
2. the current repository code and docs,
3. the exact CI job result,
4. downloadable artifacts or committed regression evidence,
5. the user's latest correction or acceptance criteria.

When these disagree, the latest user-provided acceptance criteria and the current repository state win.

## Required workflow

### 1. Build a prompt ledger before coding

Extract the prompt into an explicit ledger. Split requirements into:

- Functional behavior.
- Security/storage guarantees.
- Role/permission/masking guarantees.
- UI or evidence requirements.
- Test/CI/reporting requirements.
- Explicit non-goals and scope notes.

For each item, record:

- Exact prompt wording or a faithful paraphrase.
- Current implementation status: `unknown`, `missing`, `partial`, `implemented`, `verified`.
- Files that must prove it.
- Tests or artifacts that must prove it.

Do not rely on memory. Re-open the repository files and CI evidence whenever the user asks whether something is fixed.

### 2. Inspect the current repo before deciding what to change

For IronPage-Vault, always check the relevant areas instead of guessing:

- `internal/app/**` for API behavior, auth, masking, auditing, document/annotation/redaction logic.
- `internal/platform/**` for crypto primitives and shared platform behavior.
- `migrations/**` for actual column layout and storage guarantees.
- `docs/metadata-security.md` for encrypted-storage matrix.
- `docs/role-field-visibility.md` for Admin / Editor / Reviewer visibility matrix.
- `docs/requirement-check.md` for prompt-to-evidence mapping.
- `run_tests.sh`, `ci/**`, `API_tests/**`, `.github/workflows/**` for proof generation and CI evidence.
- `reports/regression/**` and workflow artifacts for post-merge/full-regression proof.

If a file is not present, say it is absent. Do not infer it exists from a PR body.

### 3. Generate code in small, auditable changes

Prefer one cohesive branch per acceptance gap. Each branch should contain:

- The implementation change.
- A test, contract check, or regression guard that proves the change.
- Documentation/matrix updates when the prompt requires an auditable policy.
- CI or artifact changes when the requirement is about evidence visibility.

Do not mix unrelated product changes, CI changes, and documentation-only changes unless the user specifically asks for a combined PR.

### 4. Align implementation back to the original prompt after every PR

After writing code, re-check each ledger item:

- Is the requirement implemented in code?
- Is the requirement documented in the matrix if it is a security or role policy?
- Is there a test/contract that would fail if it regresses?
- Is CI actually running that test path?
- Is the output downloadable or otherwise visible after CI finishes?

Use a table like:

| Prompt item | Code evidence | Test/CI evidence | Artifact evidence | Status |
|---|---|---|---|---|

Allowed status words:

- `Implemented + verified`: code, test/CI, and artifact/report evidence all exist.
- `Implemented, CI pending`: code exists; CI has not finished.
- `Implemented, artifact missing`: code/test exists; evidence is not downloadable or committed.
- `Partial`: one or more explicit prompt fields are not covered.
- `Missing`: no implementation found.

Never say “all fixed” if the artifact or post-merge evidence is still missing and the user asked for production acceptance.

### 5. Treat user corrections as acceptance criteria, not as noise

When the user says “不是”, “这才是你要检测的内容”, “没有 artifact”, “这不是生产前端”, or similar, immediately update the ledger and narrow the next response to that corrected scope.

Do not repeat a previously disproven claim. Example corrections to remember for this project:

- A simple UI in `public/` is a backend testing aid, not a production frontend deliverable.
- Checking that `report.html` existed inside a CI job is not the same as providing a downloadable artifact.
- PR-scoped CI passing is not the same as post-merge full regression passing.
- `run_tests.sh` lightweight contract probe is not the full local API acceptance suite.
- The exact security prompt can require AES-256 column-level encryption even when a conventional security design like bcrypt alone would otherwise be acceptable.

### 6. Evidence hierarchy

Use this hierarchy when answering acceptance questions:

1. Current code on `main` or the PR branch.
2. CI run status for the exact head SHA or merge commit.
3. Job steps showing the relevant check ran, not merely that the workflow exists.
4. Uploaded artifact list and downloadable artifact metadata.
5. Committed `reports/regression/.../summary.json` or `summary.md`.
6. Logs for the failed stage if any stage failed.

If artifacts are empty, state that directly:

> The job created the report in its workspace, but no downloadable artifact is available.

Do not convert that into “artifact verified”.

### 7. IronPage production acceptance checklist

For IronPage-Vault, the final production-style acceptance answer must cover at least the following when relevant.

#### Security and encrypted storage

- Password verifier storage:
  - Confirm bcrypt verifier is additionally AES-256-GCM sealed before writing `users.password_hash` when the prompt requires column-level AES encryption.
  - Confirm login opens the sealed verifier before bcrypt comparison.
- PII metadata matrix:
  - Confirm coverage for redaction geometry, redaction reason, annotation comment.
  - Also confirm username, display name, document title, notification message, audit source IP, and audit metadata when requested.
- Crypto primitive:
  - Confirm `EncryptString` uses AES-GCM and a 32-byte/AES-256 key derivation path.
- Plain columns:
  - Identify whether plain columns are blank compatibility placeholders, deterministic lookup keys, or legacy fallback.
  - Do not claim “plaintext is gone” if legacy fallback still exists; describe the actual migration compatibility behavior.

#### Role-based masking and visibility

- Confirm `users.password_hash` is never serialized.
- Confirm ciphertext companion fields are hidden from all roles.
- Confirm Admin / Editor / Reviewer rules are documented and reflected in API behavior.
- Confirm object-level access checks happen before opening protected document fields.
- Confirm redaction list responses do not expose geometry/reason if that is the rule.

#### UI and report evidence

- State clearly whether the UI is production frontend or backend testing aid.
- If it is a backend testing aid, do not call it a full frontend deliverable.
- Confirm `run_tests.sh` runs UI screenshot acceptance only if the script is included and invoked.
- Confirm headless browser creates screenshot evidence only if the CI/local run actually generated it.
- Confirm downloadable artifact exists by checking workflow artifacts, not only job logs.

#### CI categories

Always distinguish:

- PR scoped CI.
- Docker build diagnostic.
- Local entrypoint contract probe.
- Full local acceptance suite.
- Post-merge full regression.
- Downloadable artifact availability.

A correct final statement may look like:

> PR CI passed and post-merge full regression passed. The local-entrypoint artifact is downloadable from run X. This proves the lightweight run_tests contract, not the full local API acceptance suite. Full suite proof comes from post-merge regression summary Y.

### 8. Failure-handling loop

When CI fails:

1. Identify the exact workflow, run id, job, failed step, and failed stage.
2. Read the artifact or log for the failing stage.
3. Quote or summarize the failing error precisely.
4. Map the error to a code path.
5. Fix the smallest root cause.
6. Add a regression guard if the failure represents a prompt acceptance gap.
7. Open or update a PR.
8. Re-check CI and artifacts.
9. Report what is fixed and what is still pending.

Never say “wait” or promise background work. Perform the next check or produce the best current state in the same response.

### 9. User communication pattern

Use concise updates while working:

- First update: state the scoped plan and what evidence will be checked.
- Mid-work updates: report concrete findings, especially failed stages or missing artifacts.
- Final answer: separate facts from pending items.

Use precise language:

- “merged” only after the PR is actually merged.
- “CI passed” only for the exact workflow/commit that passed.
- “post-merge passed” only after the post-merge run or committed regression record says passed.
- “artifact downloadable” only after artifact listing shows it.
- “full suite passed” only when the full suite, not a contract probe, ran and passed.

### 10. PR creation rules

When opening a PR:

- Title should state the acceptance gap, not just the implementation detail.
- Body must include:
  - Summary.
  - Root cause or missing proof.
  - Files changed.
  - How the change is verified.
  - Scope note, especially if UI is test aid rather than production UI.
- After creating the PR, check the exact head SHA workflows.
- If artifact visibility is the point of the PR, check `fetch_workflow_run_artifacts` before claiming success.

### 11. Final acceptance answer template

Use this shape when the user asks “是否都修复了?”

```text
结论：<已修复 / 部分修复 / 未修复 / 证据不足>

逐项：
1. <prompt item>: <status>. Evidence: <code/doc/test/artifact>.
2. <prompt item>: <status>. Evidence: <code/doc/test/artifact>.
3. <prompt item>: <status>. Evidence: <code/doc/test/artifact>.

仍需注意：
- <pending CI/artifact/full-suite distinction if any>.

不能声称：
- <any thing not proven yet>.
```

If the user's latest message corrects the acceptance scope, use their corrected scope as the table rows.

## Anti-patterns

Avoid these mistakes:

- Saying “fixed” because a PR was opened.
- Saying “verified” because a script exists but was not run.
- Saying “artifact exists” because a file existed inside the CI workspace.
- Treating a test UI as production frontend.
- Treating PR CI as post-merge full regression.
- Treating a lightweight `run_tests.sh` contract probe as the full acceptance suite.
- Ignoring a user correction and re-answering the previous question.
- Relying on assistant memory when the repository can be queried.

## Minimum done definition

A production iteration is done only when:

1. The original prompt ledger has no `missing` or unintended `partial` items.
2. Code and docs are aligned.
3. Tests or contract checks cover the acceptance gap.
4. Relevant CI completed successfully.
5. Required artifacts are downloadable or committed.
6. The final answer clearly states any remaining distinction between local, PR, and post-merge evidence.
