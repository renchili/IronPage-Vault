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
3. locate all implementations, tests, schemas, migrations, configuration, deployment, API, documentation, UI assets, and comments that express the same behavior;
4. fix every statically identifiable contradiction, omission, unsafe fallback, stale path, incomplete negative path, and documentation mismatch within scope;
5. update production code, test definitions, static guards, schemas, configuration, UI implementation, and documentation together where the requirement crosses those boundaries;
6. rescan the complete affected surface after changes;
7. continue after finding a blocker so the final result contains all independently identifiable issues rather than only the first one.

A narrow change is correct only when the requirement itself is narrow. “Minimal,” “smallest,” “first step,” “good enough for CI,” and “wait for CI” are not valid scope controls.

## Required inputs

- `{{PROJECT_PROMPT}}`: original prompt, issue text, uploaded metadata, README text, or equivalent requirement source.
- `{{PROJECT_NAME}}`: optional project name when supplied.
- `{{TARGET_REPO}}`: optional repository to create, modify, validate, or review.
- `{{REPO_ROOT}}`: repository root when working in a repository.
- `{{USER_GOAL}}`: generate, repair, validate, package, or review.
- `{{CONSTRAINTS}}`: technical, product, repository, UI platform, and non-goal constraints.
- `{{USER_FEEDBACK}}`: latest correction or requested change.

## Frontend design and implementation contract

Apply this section only when the requested project or changed scope includes a production frontend, application UI, embedded UI, mobile UI, desktop UI, extension UI, plugin UI, or an implementation-guiding prototype. A backend-only project or explicitly acceptance-only browser probe does not acquire a production frontend requirement merely because this Skill contains UI rules.

### Determine the actual application surface

Before proposing or generating UI, identify from the prompt and repository:

- target platform and form factor;
- host application or shell, when embedded;
- framework and pinned version;
- existing design system and component library;
- existing icon library;
- theme support;
- navigation and routing ownership;
- target viewport, window, device, or host-panel constraints;
- platform review, marketplace, plugin, accessibility, or submission rules that apply;
- existing screenshots, Figma files, Storybook stories, components, tokens, and interaction conventions.

Use current authoritative project or platform material when it is accessible. Do not guess a host application's review rules or silently substitute generic web conventions. When required platform material is inaccessible, identify the exact evidence gap rather than inventing a rule.

### Buildable frontend requirement

A UI result is complete only when a frontend engineer can implement it without independently deciding the product's visual or interaction design. The generated result must resolve every material decision that affects implementation, including:

- screen purpose, route, entry points, exit paths, navigation placement, deep-link behavior, and permission visibility;
- concrete component choice from the repository's actual component system;
- component hierarchy, ownership, reusable boundaries, inputs, outputs, and state ownership;
- exact icon library and icon name for every icon-bearing control;
- icon visual size, stroke or fill treatment, alignment, container size, hit target, label or tooltip, and disabled/selected treatment;
- page, panel, control, table, toolbar, dialog, drawer, and content-region dimensions or constraints;
- spacing, typography, color, border, radius, elevation, density, and theme values using real repository or host tokens;
- responsive breakpoints, minimum supported width, overflow, truncation, wrapping, scrolling, sticky regions, zoom, and long-content behavior;
- default, hover, focus, pressed, selected, disabled, loading, empty, success, validation-error, request-error, permission-denied, conflict, stale-data, destructive, and terminal/read-only states where applicable;
- user-visible copy, field labels, validation messages, recovery guidance, confirmation wording, and persistence of unsaved input;
- keyboard order, focus placement and return, accessible names, status announcements, contrast-independent meaning, and non-pointer operation;
- animation or motion only when required by the platform or interaction, including trigger, duration, easing source, reduced-motion behavior, and completion state.

Do not use vague placeholders such as “appropriate icon,” “standard spacing,” “normal modal,” “reasonable size,” “similar to the app,” or “handle errors.” Name the actual implementation choice or mark a real unresolved requirement.

### Special-interaction contract

For drag-and-drop, canvas, document viewers, redaction regions, cropping, diagramming, timeline editing, reordering, multi-selection, command palettes, gesture input, inline editing, virtualized data, or any other non-trivial interaction, define:

- supported input methods;
- start condition and activation threshold;
- behavior while the interaction is active;
- coordinate system, snapping, bounds, scrolling, zoom, and selection behavior;
- commit condition and persisted result;
- cancel, Escape, pointer-cancel, route-change, and lost-focus behavior;
- undo, redo, delete, and recovery behavior where applicable;
- duplicate-submit and re-entrancy prevention;
- optimistic or pessimistic update choice;
- loading, partial failure, network failure, stale version, conflict, permission loss, and retry behavior;
- keyboard-only and assistive-technology path;
- the source state and API or domain mutation affected by completion.

