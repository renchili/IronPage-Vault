---
name: documentation-writing
description: Write or update repository documentation with evidence, safety, and gate alignment. Use this skill for README files, project docs, API docs, design docs, usage docs, bilingual docs, and documentation repair.
---

# Documentation Writing Workflow

Use this skill to create, update, repair, or review repository documentation. This is a reusable documentation workflow. Product-specific rules belong in `AGENT.md`; repository workflow rules belong in other loaded Skills.

## Gate alignment

This skill must not be used before repository gates are loaded.

Before writing documentation, the agent must follow `AGENTS.md` and load:

1. `AGENTS.md`.
2. `AGENT.md`.
3. relevant `skills/**/SKILL.md` files, including this one and any broader project-generation or repository workflow Skill.
4. `README.md`, existing docs, source layout, tests, scripts, CI, deployment files, and configuration conventions needed for the task.

The working record must include rule metadata for loaded, missing, unreadable, skipped, or blocked rule sources:

- path.
- role.
- required status.
- read status.
- stable identifier when available, such as blob SHA, commit SHA, checksum, branch, or ref.
- reason the rule source applies to the documentation task.

If `AGENT.md` is missing or unreadable during ordinary documentation work, stop before editing files and report the exact path and failed operation. Do not generate a replacement `AGENT.md` unless the user explicitly asks to create or modify it.

If the task requires a Skill and no relevant Skill under `skills/**/SKILL.md` can be found, stop before editing files. Report the searched pattern, list candidate Skill files if any, and ask which Skill applies. Do not fall back to `.chatgpt/skills/...` and do not invent a Skill path.

## Documentation scope

Use this skill for:

- README files.
- API usage documentation.
- design or architecture documentation.
- security, privacy, backup, restore, deployment, testing, and operation notes.
- clarification-answer documentation such as `docs/questions.md` when a loaded Skill defines that target.
- bilingual Chinese and English documentation.
- documentation review and repair.

Do not use this skill to invent product requirements, implementation status, roadmap, security posture, license terms, deployment topology, API behavior, or architecture.

## Evidence rules

Documentation must be evidence-backed.

- Every feature, command, API endpoint, dependency, architecture statement, security claim, deployment claim, and test claim must be traceable to repository files, user-provided requirements, tool output, CI output, logs, reports, or generated artifacts.
- If evidence is absent, omit the claim or mark it as pending. Do not fill gaps with assumptions.
- Commit titles and branch names may provide weak context only. Do not treat them as implementation proof.
- Do not claim code, tests, Docker, CI, deployment, or acceptance checks ran unless they actually ran and evidence was captured.
- Documentation must distinguish implemented behavior, intended requirement, validation evidence, checks not run, pending items, and risks.

## Documentation evidence ledger

Before writing or updating documentation, build a documentation evidence ledger. The ledger may be a working record, PR body section, or document-internal evidence section depending on the task.

Each non-trivial documentation claim must map to evidence:

- claim or section.
- evidence source path, user input, command output, CI output, log, report, or artifact.
- stable identifier when available: blob SHA, commit SHA, checksum, branch, ref, command, or run ID.
- confidence status: verified, partially verified, not executed, unavailable, or pending user input.
- notes about any limitation, dynamic behavior, generated source, or missing proof.

Do not publish broad claims such as `secure`, `production-ready`, `fully tested`, `cloud-native`, `offline`, `encrypted`, or `role-based` unless the evidence ledger contains proof from project rules, implementation, tests, and validation output.

## Safe scan rules

Scan only files needed for the documentation task.

Do not read, grep, summarize, quote, or list sensitive content from:

- version control internals.
- dependency directories and package caches.
- virtual environments.
- build output, generated output, coverage output, compiled output, and caches.
- logs, dumps, database snapshots, and runtime state.
- local-only environment files and private runtime profiles.
- credential, key, token, service-account, or kube configuration material.

Safe example files may be read when they are clearly examples, templates, or samples.

When private files are discovered by name, do not expose exact private paths in documentation. Use a general note such as: private configuration files are intentionally excluded from documentation.

## Target resolution

Before writing, determine the exact documentation target.

Use the target path from the user request, loaded Skill, existing documentation home, or repository convention. Do not invent arbitrary document names.

When a broader loaded Skill defines fixed targets, preserve them. For example, a project-generation Skill may require:

- `docs/api-spec.md` for API usage and behavior.
- `docs/design.md` for design, architecture, implementation strategy, runtime behavior, configuration, logging, validation, and requirement mapping.
- `docs/questions.md` only for project clarification answers.

If a target path cannot be resolved, ask the user for the path before writing.

If a target file already exists:

- update it in place when the task is an explicit repair, merge, or repository-maintenance task and the loaded Skill says to update existing documentation.
- ask before overwriting when the user asks to generate a standalone README or replacement document and has not granted overwrite permission.
- preserve the file name and path unless the user explicitly requests a rename.

## README target selection

For README-specific work, resolve the target explicitly before writing.

- Chinese only: use `README.md` when the repository uses Chinese as the primary README, unless the user requests `README.zh-CN.md`.
- English only: use `README.md` when the repository uses English as the primary README, unless the user requests `README.en.md`.
- Bilingual: use separate files such as `README.zh-CN.md` and `README.en.md`; create a short `README.md` index only when the repository convention supports it or the user approves.
- Existing README: do not overwrite a standalone README generation target without permission. For repair or maintenance work, update the existing target in place when the user request and loaded Skills require it.

