# Real Interaction Testing Gates

This companion gate file belongs to the `full-project-acceptance-hard-gates` skill. Apply it whenever the target project has any user-facing, operator-facing, reviewer-facing, admin-facing, CLI, API-client, prompt-driven, workflow-driven, or manual validation interaction.

Unit tests, mocked interface tests, and synthetic API checks are not enough for acceptance when the project depends on human interaction quality, operational usability, or reliable guidance.

## Core rule

A project cannot receive final PASS for an interaction-heavy scope unless at least one realistic end-to-end interaction path is exercised and recorded with reproducible evidence.

The evidence may be:

```text
browser E2E trace
CLI transcript
terminal recording or saved command log
HTTP client collection with realistic payloads
screen recording
screenshot sequence
accessibility tree/snapshot
manual scripted walkthrough signed by artifact/log
agent-prompt transcript with expected/actual outputs
operator runbook dry run
```

A screenshot alone is not enough unless it is tied to a scripted flow, input data, expected state change, and artifact/log evidence.

## Gate A: Real user/operator flow coverage

Must identify the top user/operator flows from the original spec.

Required table:

```text
Flow | User role | Realistic input | Expected interaction | Evidence artifact | Status | Gap
```

FAIL if only service functions, isolated APIs, or mocked components are tested while the project requires user/operator interaction.

CONDITIONAL is allowed only when the interaction surface is explicitly out of scope or the project is a backend-only library with no CLI, UI, prompt, manual workflow, or operator workflow requirement.

## Gate B: Modern interaction quality

User-facing or operator-facing interactions must be modern, clear, and recoverable.

Must check:

```text
clear primary action
clear next action after success
clear recovery action after failure
loading/progress state when work can take time
empty state when no data exists
validation before destructive action
confirmation for destructive or irreversible actions
consistent labels and terminology
accessible focus/order/keyboard behavior where UI exists
copy is short, specific, and not agent-like
```

FAIL if the interface leaves the user stuck, hides errors, gives generic unusable errors, or uses internal implementation terms where user-facing language is required.

## Gate C: Reliable prompts and guidance copy

For projects with prompts, generated instructions, admin/operator guidance, runbooks, or human-readable acceptance reports, prompt/copy quality must be tested.

Must check:

```text
prompts contain the actual task constraints
prompts do not invent constraints
prompts do not include assistant process residue
prompts state what input is required
prompts state what output is expected
prompts include failure/retry guidance where applicable
prompts avoid vague words such as maybe, could, consider, next, future unless explicitly scoped
operator guidance maps to real commands and paths
```

Required table:

```text
Prompt/copy path | Intended user | Claimed action | Required input | Expected output | Failure guidance | Evidence | Status
```

FAIL if prompts or guidance are vague, misleading, unrelated to the project, or cannot be followed against the current implementation.

## Gate D: End-to-end state verification

Interaction tests must verify state changes, not only status codes or rendered text.

For each critical flow, verify at least one of:

```text
database row/state changed
file/artifact created or updated
notification/audit/event emitted
workflow state changed
response body reflects persisted state
subsequent read confirms previous write
```

FAIL if tests click buttons or call endpoints but do not verify resulting state.

## Gate E: Error, recovery, and edge-case interaction

Must exercise realistic negative paths, not only happy paths.

Required categories when applicable:

```text
missing required input
invalid input
permission denied
expired or revoked session
duplicate/replay request
large or unsupported file
network/service failure or dependency unavailable
destructive action cancellation
already-finalized or terminal workflow state
empty result set
```

FAIL if the acceptance evidence only covers success paths for critical user/operator flows.

## Gate F: No fake realism

Do not count a test as real interaction evidence if it only:

```text
calls a function directly while claiming UI/CLI coverage
checks that a route exists without executing the flow
uses toy payloads that bypass required validation
uses screenshots of static HTML without interaction
uses mocked success responses for all dependencies
uses an assistant-written report as evidence
```

FAIL if simulated or mocked evidence is presented as real interaction evidence.

## Gate G: Required final report additions

A report using this skill must include:

```text
real interaction flow table
prompt/copy reliability table
error and recovery interaction table
state verification table
mock-vs-real evidence classification
```

If the project has an interaction surface and these tables are omitted, the report is incomplete and the final verdict cannot be PASS.