A picture of the final state is not a sufficient specification for a special interaction.

### Platform and app-review compliance

When the UI belongs to a host application, operating system, marketplace, plugin ecosystem, or device platform, the generated design and code must follow that target's actual review format and interaction conventions. Inspect and map the applicable requirements for navigation, icons, sizing, theme, dialogs, destructive actions, permissions, accessibility, branding, privacy, restricted APIs, screenshots, metadata, and submission artifacts.

Do not claim platform compliance from visual resemblance. Each applicable review rule must map to a screen, component, configuration path, implementation path, or explicit `N/A` reason.

### Artifact-format boundary

Use the UI artifact format requested by the user or already established by the repository, such as application source, Figma, Storybook, HTML/CSS/JS, native layout files, design-system components, screenshots with measurements, or an existing design document.

Do not invent YAML, JSON, schema, manifest, token-registry, or “review pack” deliverables merely to make the result look structured. Create such files only when the user requests them, the target platform requires them, a loaded Skill requires them, or the repository already uses that exact convention.

Do not replace an interactive prototype with a prose specification when the user asked for a prototype. Do not replace implementation with mockups when the user asked to build the frontend. When the request is design-only, the prototype and annotations must still resolve the implementation-critical decisions above.

### UI traceability and implementation readiness

For every material screen and interaction, trace:

```text
Requirement | Screen/route | Component | Visual specification | Interaction/state path | Data/API dependency | Permission | Test/static evidence
```

Before delivery, answer from the generated artifacts and source:

1. Can an engineer identify the exact component, icon, dimensions, spacing, state ownership, and platform pattern without redesigning the feature?
2. Can an engineer determine what happens after every supported action, cancellation, failure, denial, conflict, and retry?
3. Can a reviewer or tester derive concrete visual, interaction, accessibility, and platform-review assertions?

If any answer is no for a required flow, the frontend portion is incomplete. If frontend code is in scope, implement the resolved design in production source and tests; do not stop at the handoff document.

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
- `docs/design.md`: architecture, requirement implementation, module boundaries, data flow, security, workflow, storage, constraints, UI decisions when applicable, and static validation strategy.
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

`Acceptance evidence` under this Skill means static proof paths and acceptance conditions: production implementation, negative-path logic, test definitions, schemas, configuration, documentation, UI implementation when applicable, and expected state or error semantics. It must not require the agent to execute tests, CI, a build, a container, a database, a browser, or deployment.

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

Do not replace the repository language, framework, database, architecture, build path, security model, deployment model, design system, or host application conventions unless explicitly requested.

## Static development workflow

For repository changes:

1. resolve the base revision and current branch state;
2. build an atomic requirement ledger;
3. map every requirement to all affected source, tests, schema, config, deployment, API, UI, docs, and comments;
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

## Code, schema, UI, and annotation contract

- Preserve package boundaries and dependency direction.
- Follow existing error, response, logging, configuration, migration, design-system, component, icon, and test conventions.
- Add or update tests in the existing layout for every changed behavior, including negative paths and required UI states.
- Do not hard-code production secrets, deployment-specific values, local absolute paths, ports, hostnames, identities, or machine assumptions.
- Comments must explain intent, invariants, security boundaries, state transitions, storage contracts, special-interaction behavior, and non-obvious failure handling.
- Exported APIs and public handlers must use language/framework documentation conventions.
- API schemas, request/response models, error models, route definitions, UI data dependencies, and generated-contract source must agree statically.

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
3. current production source, schema, migrations, configuration, manifests, deployment files, UI assets, and API contracts;
4. current test definitions and static guards;
5. current documentation and comments;
6. pre-existing external logs or artifacts, read-only and optional.

A missing external artifact does not demote a complete static implementation. Conversely, a green external artifact does not excuse a static contradiction.

## Algorithm

1. Read the requirement and rule sources.
2. Resolve repository root, base revision, branch, and complete affected surface.
3. Build an atomic requirement ledger.
4. Inspect every mapped implementation and documentation path.
5. Generate or repair all statically identifiable requirements, including implementation-ready UI decisions when UI is in scope.
6. Add or update complete positive, negative, boundary, failure-path, interaction-state, and accessibility test definitions.
7. Update schemas, configuration, deployment, comments, UI assets, and docs consistently.
8. Perform a complete static rescan and diff review.
9. Continue until no known in-scope static defect is deferred.
10. Report the static conclusion without waiting for CI, tests, builds, containers, browsers, or deployment.

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
