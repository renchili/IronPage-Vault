# Prompt-Driven Project Generation Workflow

Use this skill to generate, repair, validate, package, or review a software project from user-provided input. When a target repository is involved, preserve repository hygiene, project structure, agent capability boundaries, and code quality.

## Required inputs

- `{{PROJECT_PROMPT}}`: original prompt, issue text, uploaded metadata, README text, or equivalent requirement source.
- `{{PROJECT_NAME}}`: optional project name when supplied.
- `{{TARGET_REPO}}`: optional repository to create, modify, validate, or review.
- `{{REPO_ROOT}}`: required when working in a repository; repository root for reading, writing, testing, and packaging.
- `{{USER_GOAL}}`: generate, repair, validate, package, or review.
- `{{CONSTRAINTS}}`: optional technical and non-goal constraints.
- `{{USER_FEEDBACK}}`: optional latest correction or requested change.

## Repository hygiene rules

When `{{TARGET_REPO}}` or `{{REPO_ROOT}}` is present, repository context is mandatory.

- Treat `{{REPO_ROOT}}` as the only project root.
- Inspect the file tree, current branch, changed files, tests, scripts, migrations, CI, Docker/deployment files, docs, and generated artifacts before planning.
- Do not create parallel projects, sample apps, placeholder files, noop files, or unrelated generated outputs.
- Do not place source code outside `{{REPO_ROOT}}`.
- New files must fit existing package, directory, naming, and ownership conventions.
- Root-level files require a clear repository-convention reason.
- Exclude accidental files, runtime databases, caches, compiled output, temporary files, and unrelated artifacts from delivery.

## Generated artifact naming rules

Generated documents and artifacts must use names derived from the request or from existing project conventions.

- Do not invent arbitrary names for generated documents, reports, exports, ZIP files, PDFs, DOCX files, spreadsheets, slides, screenshots, or evidence bundles.
- Prefer the user-provided title, project name, repository name, existing document name, task identifier, PR/issue number, or established repository artifact naming pattern.
- If no stable name is available, use a plain descriptive fallback based on the exact deliverable type, such as `project-report.md`, `acceptance-evidence.zip`, or `api-summary.docx`.
- Keep generated file names lowercase or repository-conventional, portable across filesystems, and free of local machine names, personal guesses, timestamps, random IDs, or marketing-style names unless explicitly requested.
- When replacing or updating an existing document, preserve its file name and path unless the user explicitly requests a rename.
- Report the final artifact path or file name exactly; do not describe a generated artifact with a different display name from the actual file.

## Documentation file rules

Documentation files are project artifacts and must follow repository purpose, naming, and evidence rules.

- Update an existing documentation file before creating a new one when the topic already has a home.
- Create a new documentation file only when the user asks for it, the task requires a durable record, or the repository has a matching docs convention.
- Place documentation under the repository's existing docs path, artifact path, or requested path; do not create loose root documents without a repository-convention reason.
- Documentation file names must follow the generated artifact naming rules and repository naming style.
- Project documentation must distinguish requirements, implementation notes, validation evidence, checks not run, and pending items.
- Do not use documentation to claim completion that is not backed by code, tests, CI, logs, reports, or artifacts.

## Default documentation outputs

For project generation or repair work, use these fixed documentation targets when the repository does not define a stricter documentation convention:

- `docs/api-spec.md`: API usage specification. It documents endpoints, methods, auth model, request fields, response fields, error behavior, examples, command examples, and API acceptance checks.
- `docs/design.md`: project design and requirement implementation record. It explains the whole project design, how requirements are implemented, architecture, modules, data flow, security boundaries, workflow rules, storage model, constraints, and validation strategy.
- `docs/questions.md`: clarification answer record. It answers unclear process, acceptance, testing, runtime, delivery, usage, and verification points. It explains what the unclear point means in this project, why the answer is reasonable, and how the answer follows from user feedback, requirements, repository constraints, existing implementation, and validation goals.

Do not merge these three document purposes into one file unless the user explicitly asks for a single document. If one of these files already exists, update it in place. If a required section has no content yet, write `None currently known` rather than inventing content. Do not record agent execution failures, tool failures, PR creation failures, or platform capability limits in `docs/questions.md`; those belong in the final response or working record, not in project clarification documentation.

