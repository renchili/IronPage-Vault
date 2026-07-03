# Agent Execution Bootstrap

This file is the repository entrypoint for agents. It tells agents what to read, what to generate or update, and which rule sources control repository work.

This file does not replace `AGENT.md` and does not replace `.chatgpt/skills/ironpage-production-workflow/SKILL.md`.

## Rule sources

Agents must read and apply these files before planning, editing, generating files, reviewing, or reporting completion:

1. `AGENT.md` — IronPage Vault project rules. This controls the product scope, domain model, architecture, security, RBAC, workflow, audit, PDF lifecycle, database, backup, API behavior, and required tests.
2. `.chatgpt/skills/ironpage-production-workflow/SKILL.md` — agent workflow rules. This controls repository hygiene, documentation output, evidence, validation, branch/PR behavior, final responses, and compact-safe working records.
3. `README.md`, when present.
4. Existing `docs/` files, when present.
5. Existing source layout, tests, scripts, CI, Docker/deployment files, migrations, and configuration files.

If a required rule source cannot be read, stop and ask the user. Do not continue from memory or guess missing rules.

## What agents must generate or update

For repository work, generate or update only the artifacts required by the current user request, `AGENT.md`, the Skill, and existing repository conventions.

Allowed output categories are:

- production code in the existing source layout.
- tests in the existing test layout.
- migrations or schema files when data shape changes.
- configuration files when runtime behavior requires configuration.
- scripts only when they fit the existing repository workflow or are required for validation.
- `docs/api-spec.md` when API usage or behavior changes.
- `docs/design.md` when architecture, implementation strategy, runtime behavior, configuration, logging, validation, or requirement mapping changes.
- `docs/questions.md` only for clarification answers about unclear process, acceptance, testing, runtime, delivery, usage, or verification points.
- PR notes and final responses containing evidence, checks run, checks not run, and remaining gaps.

Do not generate duplicate project roots, sample applications, placeholder files, noop files, arbitrary reports, unrelated demos, or artifacts outside repository convention.

## What agents must obey

- `AGENT.md` controls what IronPage Vault is and what the implementation must enforce.
- The Skill controls how repository work is performed, documented, validated, and reported.
- Existing repository structure controls where files belong.
- Latest user feedback controls the current correction or narrowed scope.

If these sources appear to conflict, stop and ask the user which rule controls. Do not silently choose one.

## Boundary requirements

Agents must keep implementation, tests, demos, fixtures, and acceptance aids separate.

- Production code must not depend on test helpers, mocks, random sample data, or demo-only configuration.
- Tests, fixtures, mocks, and sample data must stay in test, fixture, example, or docs paths.
- Acceptance probing aids must be documented as testing aids, not product frontend scope.
- Documentation must describe implemented behavior and verified evidence, not desired behavior without code or proof.

## Context continuation

After compaction, model switch, long pause, or a new continuation, agents must re-read this file, `AGENT.md`, and the Skill before continuing repository work.

A compact working record must preserve:

- current branch and base branch.
- files changed.
- user corrections that changed the requirement.
- checks run.
- checks not run.
- open risks or evidence gaps.
- user-local commands given and whether results were received.

## Final response requirements

Every final response for repository work must include:

- exact files changed.
- branch name and PR number when created.
- checks run.
- checks not run.
- remaining evidence gaps or risks.

Do not present generated artifacts or documentation under names different from their actual file paths.
