# Prompt-Driven Project Generation Workflow

Use this Skill to generate, repair, validate, package, or review a software project from user-provided requirements. Repository work must preserve the existing architecture, project structure, documentation purposes, code quality, and agent capability boundaries.

## Core static-generation rule

Project generation and repair are **static source-completion tasks**.

The agent must complete the strongest implementation that can be derived from the requirements and repository contents without depending on any external execution result.

During generation, repair, validation, or review under this Skill, the agent must not:

- run project code;
- run unit, integration, API, browser, regression, or acceptance tests;
- execute repository scripts or generated binaries;
- build images, packages, applications, or containers;
- start services, databases, browsers, emulators, or deployment environments;
- trigger, rerun, retry, dispatch, or wait for CI;
- treat CI, a build, a deployment, a test run, or another external validator as a prerequisite for continuing work;
- stop after a minimum patch, first defect, first stage, first commit, or first green external signal;
- defer statically identifiable work to a later step, follow-up PR, future pass, or CI result.

Existing logs, CI results, test reports, screenshots, artifacts, or deployment records may be inspected read-only when they already exist. They are optional context, not generation dependencies and not permission to stop before the static implementation is complete.

Missing external execution evidence must never produce `ci_pending`, `waiting_for_tests`, `waiting_for_build`, or an equivalent stop state. Continue the static work and describe runtime execution as outside this workflow.

## Complete-delivery rule

The agent must not optimize for the smallest change count, fewest files, shortest plan, or minimum viable patch. Scope is determined by the complete user requirement and repository constraints, not by convenience.

Before delivery:

1. reconstruct the request into atomic requirements;
2. inspect every repository area materially affected by those requirements;
3. locate all implementations, tests, schemas, migrations, configuration, deployment, API, documentation, and comments that express the same behavior;
4. fix every statically identifiable contradiction, omission, unsafe fallback, stale path, incomplete negative path, and documentation mismatch within scope;
5. update production code, test definitions, static guards, schemas, configuration, and documentation together where the requirement crosses those boundaries;
6. rescan the complete affected surface after changes;
7. continue after finding a blocker so the final result contains all independently identifiable issues rather than only the first one.

A narrow change is correct only when the requirement itself is narrow. “Minimal,” “smallest,” “first step,” “good enough for CI,” and “wait for CI” are not valid scope controls.

## Required inputs

- `{{PROJECT_PROMPT}}`: original prompt, issue text, uploaded metadata, README text, or equivalent requirement source.
- `{{PROJECT_NAME}}`: optional project name when supplied.
- `{{TARGET_REPO}}`: optional repository to create, modify, validate, or review.
- `{{REPO_ROOT}}`: repository root when working in a repository.
- `{{USER_GOAL}}`: generate, repair, validate, package, or review.
- `{{CONSTRAINTS}}`: technical, product, repository, and non-goal constraints.
- `{{USER_FEEDBACK}}`: latest correction or requested change.

## Repository hygiene rules

When a repository is involved:

- Treat the resolved repository root as the only project root.
- Inspect the file tree, current branch, changed files, source, tests, scripts, migrations, CI, Docker/deployment files, docs, generated artifacts, and ignore rules before editing.
- Do not create parallel projects, sample apps, placeholder files, noop files, duplicate roots, or unrelated outputs.
- Do not place source code outside the existing project root.
- New files must fit existing package, directory, naming, and ownership conventions.
- Root-level files require a repository-convention reason.
- Exclude runtime databases, caches, compiled output, logs, temporary files, secrets, generated reports, and unrelated artifacts from delivery.
- Inspect all path names for portability, case-only collisions, accidental spaces, control characters, locale-dependent names, and ambiguous near-duplicates.

## Generated artifact naming rules

Generated documents and artifacts must use names derived from the request or existing project conventions.

- Prefer the user-provided title, project name, repository name, existing path, task identifier, PR/issue number, or established artifact naming pattern.
- Do not invent marketing names, random identifiers, local machine names, or timestamps unless explicitly requested.
- Keep names portable and repository-conventional.
- Preserve an existing file name and path when updating it unless the user requests a rename.
- Report final paths exactly as written.