## Questions document clarification-answer contract

`docs/questions.md` is not a FAQ, not a question-and-answer transcript, not a discussion record, not a decision log, not an agent error log, and not a generic TODO list. It is a project clarification answer record.

A record belongs in `docs/questions.md` only when it explains a project-relevant unclear point, such as:

- how a process should work.
- how acceptance should be judged.
- how tests should prove the requirement.
- how local, Docker, CI, or manual verification should be interpreted.
- how an API behavior, error behavior, permission rule, workflow rule, data rule, or runtime rule should be understood.
- how user feedback changes the correct interpretation of a requirement.
- how an implementation detail should be understood when the repository or prompt leaves it ambiguous.

Each entry must be written as an explanation, not as a question. Use this structure:

1. `Clarification area`: the process, acceptance point, test point, API behavior, workflow, runtime behavior, or delivery concern being clarified.
2. `Unclear point`: what was unclear or previously easy to misinterpret.
3. `Clarification`: the project-specific answer.
4. `Reasoning`: why this answer is reasonable based on user feedback, prompt requirements, repository constraints, existing implementation, security model, or validation goals.
5. `What this resolves`: what ambiguity, incorrect implementation path, weak test, or weak acceptance claim this clarification prevents.
6. `Effect on project`: what code behavior, documentation section, test case, acceptance evidence, or delivery claim this clarification affects.

Rules:

- Do not write entries as questions.
- Do not write entries as yes/no items.
- Do not include `Next action` sections.
- Do not include agent tool failures, blocked tool calls, PR creation failures, or platform limitations.
- Do not copy the full chat history; summarize only the project-relevant clarification trail.
- If the answer cannot be derived from available context, do not invent it in `docs/questions.md`; ask the user directly and wait.

## Documentation output contract

When writing or updating a documentation file, the agent must produce a reviewable project record, not a loose summary.

Before writing the document:

1. Resolve the exact target path from the user request, existing documentation home, repository convention, or generated artifact naming rules.
2. If the target path cannot be resolved, ask the user for the path or document name before writing.
3. State whether the task updates an existing document or creates a new one.

A project documentation file must use this structure unless the repository already has a stricter template:

1. `# {{TITLE}}`
2. `## Purpose`
3. `## Source inputs`
4. `## Requirement ledger`
5. `## Decisions and constraints`
6. `## Implementation notes`
7. `## Validation evidence`
8. `## Checks not run`
9. `## Pending items`
10. `## Conversation record`

Section rules:

- `Source inputs` must list the prompt, issue, PR, existing file, uploaded file, or user message source used for the document.
- `Requirement ledger` must map each requirement to status and evidence.
- `Validation evidence` must cite concrete tests, CI, logs, generated artifacts, screenshots, reports, or file paths.
- `Checks not run` must be explicit when local tests, Docker acceptance, CI, or manual checks were not executed.
- `Pending items` must carry unresolved user feedback and missing evidence forward.
- `Conversation record` must record only operational facts: user corrections, decisions, commands given to the user, merged PRs, failed operations, successful operations, and remaining blockers.

Final response after writing a document must include the exact document path, whether it was created or updated, checks run, checks not run, and pending items.

## Conversation record rules

For multi-turn work, maintain a current working record before changing files or reporting completion.

- Track user corrections, confirmed decisions, merged PRs, branch state, failed operations, successful operations, user-provided commands, and pending items.
- Treat the latest user feedback as a constraint that overrides earlier assumptions.
- Carry unresolved feedback forward until it is fixed, explicitly declined by the user, or marked pending with a reason.
- If a required step must be executed by the user locally, provide exact commands and wait for the result before claiming the step is complete.
- If correct implementation needs missing user input, ask one direct question for that input and wait before changing files or artifacts.
- Do not offer alternate edits or scope changes for that feedback while the required user input is missing.

## Repository constraint rules

Before generating code for a repository, read repository constraints from existing files and structure:

