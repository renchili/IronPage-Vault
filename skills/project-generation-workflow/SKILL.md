# Prompt-Driven Project Generation Workflow

Use this skill to generate, repair, validate, package, or review a software project from user-provided requirements. Preserve the target repository, its architecture, its conventions, and the distinction between project truth and agent working history.

## Required inputs

- `{{PROJECT_PROMPT}}`: original prompt, issue, specification, uploaded metadata, or equivalent authoritative requirement source.
- `{{PROJECT_NAME}}`: optional project name.
- `{{TARGET_REPO}}`: optional repository.
- `{{REPO_ROOT}}`: repository root when repository work is involved.
- `{{USER_GOAL}}`: generate, repair, validate, package, or review.
- `{{CONSTRAINTS}}`: technical constraints and non-goals.
- `{{USER_FEEDBACK}}`: latest correction or confirmed requirement.

## Rule precedence

Use sources in this order:

1. latest explicit user instruction;
2. original project prompt or specification;
3. committed repository rules such as `AGENT.md`;
4. current implementation, tests, schemas, scripts, and configuration;
5. existing project documentation;
6. agent analysis and working notes.

Lower-priority sources must not override higher-priority sources. Agent analysis, generated reports, failed attempts, commit messages, PR descriptions, test failures, and previous generated documents are never project requirements by themselves.

## Repository boundaries

When a repository is involved:

- Treat `{{REPO_ROOT}}` as the only project root.
- Inspect the current tree, branch, changed files, source, tests, migrations, scripts, CI, deployment files, docs, and generated artifacts before editing.
- Preserve the existing language, framework, database, package layout, build path, test runner, security model, and deployment model unless the user explicitly changes them.
- Modify only files required by the current request.
- Do not create parallel projects, placeholders, noop files, caches, runtime databases, compiled output, or unrelated artifacts.
- New files must match repository naming and ownership conventions.
- Root-level files require a clear repository-convention reason.

## Artifact naming

- Use names from the user request or existing repository convention.
- Preserve an existing file path when updating it unless the user explicitly requests a rename.
- Do not invent marketing names, random identifiers, machine names, or unnecessary timestamps.
- Report the exact final path.

## Permanent documentation rule

Permanent project documentation contains current project truth only.

Do not write any of the following into README, design, API, usage, security, testing, operations, or questions documents:

- agent mistakes or self-corrections;
- failed tool calls or unavailable tools;
- failed implementation attempts or experiment history;
- branch, commit, PR, or review chronology;
- conversation chronology;
- speculative recommendations or unrequested roadmap items;
- statements such as `next step`, `follow-up`, `remaining work`, `we can also`, or `could be improved` unless they are explicit tracked project requirements;
- claims copied from an earlier document without checking current code and requirements.

When an earlier claim is wrong, replace or remove it. Do not preserve the correction history as a project decision, limitation, rationale, or requirement.

Operational history belongs only in a temporary working record, final response, issue, PR, or acceptance report when that format explicitly requires execution history.

## Documentation targets

Use existing repository documentation targets. When the repository has no stricter convention:

- `README.md`: onboarding, setup, first useful run, usage, repository map, and maintenance commands.
- `docs/api-spec.md`: current API contract, examples, errors, authentication, and API test design.
- `docs/design.md`: current architecture, requirement implementation, domain rules, security boundaries, data flow, storage, configuration, and validation design.

`docs/questions.md` is optional. It is not generated automatically during project generation, repair, validation, testing, or review.

## Questions document hard contract

`docs/questions.md` is a current project clarification register. It is not:

- a FAQ or Q&A document;
- a transcript or discussion record;
- a decision log or design document;
- a record of agent reasoning;
- an error, incident, experiment, or correction log;
- a progress report;
- a TODO, roadmap, recommendation, next-action, or follow-up list.

### Eligibility

Create or update `docs/questions.md` only when all conditions are true:

1. a real ambiguity exists in an authoritative project source;
2. the user or repository convention requires a durable clarification register;
3. the current project rule is explicitly confirmed or directly established by an authoritative source;
4. the entry can be written without describing agent behavior, history, attempts, or alternatives.

An authoritative source is limited to:

- the original project prompt or specification;
- explicit user feedback or confirmation;
- an authoritative issue or requirement document;
- a committed repository rule that is not contradicted by a higher-priority source.

The following are not authoritative clarification sources:

- agent analysis or memory;
- generated reports;
- existing `docs/questions.md` content by itself;
- failed tests or failed commands;
- implementation mistakes;
- commit messages, branch names, PR descriptions, or review comments;
- assistant recommendations.

### Required format

Do not use numbered questions, question headings, or answer sections. Use only:

```markdown
# Project Clarifications

| Topic | Current project rule | Authoritative source |
|---|---|---|
| {{TOPIC}} | {{CURRENT_RULE}} | {{SOURCE}} |
```

