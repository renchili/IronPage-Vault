# Documentation Pollution Gates

This companion gate file belongs to the `full-project-acceptance-hard-gates` skill. Apply it whenever the target repository or package contains README, design, architecture, question, FAQ, prompt, planning, or generated documentation files.

Documentation must describe the current project only. It must match the implementation, tests, commands, paths, and evidence exactly. It must not contain agent process residue, invented obligations, or speculative roadmap items that are unrelated to the project contract.

## Scope

Apply these checks especially to:

```text
README.md
DESIGN.md
ARCHITECTURE.md
QUESTION.md
QUESTIONS.md
FAQ.md
PROMPT_USED.md
DELIVERY_PLAN.md
EVIDENCE_MAP.md
REQUIREMENT_LEDGER.md
docs/**/*.md
```

## Gate A: Project-only documentation

FAIL if documentation contains unrelated assistant process notes, model self-corrections, conversational residue, or content that is not part of the project contract.

Red-flag wording includes:

```text
next step
next steps
follow-up
future work
we can also
could also
recommended next
suggested improvement
nice to have
assistant notes
model notes
previous mistake
correction
```

These phrases are not automatically forbidden, but each occurrence must be justified by an original requirement, implemented feature, test, tracked issue, or explicit out-of-scope note. Otherwise mark it as documentation pollution.

## Gate B: No invented project obligations

Agents must not turn their own mistakes, caveats, or speculative improvements into project requirements.

FAIL if README, design, question, FAQ, or architecture docs state that the project must obey a behavior, architecture, rule, limitation, or roadmap item that is not present in the original requirement, implementation, test suite, or tracked issue.

Required check table:

```text
Doc path | Claimed obligation | Source of obligation | Implementation/test path | Valid? | Gap
```

If the source of obligation is only an assistant report, generated summary, or reviewer opinion, the claim is not valid project evidence.

## Gate C: README, design, and question docs consistency

README, design, and question or FAQ docs must be mutually consistent.

Compare:

```text
project purpose
runtime model
supported features
unsupported or out-of-scope features
security model
API behavior
storage model
deployment model
test commands
acceptance status
known limitations
```

FAIL if README says a feature is supported while design or question docs say it is future work, optional, unknown, or not implemented.

FAIL if question or FAQ docs answer based on an older project version and contradict current implementation.

## Gate D: No process roadmap unless explicitly required

Generic roadmap sections are not accepted unless the original spec requested a roadmap or the items are linked to tracked issues.

Red-flag headings:

```text
Next Steps
Future Work
Recommendations
Suggested Improvements
Possible Enhancements
Roadmap
TODO
Open Questions
```

These sections are FAIL unless every item is one of:

```text
explicitly requested by original spec
implemented and tested
tracked by an issue or task ID
clearly labeled as out of scope and not required for acceptance
```

## Gate E: Documentation status must match evidence status

Docs must not claim complete, accepted, verified, deployed, or production-ready status unless evidence supports that exact status.

FAIL if docs claim a status that is contradicted by evidence that is missing, not verified, probe-only, local-only, or conditional.

## Gate F: Agent residue scan

Every acceptance report must include a documentation pollution scan table:

```text
Doc path | Pollution pattern | Excerpt | Classification | Required action
```

Classifications:

```text
OK               legitimate project documentation
P2 wording       harmless wording issue
P1 pollution     unrelated agent/process text or speculative roadmap
P0 contradiction doc claim contradicts implementation, tests, or evidence
```

A project cannot receive final PASS with unresolved P0 contradiction or P1 pollution in README, design, question, architecture, setup, API, security, or acceptance docs.

## Gate G: Required final report additions

A report using this skill must include:

```text
documentation pollution scan
README/design/question consistency table
invented-obligation table
roadmap/future-work validation table
```

If these tables are omitted while such docs exist, the report is incomplete and the final verdict cannot be PASS.
