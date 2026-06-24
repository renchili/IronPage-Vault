# Production Prompt-to-Delivery Workflow

Use this workflow for software projects that start from a product prompt and must end with working code, tests, CI results, artifacts, and a clear acceptance report.

## Goal

Turn the user's requested outcome into a traceable delivery loop:

1. Extract requirements from the prompt.
2. Build an acceptance ledger.
3. Inspect the repository before making claims.
4. Implement changes in small increments.
5. Add tests and CI coverage.
6. Preserve durable evidence such as reports or workflow artifacts.
7. Reconcile the result against the original prompt.
8. Incorporate review feedback and iterate until the ledger is closed.

## Acceptance ledger

Track requirements with this table:

| Requirement | Code evidence | Test evidence | CI evidence | Artifact evidence | Status |
|---|---|---|---|---|---|

Use conservative statuses:

- Unknown
- Missing
- Partial
- Implemented
- Implemented with CI pending
- Implemented with artifact missing
- Verified

## Repository inspection

Check the files that can prove the requirement. Typical evidence includes runtime code, migrations, API handlers, authorization logic, tests, documentation, CI workflows, workflow artifacts, committed reports, and failure logs.

Do not rely on memory, PR titles, broad CI badges, or previous summaries when current repository evidence is available.

## Implementation loop

For each acceptance gap:

1. Make the smallest coherent code change.
2. Add a regression test or contract check.
3. Update documentation when the requirement is policy-like.
4. Update CI if the proof would otherwise not run.
5. Upload or commit reports when the proof must remain visible after CI finishes.

## Evidence rules

Distinguish these states:

- Code exists.
- A test exists.
- The test ran in CI.
- CI passed for the exact commit.
- A report was generated in the CI workspace.
- A report was uploaded as a downloadable artifact.
- A full suite ran rather than a lightweight probe.

Only call a requirement verified when the required implementation and evidence both exist.

## Feedback loop

When review feedback identifies a mismatch, update the acceptance ledger and re-check the repository. Patch code, tests, docs, CI, or artifact handling as needed. Report the new status without repeating unsupported claims.

## Final report

Use this structure:

```text
Conclusion: <Verified / Partially verified / Implemented but evidence missing / Not fixed>

Acceptance ledger:
1. <Requirement>: <Status>. Evidence: <code/test/CI/artifact>.
2. <Requirement>: <Status>. Evidence: <code/test/CI/artifact>.
3. <Requirement>: <Status>. Evidence: <code/test/CI/artifact>.

Still pending:
- <missing CI, artifact, post-merge, or full-suite proof>

Do not claim yet:
- <anything not proven by current evidence>
```

## Done definition

The work is done only when required ledger items are implemented, tested, run in CI, and supported by durable evidence when evidence is required.