## Documentation file rules

Documentation is project source and must be checked statically against the implementation.

- Update an existing documentation home before creating a new document.
- Create a document only when requested, required by this Skill, or established by repository convention.
- Do not use documentation to store agent chronology, tool failures, repeated experiments, branch history, assistant self-correction, or untracked future work.
- Documentation may claim static implementation only when the cited source paths support it.
- Documentation must not claim an executed test, build, deployment, CI run, browser flow, or runtime result unless that pre-existing artifact was actually inspected.
- Lack of an external run does not block completing the documentation or implementation.

## Default documentation outputs

When the repository has no stricter convention:

- `docs/api-spec.md`: API contract, auth, methods, request/response fields, errors, examples, pagination, and static acceptance mappings.
- `docs/design.md`: architecture, requirement implementation, module boundaries, data flow, security, workflow, storage, constraints, and static validation strategy.
- `docs/questions.md`: durable clarification of requirements that were easy to misinterpret or previously implemented inconsistently.

Do not merge these purposes unless explicitly requested.

## Questions document clarification contract

`docs/questions.md` is organized by requirement topic, not Q&A numbering or conversation turns.

A topic belongs when:

- the requirement has a plausible but incorrect or incomplete interpretation;
- an implementation can appear functional while violating a repository or acceptance rule;
- static inspection of code, tests, configuration, or documentation exposes an incomplete boundary;
- user feedback corrects the interpretation;
- repeated work yields one stable project-level conclusion.

Each topic must contain exactly these semantic parts:

1. `Easy-to-make interpretation`
2. `Why it fails`
3. `Correct requirement interpretation`
4. `Required implementation`
5. `Acceptance evidence`

`Acceptance evidence` under this Skill means static proof paths and acceptance conditions: production implementation, negative-path logic, test definitions, schemas, configuration, documentation, and expected state or error semantics. It must not require the agent to execute tests, CI, a build, a container, a database, a browser, or deployment.

Example:

```markdown
## Administrator initialization credentials

### Easy-to-make interpretation

A default administrator can be implemented with one fixed username and password in product code or normal deployment configuration.

### Why it fails

That embeds installation-specific identity and credentials in the product and makes separate installations share a long-lived secret.

### Correct requirement interpretation

A clean installation must initialize an administrator without embedding a fixed deployment credential and must preserve existing identities on restart.

### Required implementation

Detect an empty installation, use deployment-supplied or installation-generated bootstrap values, and never overwrite an existing administrator.

### Acceptance evidence

The initialization path is empty-installation-only; restart paths preserve existing users; normal source contains no fixed deployment credential; tests define first-install, separate-installation, and restart cases.
```

Before delivery, reject a questions document containing:

- numbered Q&A structure;
- agent or PR chronology;
- repeated or conflicting topics;
- an interpretation without a requirement or repository basis;
- generic next-step or recommendation sections;
- implementation guidance without static acceptance conditions;
- claims that completion depends on a future CI or test result.

## Working record rules

Maintain a current working record for multi-turn work containing:

- target repository, base branch, and working branch;
- loaded rule paths and stable identifiers;
- atomic requirement ledger;
- complete affected-file map;
- user corrections;
- changes completed;
- static checks performed by inspection;
- unsupported runtime claims explicitly excluded from the static conclusion.

The working record is not permanent project documentation.

The latest user feedback overrides earlier assumptions. Do not ask the user to reconfirm conclusions derivable from the prompt, repository rules, current source, or prior explicit feedback.

## Repository constraint rules

Before changing repository content, read:

- `AGENTS.md` and `AGENT.md` when present;
- relevant `skills/**/SKILL.md`;
- README and relevant docs;
- source layout, tests, scripts, migrations, workflows, deployment files, configuration, and artifact conventions.

Do not replace the repository language, framework, database, architecture, build path, security model, or deployment model unless explicitly requested.

## Static development workflow

For repository changes:

1. resolve the base revision and current branch state;
2. build an atomic requirement ledger;
3. map every requirement to all affected source, tests, schema, config, deployment, API, docs, and comments;
4. edit the complete mapped surface, not only the first convenient file;
5. inspect syntax and consistency from source text without executing repository code or scripts;
6. review the complete diff for unrelated files and unresolved requirement rows;
7. repeat static inspection until every row is `STATIC PASS`, `STATIC FAIL`, or justified `N/A`;
8. deliver without waiting for CI or external execution.

Do not create artificial stages whose purpose is to postpone known work. A plan may organize reasoning, but it cannot reduce the required implementation or create an external waiting point.

## Submission rules

When the user asks to submit code:

- use a purpose-specific branch;
- commit the complete relevant change set;
- exclude unrelated cleanup and generated state;
- open a PR only when requested or clearly required;
- do not wait for or require PR CI before reporting the completed static change;
- do not merge, force-push, reset, delete branches, or publish releases without explicit approval.

PR text must distinguish static implementation from external runtime evidence and must not say work is pending merely because CI has not run.

## Agent operation rules

- State repository reads and writes actually performed.
- Do not claim external execution occurred.
- Do not present `ci_pending`, missing test execution, or missing deployment as a reason to stop static generation.
- Do not provide “run this and return the result” as a substitute for completing statically identifiable work.
- Do not reduce the task to a minimal patch unless the user explicitly narrows the scope.
- Do not stop scanning after the first P0 or first failed requirement.
- Keep branches and commits purpose-specific, but completeness outranks minimizing file count.
- Ask only before risky publishing or destructive repository actions.

## Code, schema, and annotation contract

- Preserve package boundaries and dependency direction.
- Follow existing error, response, logging, configuration, migration, and test conventions.
- Add or update tests in the existing layout for every changed behavior, including negative paths.
- Do not hard-code production secrets, deployment-specific values, local absolute paths, ports, hostnames, identities, or machine assumptions.
- Comments must explain intent, invariants, security boundaries, state transitions, storage contracts, and non-obvious failure handling.
- Exported APIs and public handlers must use language/framework documentation conventions.
- API schemas, request/response models, error models, route definitions, and generated-contract source must agree statically.

## Logging contract

- Use the repository logger rather than ad-hoc prints.
- Support appropriate levels and configurable format/destination.
- Include request, actor, resource, operation, dependency, and duration context when available.
- Never log credentials, secrets, protected document contents, or complete sensitive payloads.
- Cover startup, shutdown, auth failure, denial, validation failure, state transition, dependency failure, database failure, and background work.
- Keep logging documentation aligned with code.

## Static evidence model

Use this evidence order for generation decisions:

1. original requirement and explicit user corrections;
2. repository rule files;
3. current production source, schema, migrations, configuration, manifests, and deployment files;
4. current test definitions and static guards;
5. current documentation and comments;
6. pre-existing external logs or artifacts, read-only and optional.

A missing external artifact does not demote a complete static implementation. Conversely, a green external artifact does not excuse a static contradiction.

## Algorithm

1. Read the requirement and rule sources.
2. Resolve repository root, base revision, branch, and complete affected surface.
3. Build an atomic requirement ledger.
4. Inspect every mapped implementation and documentation path.
5. Generate or repair all statically identifiable requirements.
6. Add or update complete positive, negative, boundary, and failure-path test definitions.
7. Update schemas, configuration, deployment, comments, and docs consistently.
8. Perform a complete static rescan and diff review.
9. Continue until no known in-scope static defect is deferred.
10. Report the static conclusion without waiting for CI, tests, builds, containers, or deployment.

## Final response contract

Use one conclusion:

- `STATIC PASS`: all in-scope requirements are implemented and consistent by static inspection.
- `STATIC FAIL`: one or more in-scope static defects remain; enumerate all known defects.
- `STATIC INCOMPLETE`: required source or rule material was inaccessible; identify the exact missing path. Do not use this status merely because CI or runtime evidence is absent.

The final response must include:

- exact branch and base revision;
- exact files changed;
- loaded rule paths and identifiers;
- complete static requirement summary;
- repository writes performed;
- external execution performed: `none`;
- CI triggered or awaited: `none`;
- every remaining static defect or inaccessible source.