- `AGENT.md`, when present.
- `README.md`, when present.
- relevant files under `docs/`, when present.
- existing source layout, tests, scripts, migrations, CI workflows, Docker/deployment files, and artifact conventions.

Do not replace the repository language, framework, database, build path, test runner, project layout, security model, or deployment model unless the user explicitly asks to change that direction.

## Development workflow rules

For code changes inside a repository:

1. Identify the base branch and current dirty state before writing.
2. Summarize the intended file-level change set before large edits.
3. Modify only files required by the current requirement.
4. Keep generated code inside the existing project tree and package layout.
5. Run repository-standard checks when available.
6. Review the final diff for unrelated files before committing or proposing a PR.
7. Report exact files changed, checks run, checks not run, and remaining risks.

If the user asks the agent to submit code:

1. Use or create a purpose-specific branch from the current base branch.
2. Commit only the relevant project-compliant changes.
3. Do not include unrelated cleanup, placeholder files, generated caches, local runtime state, or accidental files.
4. Open a PR only when requested or clearly required by the task.
5. PR body must include summary, changed files, validation, not-run checks, and known gaps.
6. Do not merge, force-push, reset, delete branches, or publish releases without explicit user approval.

## Commit message rules

Use concise, reviewable commit messages:

- Format: `<type>: <imperative summary>` or `<type>(<scope>): <imperative summary>`.
- Allowed types include `feat`, `fix`, `docs`, `test`, `ci`, `refactor`, `chore`, and `skill`.
- Summary should be specific and normally under 72 characters.
- Use body lines when needed: `Why`, `What changed`, and `Validation`.
- Do not use placeholder messages such as `noop`, `update`, `changes`, or `fix stuff`.

## Agent operation rules

- State which operations were actually executed and which were not.
- Do not claim tests, builds, CI, container runs, deployment, commits, or PR changes succeeded without tool evidence.
- If an environment dependency is unavailable, mark the item as `not_executed` or `ci_pending` and provide project-integrated commands or scripts.
- Do not make repository writes that contain unrelated files, placeholder files, or cleanup noise.
- Keep branches and commits reviewable and purpose-specific.
- Ask before risky repository actions that rewrite or publish work.
- When user feedback requires missing input to fix correctly, ask one direct clarification question and wait before changing files or artifacts.
- Do not propose alternate edits or continue with a different scope while that required input is missing.

## Code annotation and schema contract

Code comments and API schemas are part of the deliverable, not optional cleanup.

- Comments must explain intent, invariants, constraints, data contracts, error semantics, or non-obvious domain decisions.
- Do not add comments that only restate what the next line of code already says.
- Exported APIs, public handlers, domain rules, migrations, security-sensitive code, workflow transitions, and complex validation logic must have useful comments or annotations.
- Go exported identifiers must follow Go documentation comment conventions, starting with the identifier name where appropriate.
- Go HTTP API handlers must use swaggo-compatible annotations when the project exposes HTTP APIs.
- Go swaggo annotations should include `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`, and `@Security` when applicable.
- Go request and response structs used by HTTP APIs must include JSON tags and field comments suitable for generated API documentation.
- Python request, response, configuration, and persisted data contracts must use Pydantic models instead of raw untyped dictionaries when validation or external interface shape matters.
- Python API projects must expose an OpenAPI schema and a Swagger-like documentation UI using the convention appropriate for the chosen framework.
- For FastAPI, use its native OpenAPI generation and keep `/docs` or an equivalent Swagger UI available.
- For Flask, use a compatible OpenAPI stack such as flask-smorest, flask-restx, apispec, or flasgger when the project does not already define a stricter convention.
- For Django REST Framework, use a compatible OpenAPI stack such as drf-spectacular or drf-yasg when the project does not already define a stricter convention.
- For other Python API frameworks, select the framework-compatible OpenAPI documentation approach instead of forcing FastAPI-specific assumptions.
- Python API documentation must connect the Pydantic models, route definitions, response models, error models, examples, and generated OpenAPI schema.
- If a repository already has a stricter OpenAPI, swaggo, Pydantic, or schema generation convention, follow the existing convention and preserve compatibility.

## Logging contract

Service logging is part of the deliverable. Logs must be simple to view, reasonable by default, and configurable for local, container, and production use.

