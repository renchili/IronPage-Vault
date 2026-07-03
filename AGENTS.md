# IronPage Vault Agent Bootstrap Guardrails

These instructions are repository-level guardrails for any agent working in this repository. They exist because project-specific Skill instructions may not be loaded automatically and may be lost after context compaction.

## Mandatory bootstrap

Before planning, editing, generating files, reviewing, or reporting completion, the agent must read:

1. `.chatgpt/skills/ironpage-production-workflow/SKILL.md`
2. `README.md`, when present
3. existing files under `docs/`, when present
4. existing source layout, tests, scripts, CI, Docker/deployment files, migrations, and configuration files

If the Skill file cannot be read, stop and ask the user for instructions. Do not continue with guessed project rules.

## Context survival after compaction

After any context compaction, model switch, long pause, or new continuation, the agent must re-read this file and the Skill file before continuing repository work.

The agent must preserve a compact working record in user-visible updates or PR notes when work spans multiple turns. The record must include:

- current branch and base branch
- files changed
- user corrections that changed the requirement
- checks run
- checks not run
- open risks or evidence gaps
- user-local commands that were given and whether results were received

Do not rely on memory alone after compaction.

## Repository write rules

- Do not write directly to `main` unless the user explicitly asks.
- Use a purpose-specific branch for every change.
- Do not create placeholder files, noop files, dummy files, sample projects, or unrelated artifacts.
- Modify only files required by the current requirement.
- Report exact file paths changed.
- Do not claim tests, builds, CI, Docker runs, or acceptance checks passed without evidence.

## Documentation output contract

For project generation or repair work, use these documentation targets unless a stricter repository convention is already present:

- `docs/api-spec.md`: API usage specification.
- `docs/design.md`: project design and requirement implementation explanation.
- `docs/questions.md`: clarification answer record for unclear process, acceptance, testing, runtime, delivery, usage, and verification points.

Do not merge those document purposes into one loose summary.

`docs/questions.md` is not a FAQ, not a question-and-answer transcript, not a discussion record, not a decision log, not an agent execution log, and not a TODO list. Entries must explain unclear project-relevant points and their impact on implementation or verification.

Agent execution problems such as tool failures, blocked PR creation, platform limits, or failed agent actions belong in the final response or PR notes, not in `docs/questions.md`.

## Python project generation contract

Python projects must be generated as a small, reviewable, runnable project, not as scattered scripts.

Before writing Python code, the agent must identify and state the project type:

- API service
- CLI tool
- background worker
- library/package
- data-processing job
- test/demo utility

If the project type cannot be determined from the user request or repository context, ask one direct clarification question before writing files.

### Python layout

Use the existing repository layout when present. For a new Python project, prefer a simple layout such as:

```text
src/<package_name>/
tests/
pyproject.toml
README.md
docs/api-spec.md
docs/design.md
docs/questions.md
```

Do not create multiple competing entry points or framework variants.

### Python API requirements

For Python API projects:

- Use typed request, response, configuration, and persisted data contracts.
- Use Pydantic models when validation or external interface shape matters.
- Expose an OpenAPI schema and a Swagger-like documentation UI using the selected framework's convention.
- For FastAPI, use native OpenAPI and keep `/docs` or an equivalent Swagger UI available.
- For Flask, use a compatible OpenAPI stack such as flask-smorest, flask-restx, apispec, or flasgger unless the repo already has a stricter convention.
- For Django REST Framework, use drf-spectacular or drf-yasg unless the repo already has a stricter convention.
- For other Python API frameworks, use that framework's compatible OpenAPI documentation approach. Do not force FastAPI assumptions onto another framework.
- Document endpoint behavior in `docs/api-spec.md`.

### Python configuration and logging

Python services must have simple, visible, configurable logging:

- Do not use ad-hoc `print()` as the primary logging system.
- Use `logging`, `structlog`, `loguru`, or the repository's existing logger.
- Support log levels such as debug, info, warning, and error.
- Support configurable log format, preferably human-readable locally and JSON for container or production use when appropriate.
- Default to stdout or stderr for container-friendly operation.
- Control log level and format through configuration or environment variables.
- Do not log secrets, tokens, private values, full request bodies, file contents, or sensitive data.

Configuration must be explicit and testable. Prefer typed settings models when the project uses Pydantic or a compatible settings library.

### Python validation and tests

Python deliverables must include project-standard validation:

- unit tests for changed behavior
- API tests for API routes when applicable
- model validation tests for Pydantic models when applicable
- documented commands to run tests and the app
- clear `checks not run` when commands were not executed

Do not claim runnable status unless commands were executed or CI evidence exists.

## Go API contract

For Go HTTP APIs, use swaggo-compatible annotations when the project exposes HTTP handlers. Include `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`, and `@Security` when applicable.

Go request and response structs used by HTTP APIs must include JSON tags and useful comments suitable for generated API documentation.

## Final response requirements

Every final response for repository work must include:

- exact files changed
- branch name and PR number when created
- checks run
- checks not run
- remaining evidence gaps or risks

Do not present generated artifacts or documentation under names different from their actual file paths.