## README handling

README files must be useful, concise, and grounded in repository evidence.

For README work:

1. detect the implementation language and project type before writing.
2. choose commands from manifests, build files, scripts, existing docs, or clear framework conventions.
3. include only sections with meaningful evidence.
4. keep the README focused instead of cataloging every file.
5. use maintainer voice for security, privacy, and license notes.

Common README sections, used only when supported by evidence:

- Project summary.
- Overview.
- Features.
- Tech stack.
- Project structure.
- Getting started.
- Configuration.
- Run.
- Test.
- API.
- Deployment.
- Development notes.
- Security and privacy.
- License.

Skip empty sections and placeholders such as `TODO`, `TBD`, or `coming soon`.

## Language and project-type detection

Detect languages from manifests, build files, and entry points. For mixed repositories, describe each language's role instead of forcing a single label.

Use ecosystem evidence:

- Node projects: package manifests, lockfiles, framework configs, entry points, routes.
- Python projects: dependency manifests, package metadata, framework entry points.
- Java or Kotlin projects: build files, source layout, framework annotations.
- Go projects: module files, main packages, internal packages, public packages.
- Frontend projects: UI config, browser entry files, pages, routes, static assets.
- Backend projects: server entry, controllers, routes, database or migration folders.
- Full-stack projects: both frontend and backend signals or workspace split.

Do not replace or reinterpret the repository's language, framework, database, build path, test runner, project layout, security model, or deployment model.

## API documentation rules

Only include API documentation when API evidence exists.

Accepted evidence includes:

- OpenAPI or Swagger files.
- route or controller definitions.
- request or response schemas, DTOs, validation models, or typed handlers.
- existing docs or tests that explicitly call endpoints.

Rules:

- extract method and path only when both are explicit.
- include handler, controller, or source file when obvious.
- do not infer request or response bodies unless schema, type, validation model, or docs define them.
- do not guess authentication, roles, rate limits, side effects, status codes, or error codes.
- if routes are dynamic or unclear, document only directly confirmable entries and state that runtime assembly prevents complete static listing.
- if endpoints are numerous, group them and point to the source or OpenAPI file.

Preferred compact table:

```markdown
| Method | Path | Source | Notes |
|--------|------|--------|-------|
| GET | `/api/users` | `src/routes/users.ts` | User list endpoint |
```

## Structure and diagrams

A directory tree must be compact and evidence-based.

- max depth 3 unless deeper levels are essential.
- max 35 lines.
- show source, config, tests, docs, deployment files, and manifests.
- omit dependencies, build output, generated code, caches, and private config.

Include Mermaid only when repository evidence shows multiple components or layers.

- Do not draw databases, queues, cloud services, external APIs, model providers, or microservices unless explicitly present in code, config, dependencies, or docs.
- Prefer request-flow or module-relationship diagrams over broad architecture maps.
- Omit diagrams when evidence is weak.
- Keep diagrams small, normally 5 to 9 nodes.

## Output length control

Keep generated documentation practical and reviewable.

- Normal README target: 120 to 220 lines.
- Hard README limit: 300 lines unless the user asks for exhaustive documentation.
- Feature list: normally no more than 8 items.
- Tech stack table: normally no more than 12 rows.
- Directory tree: normally no more than 35 lines.
- API table: normally no more than 20 rows; group or point to the source when larger.
- Command list: normally no more than 8 install, run, test, build, or deploy commands unless the project requires more.
- Do not paste long generated documents into chat unless the user explicitly asks.

## Bilingual documentation

When generating both Chinese and English:

- create separate files such as `README.zh-CN.md` and `README.en.md`, unless the user requests different names.
- keep the two versions equivalent in structure and facts.
- localize section names and explanatory text naturally.
- avoid mixing languages in one paragraph except for commands, code, paths, package names, API names, and standard technical terms.
- if `README.md` is used as an index, keep it short and link to both language files.

## Documentation quality checklist

Before finalizing, verify:

- rule metadata was captured according to `AGENTS.md`.
- no excluded private files were read, summarized, or exposed.
- target paths were resolved and existing-file handling was respected.
- README target language and overwrite behavior were resolved when relevant.
- language and project type were detected before generation.
- commands come from manifests, scripts, docs, or clear framework conventions.
- API entries are backed by explicit route, schema, OpenAPI, docs, or test evidence.
- architecture and deployment claims are backed by repository evidence.
- directory tree matches real files and excludes generated, private, dependency, and build output.
- documentation evidence ledger covers non-trivial claims.
- security and license notes avoid invented claims and use maintainer voice.
- bilingual documents are factually aligned when both are generated.
- documentation contains no placeholders.
- final response lists exact files changed, loaded rules with identifiers when available, checks run, checks not run, and remaining evidence gaps.

## Final response for documentation tasks

Every documentation task final response must include:

- exact documentation files created or updated.
- branch name and PR number when created.
- loaded rule files and metadata identifiers when available.
- source evidence used.
- checks run.
- checks not run.
- pending items, missing evidence, or risks.

Do not paste long generated documents into chat unless the user explicitly asks. Summarize the changed files and key sections instead.
