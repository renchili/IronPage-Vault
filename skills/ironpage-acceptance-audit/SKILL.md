---
name: full-project-acceptance-methodology
description: Methodology for accepting a complete software project from original requirement to final pass, conditional pass, or fail decision.
---

# Full Project Acceptance Methodology

Use this skill when the user wants to know **how to perform full-project acceptance**.

This is not a record of one previous project result. It is a reusable acceptance method.

## Goal

Given an original requirement and a current repository, determine:

1. what must be checked;
2. where each requirement is implemented;
3. what evidence proves it;
4. which tests or workflows validate it;
5. what artifacts were generated;
6. what gaps remain;
7. whether the project is pass, conditional pass, or fail.

## Evidence order

Prefer evidence in this order:

1. executed test logs and generated artifacts;
2. CI workflow runs and uploaded artifacts;
3. generated summaries committed by automation;
4. source code and schema inspection;
5. static contract scripts;
6. API and product documentation;
7. manual smoke UI;
8. user claims;
9. assistant-written summary.

Assistant-written reports are review summaries only. They are not test artifacts.

## Acceptance workflow

### 1. Build the requirement matrix

Read the original spec or prompt. Split it into atomic requirements.

Use this table:

```text
ID | Requirement | Category | Expected behavior | Implementation | Evidence | Status | Gap
```

Categories:

```text
architecture
functionality
data model
workflow
access control
API contract
runtime/deployment
storage
observability/audit
backup/restore
UI/manual validation
documentation
testing/CI
```

### 2. Map requirements to implementation

For each requirement, inspect:

```text
routes
handlers
services
repositories
migrations
configuration
middleware
scripts
Docker/deployment files
documentation
```

Mark each item as:

```text
Implemented
Partially implemented
Documented only
Missing
Cannot verify
```

### 3. Verify critical controls separately

Do not hide critical controls inside general feature checks.

Check identity, permissions, sensitive data handling, session behavior, auditability, and configuration behavior as separate sections.

For any protected-data claim, verify:

```text
helper exists
write path stores protected value
read path opens it only where allowed
plain compatibility fields are not source of truth
tests or static checks guard the behavior
```

### 4. Verify roles and denied paths

Build a role matrix:

```text
Role | Allowed actions | Forbidden actions | Visible fields | Hidden fields | Positive evidence | Negative evidence
```

A project is not fully accepted if only allowed paths are tested and denied paths are ignored.

### 5. Verify state machine and business rules

Check:

```text
required states
valid transitions
invalid transition rejection
terminal-state immutability
history rows
audit rows
positive tests
negative tests
```

### 6. Verify test entrypoints

For every test script or command, identify:

```text
command
normal mode
probe mode
required services
output directory
generated reports
logs
skipped stages
exit behavior
```

Never confuse probe evidence with full-suite evidence.

Example:

```text
./run_tests.sh                  = full local test entrypoint
PROBE=1 ./run_tests.sh          = lightweight probe
ci/run_full_regression.sh <dir> = full regression entrypoint
```

### 7. Verify CI workflows

Read workflow files. Check:

```text
trigger
path filters
jobs
conditions
skipped jobs
commands
artifact upload
retention
failure artifact behavior
```

A green workflow is not enough if the relevant job was skipped.

### 8. Verify generated artifacts

For every generated artifact, record:

```text
artifact path/name
source command/workflow
commit or run id
full-suite or probe
summary status
stage count
failed stages
skipped stages
logs
```

Download artifacts when available. Parse summary files. Do not infer pass from artifact existence alone.

### 9. Verify API, docs, deployment, and manual UI

Check:

```text
API route coverage
request/response contract
status codes
error envelope
pagination
setup docs
operation docs
backup/restore docs
Docker/build packaging
manual smoke UI if required
```

Docs support acceptance but do not replace runtime evidence.

## Gap classification

Classify every gap:

```text
P0 blocker: cannot accept
P1 conditional: can accept with explicit caveat
P2 quality issue: non-blocking improvement
Evidence gap: implementation may exist but proof is missing
Spec interpretation gap: requirement wording needs decision
```

## Decision rubric

### Pass

Use when:

```text
requirements are implemented
critical controls are verified
full regression or equivalent evidence passed
test artifacts are understood
docs and deployment are adequate
remaining caveats are non-blocking
```

### Conditional pass

Use when:

```text
core project works
but one evidence source is missing or probe-only
or a non-critical item lacks strong proof
```

### Fail

Use when:

```text
core requirement missing
critical control contradicts spec
build/runtime/regression fails
critical evidence missing
conclusion depends on stale or fabricated evidence
```

## Report template

```markdown
# Full Project Acceptance Report

## 1. Scope
- Repository:
- Branch/commit:
- Requirement source:
- Runtime target:
- Acceptance standard:
- Timestamp:

## 2. Executive decision
- Verdict:
- Reason:
- Blocking gaps:
- Caveats:

## 3. Requirement coverage matrix
| ID | Requirement | Category | Implementation | Evidence | Status | Notes |
|---|---|---|---|---|---|---|

## 4. Implementation verification

## 5. Critical controls verification

## 6. Test evidence
| Source | Command/workflow | Full/probe | Result | Artifact/log |
|---|---|---|---|---|

## 7. CI and artifact evidence

## 8. Manual/UI evidence

## 9. Documentation/API/deployment evidence

## 10. Gaps and risks
| Severity | Gap | Why it matters | Required fix/evidence |
|---|---|---|---|

## 11. Final acceptance decision
```

## Anti-false-acceptance checklist

```text
[ ] Current code checked
[ ] Original requirements reconstructed
[ ] Requirement matrix built
[ ] Implementation mapped
[ ] Critical controls checked separately
[ ] Denied paths checked
[ ] CI skipped jobs inspected
[ ] Artifacts downloaded or parsed when available
[ ] Probe and full-suite distinguished
[ ] Assistant report not treated as test artifact
[ ] Caveats stated clearly
```

## Handling user corrections

If the user says the audit is wrong:

1. Stop defending the old conclusion.
2. Identify the wrong assumption.
3. Re-read current files or artifacts.
4. Update the matrix and verdict.
5. State exactly what changed.

If the user asks for methodology, provide the method, not a previous result.