Each row records one current rule. It must not describe how the rule was discovered or corrected.

Do not use these headings or labels:

```text
Q1
Question
Answer
Why
Reasoning
Decision
Correction
Mistake
Experiment
Attempt
What this resolves
Effect on project
Next step
Next action
Remaining work
Follow-up
Recommendation
```

### Update behavior

- Never append one entry per agent iteration.
- Rebuild the file from the canonical current clarification set.
- Deduplicate by topic.
- Remove superseded wording, rejected alternatives, old implementation status, and historical narration.
- Do not infer a rule from a failed implementation or from the way a test happened to be written.
- When a clarification becomes ordinary design, API, setup, or usage truth, move the current fact to the correct document and remove the redundant questions entry unless the user explicitly requires it to remain.
- When the answer is unknown, ask the user. Do not write the unanswered question or a guessed answer into the file.

When no qualifying clarification exists, do not create or modify `docs/questions.md`. When a repository rule requires the file to exist, its complete content must be:

```markdown
# Project Clarifications

No project clarifications are currently required.
```

### Mandatory pre-write gate

Before writing `docs/questions.md`, build this temporary table outside the repository:

```text
Candidate topic | Authoritative source | Current rule | Eligible? | Exclusion reason
```

Block the write if any candidate:

- has no authoritative source;
- originates from an agent mistake, experiment, tool failure, test failure, PR history, or review history;
- duplicates design, API, README, usage, or test documentation;
- is written as Q&A;
- contains future work, recommendations, progress, or chronology;
- contradicts current implementation or a higher-priority requirement.

After writing, scan the whole file. The task fails until all Q&A formatting, process residue, source-less rules, duplicate topics, stale statements, and contradictions are removed.

This contract overrides generic clarification-answer or generic documentation-template wording in any other loaded Skill.

## Target-specific document structure

Do not force one universal template onto every document.

- README uses an onboarding structure.
- API documentation uses an API contract structure.
- Design documentation uses a current design and requirement-mapping structure.
- Questions documentation uses only the hard contract above.
- Acceptance reports may include evidence, execution status, checks not run, and gaps.

Do not automatically add `Source inputs`, `Decisions and constraints`, `Checks not run`, `Pending items`, or `Conversation record` to permanent product documentation. Include such sections only when the target document's explicit purpose requires them.

## Temporary working record

For multi-turn work, keep user corrections, branch state, failed operations, successful operations, commands, and blockers in a temporary working record.

- The latest user feedback overrides earlier assumptions.
- Correct affected files directly when an earlier agent claim was wrong.
- Do not commit the working record unless the user explicitly requests a historical log.
- Do not use the working record as evidence for a project requirement.
- Ask one direct clarification question when required input is missing.

## Development workflow

1. Resolve the repository root and base branch.
2. Inspect relevant current files before planning.
3. Build a requirement ledger mapped to existing touchpoints.
4. Modify only required files.
5. Preserve package and dependency direction.
6. Add or update tests in the existing test layout.
7. Run repository-standard checks when available.
8. Review the final diff for unrelated changes.
9. Report exact changed paths, checks run, checks not run, and remaining evidence gaps.

Do not merge, force-push, reset, delete branches, publish releases, or perform other risky repository actions without explicit user approval.

## Code quality contract

- Use idiomatic naming and formatting for the selected language.
- Use existing error, response, logging, configuration, migration, and test conventions.
- Keep functions, types, modules, and packages focused and readable.
- Comments must explain intent, invariants, constraints, data contracts, error semantics, or non-obvious domain decisions.
- Comments must match implementation and project documentation.
- Do not add comments that restate obvious code or preserve implementation history.
- Public APIs and persisted data must use typed or schema-backed contracts appropriate to the language and framework.
- API projects must expose framework-appropriate OpenAPI or equivalent contract documentation.
- Do not hard-code production secrets, local absolute paths, or machine-specific assumptions.

## Logging contract

- Use the repository logger or an ecosystem-appropriate structured logger.
- Support reasonable levels and configurable output.
- Include useful operation, request, actor, resource, dependency, and duration context when available.
- Do not log credentials, keys, tokens, private file contents, or full sensitive request bodies.
- Log important lifecycle, authentication, permission, validation, workflow, dependency, database, and background-job failures.

## Evidence contract

Do not mark a requirement verified because code or a test file merely exists. Distinguish:

- implementation exists;
- test exists;
- test executed;
- local acceptance executed;
- Docker build executed;
- Docker/runtime acceptance executed;
- CI executed for the exact revision;
- logs, reports, screenshots, and artifacts were inspected.

Assistant-written summaries are not test evidence.

## Final response

Report:

- conclusion: `verified`, `partially_verified`, `implemented_but_evidence_missing`, or `not_fixed`;
- exact files changed;
- requirement-to-evidence summary;
- checks actually run;
- checks not run;
- unresolved blockers or missing user input;
- claims that must not yet be made.
