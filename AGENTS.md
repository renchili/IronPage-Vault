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

## Language-neutral generation contract

The same generation constraints apply no matter which programming language or framework is used. Do not rely on Go-only, Python-only, or framework-specific assumptions unless the repository or user explicitly selects that stack.

Before writing code, identify and state the project type:

- API service
- CLI tool
- background worker
- library/package
- data-processing job
- test/demo utility

If the project type cannot be determined from the user request or repository context, ask one direct clarification question before writing files.

## Production, test, demo, and fixture boundaries

Generated projects must keep production code, tests, examples, demos, mocks, and fixtures clearly separated.

- Production runtime code must live in the repository's existing source layout or a conventional source directory.
- Tests must live in the repository's existing test layout or a conventional test directory.
- Test fixtures, mock data, sample files, and fake credentials must live under test, fixture, example, or docs paths, not inside production runtime paths.
- Demo entry points must be marked as demo/example code and must not be required for production startup.
- Do not make production behavior depend on test helpers, mocks, random sample data, or demo-only configuration.
- Do not include test dependencies as runtime dependencies unless the ecosystem requires it and the reason is documented.
- Do not mix acceptance scripts, manual probes, browser demo pages, or screenshot helpers into the product scope unless the user explicitly asks.
- When a browser UI or script exists only for acceptance probing, document it as an acceptance/testing aid, not as a product frontend.

## Project layout contract

Use the existing repository layout when present. For a new project, choose the smallest conventional layout for the selected language and project type.

A generated project must have clear locations for:

- production source code
- tests
- configuration
- migrations or schema files, when applicable
- scripts or task runners, when applicable
- documentation
- examples or fixtures, only when needed

Do not create multiple competing entry points, package roots, framework variants, or parallel sample applications.

## API contract across languages

For API services in any language, the agent must provide a typed or schema-backed external contract appropriate for the selected ecosystem.

- API routes must have documented methods, paths, auth behavior, request fields, response fields, error responses, and examples.
- API projects must expose an OpenAPI schema or ecosystem-equivalent API schema when the framework supports it.
- API projects should expose a Swagger-like or Redoc-like documentation UI when practical for the selected framework.
- Request, response, error, and configuration shapes must be represented with the selected ecosystem's normal schema/model mechanism.
- Document endpoint behavior in `docs/api-spec.md`.

Examples of acceptable ecosystem-specific mechanisms:

- Go HTTP APIs: swaggo-compatible annotations and JSON-tagged request/response structs.
- Python APIs: Pydantic models plus framework-appropriate OpenAPI/Swagger UI, such as FastAPI native docs, flask-smorest/flask-restx/apispec/flasgger, or drf-spectacular/drf-yasg.
- TypeScript/Node APIs: TypeScript types plus Zod, TypeBox, OpenAPI decorators, tRPC schema, or framework-supported OpenAPI generation.
- Java/Kotlin APIs: typed DTOs plus Springdoc/OpenAPI, Swagger annotations, or framework-supported schema generation.
- .NET APIs: typed request/response models plus Swashbuckle, NSwag, or built-in OpenAPI support.
- Ruby APIs: typed/request validation and OpenAPI generation compatible with the chosen framework when available.

If the framework lacks a mature OpenAPI toolchain, document the limitation in `docs/api-spec.md` and provide a manual API contract with examples.

## Configuration and logging contract

Services in any language must have simple, visible, configurable logging and explicit runtime configuration.

- Do not use ad-hoc print statements as the primary logging system.
- Use the selected ecosystem's standard logger or the repository's existing logger.
- Support log levels such as debug, info, warning/warn, and error.
- Support configurable log format, preferably human-readable locally and JSON for container or production use when appropriate.
- Default logs to stdout or stderr for container-friendly operation.
- Control log level and format through configuration, flags, or environment variables.
- Do not log secrets, tokens, private values, full request bodies, file contents, or sensitive data.
- Configuration must be explicit and testable.
- Prefer typed configuration models or schemas when the ecosystem supports them.

Document logging and configuration behavior in `docs/design.md`. Document request ID, trace ID, or correlation header behavior in `docs/api-spec.md` when the API exposes or accepts those fields.

## Validation and test contract

Generated code must include tests and runnable validation appropriate to the selected language and project type.

- Use the repository's existing test framework when present.
- Add unit tests for changed behavior.
- Add API tests for API routes when applicable.
- Add schema/model validation tests when external contracts are modeled.
- Add configuration tests when configuration parsing or defaults are introduced.
- Keep tests deterministic and independent from production runtime data.
- Document commands to run tests and start the app.
- Clearly report checks not run when commands were not executed.

Do not claim runnable status unless commands were executed locally or CI evidence exists.

## Language-specific minimums

Language-specific rules are examples of the language-neutral contract, not replacements for it.

- Go HTTP APIs must use swaggo-compatible annotations when the project exposes HTTP handlers. Include `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`, and `@Security` when applicable.
- Go request and response structs used by HTTP APIs must include JSON tags and useful comments suitable for generated API documentation.
- Python projects must use Pydantic or an equivalent typed schema layer when validation or external interface shape matters.
- Python API projects must use framework-appropriate OpenAPI documentation instead of forcing FastAPI assumptions onto other frameworks.

## Final response requirements

Every final response for repository work must include:

- exact files changed
- branch name and PR number when created
- checks run
- checks not run
- remaining evidence gaps or risks

Do not present generated artifacts or documentation under names different from their actual file paths.