- Do not use ad-hoc print statements as the primary logging system.
- Use the repository's existing logger when present.
- Go projects should use `slog`, `zap`, `zerolog`, or the existing project logger.
- Python projects should use `logging`, `structlog`, `loguru`, or the existing project logger.
- Logs must support levels such as debug, info, warn, and error.
- Log format must be configurable, with human-readable output for local development and structured JSON output for containers or production when appropriate.
- Log destination must be configurable, with stdout or stderr as the default for container-friendly operation.
- Runtime configuration should control log level, format, and destination through config files, flags, or environment variables.
- Logs must include useful context such as operation, request ID, actor ID, resource ID, component, dependency name, and duration when available.
- Logs must avoid private values, credentials, full request bodies, file contents, and other sensitive data.
- Important lifecycle and workflow events should be logged: startup, shutdown, configuration load, request entry, auth failure, permission denial, validation failure, state transition, external dependency failure, database failure, background job start, background job completion, and background job failure.
- Error logs must include enough context to debug the failure without exposing sensitive data.
- `docs/design.md` must describe the logging strategy when logging behavior is added or changed.
- `docs/api-spec.md` must document request ID, trace ID, or correlation header behavior when the API exposes or accepts those fields.

## Code generation standards

- Preserve existing package boundaries and dependency direction.
- Use existing error handling, response, logging, configuration, migration, and test conventions.
- Add tests in the existing test layout for changed behavior.
- Do not hard-code production secrets, local absolute paths, or machine-specific assumptions.
- Use portable script execution such as `bash run_tests.sh` or `bash scripts/name.sh`.
- Add comments for exported APIs, security-sensitive logic, workflow rules, non-obvious domain decisions, SQL migrations, and complex error handling.
- Avoid comments that merely restate obvious code.

## Working state

- `{{REQUIREMENT_LEDGER}}`: requirements from prompt plus project and repository constraints.
- `{{DELIVERY_PLAN}}`: plan mapped to existing project touchpoints.
- `{{CHANGE_SET}}`: files changed in the project-compliant plan.
- `{{EVIDENCE_MAP}}`: proof from code, tests, CI, logs, reports, and artifacts.

## Algorithm

1. Read `{{PROJECT_PROMPT}}`.
2. Resolve `{{PROJECT_NAME}}`, `{{TARGET_REPO}}`, and `{{REPO_ROOT}}` when supplied.
3. If a repository is involved, inspect project structure, constraints, dirty state, tests, scripts, CI, deployment files, migrations, docs, and artifacts.
4. Build `{{REQUIREMENT_LEDGER}}` with source paths and existing project touchpoints.
5. Build `{{DELIVERY_PLAN}}` that respects project boundaries and architecture.
6. Modify or generate only files that fit the existing project structure.
7. Add tests using the existing test layout.
8. Run available checks through project-standard commands.
9. Compare evidence back to the ledger and mark anything not executed honestly.
10. Apply `{{USER_FEEDBACK}}` as corrected ledger items and repeat until verified or explicitly pending.

## Evidence rules

Do not mark a requirement as `verified` only because code exists. Distinguish code existence, test existence, test execution, local acceptance, Docker build, Docker acceptance, CI for the exact commit, logs, reports, and full acceptance execution.

## Final response contract

Conclusion: `<verified | partially_verified | implemented_but_evidence_missing | not_fixed>`

Ledger summary:
1. `{{REQUIREMENT}}`: `{{STATUS}}`. Touchpoint: `{{EXISTING_TOUCHPOINT}}`. Evidence: `{{EVIDENCE}}`.
2. `{{REQUIREMENT}}`: `{{STATUS}}`. Touchpoint: `{{EXISTING_TOUCHPOINT}}`. Evidence: `{{EVIDENCE}}`.
3. `{{REQUIREMENT}}`: `{{STATUS}}`. Touchpoint: `{{EXISTING_TOUCHPOINT}}`. Evidence: `{{EVIDENCE}}`.

Still pending:
- `{{PENDING_ITEM}}`

Do not claim yet:
- `{{UNPROVEN_CLAIM}}`
