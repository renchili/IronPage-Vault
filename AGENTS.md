# Agent Execution Bootstrap

This file is the mandatory repository entrypoint for agents.

It does not define product requirements. Project-adapted rules belong in `AGENT.md`.
It does not define reusable workflow details. Reusable workflow rules belong in concrete files discovered with the `skills/**/SKILL.md` pattern.

## Canonical path spelling and pattern semantics

Repository paths are case-sensitive and must be read, written, cited, and reported exactly as stored.

- `AGENTS.md` is the concrete bootstrap file at the repository root.
- `AGENT.md` is the distinct concrete project-adapted rule file at the repository root.
- `AGENTS.md` and `AGENT.md` are not aliases, alternatives, singular/plural variants, or case-insensitive spellings of one file.
- `skills/**/SKILL.md` is a discovery glob, not a concrete file path. The canonical directory name is lowercase `skills`; the canonical file name is uppercase `SKILL.md`; `**` represents repository subdirectories.
- An agent must enumerate or search for concrete matches, select the relevant concrete file or files, and report those exact paths, for example `skills/full-project-acceptance-hard-gates/SKILL.md`.
- An agent must not try to open a literal path containing `**`, and must not rewrite the pattern as `skill/**/skill.md`, `skills/**/skill.md`, `Skills/**/Skill.md`, or any other case, number, or spelling variant.
- When a client or filesystem is case-insensitive, the canonical spelling returned by the repository remains authoritative.
- A case-only duplicate or alternate rule file is a repository conflict, not a fallback. Do not create or use one.

## Mandatory rule loading

Before planning, editing, generating files, reviewing, committing, opening a PR, or reporting completion, the agent must load rule sources in this order:

1. the concrete file `AGENTS.md`;
2. the distinct concrete file `AGENT.md`;
3. relevant concrete Skill files discovered with `skills/**/SKILL.md`;
4. `README.md`, when present;
5. existing docs, tests, scripts, CI, deployment files, configuration files, and source layout.

`AGENT.md` is mandatory for ordinary repository work in this repository.

At least one relevant concrete Skill under `skills/` with the exact file name `SKILL.md` must be read when the task involves project generation, repair, validation, documentation, repository hygiene, PR creation, or final delivery.

Do not use `.chatgpt/skills/...` as a repository Skill path.
Do not hard-code a concrete Skill path in `AGENTS.md` as the only allowed Skill; discover the relevant concrete match from the canonical pattern.

## Missing `AGENT.md` behavior

If the exact root file `AGENT.md` is missing or unreadable during ordinary implementation, repair, validation, documentation, or PR work:

- stop before editing files;
- report the exact missing or unreadable path `AGENT.md`;
- report which operation failed;
- ask the user whether to restore `AGENT.md`, provide the correct path, or explicitly proceed without project-adapted rules;
- do not substitute `AGENTS.md`, `agent.md`, `agents.md`, or another case/plural variant;
- do not generate, regenerate, replace, summarize over, or synthesize `AGENT.md`;
- do not create an alternate project rule file;
- do not return a raw tool error such as `404` as the final answer.

The only exception is when the user explicitly asks to create or modify `AGENT.md`. In that case, treat the task as rule-file work and mark missing project constraints as unresolved instead of inventing them.

## Missing Skill behavior

If the task requires a Skill and no relevant concrete file matching the canonical discovery pattern `skills/**/SKILL.md` can be found:

- stop before editing files;
- report the searched discovery pattern exactly as `skills/**/SKILL.md`;
- report that the pattern is a glob and not a literal path;
- report candidate concrete Skill paths if any, and why they were not selected;
- ask the user which concrete Skill applies or whether a Skill should be created;
- do not fall back to `.chatgpt/skills/...`;
- do not search for or accept lowercase/case-variant names as substitutes;
- do not invent a Skill path;
- do not generate a replacement Skill unless the user explicitly asks to create or modify a Skill.

If `AGENT.md` references a specific concrete Skill path and that path is missing or unreadable:

- stop before editing files;
- report the exact referenced path with repository capitalization preserved;
- report that `AGENT.md` referenced it;
- ask the user to restore the Skill, correct the reference, or explicitly change the rule source.

Reusable workflow Skills must stay under `skills/` and use the exact file name `SKILL.md`.

## Rule metadata integrity

The agent must prove which rules were loaded. A bare statement such as `read the rules` is not enough.

Before editing files, the working record must include rule metadata for every loaded, missing, unreadable, skipped, or blocked rule source:

- exact concrete path with canonical capitalization;
- role;
- required status;
- read status;
- stable identifier when available: commit SHA, blob SHA, checksum, branch, or ref;
- reason the rule source applies to the current task.

A discovery glob is not a loaded rule source. The working record, final response, and PR body must name the concrete Skill files actually read.

The final response and PR body must include loaded rule file paths and any missing, unreadable, skipped, or blocked rule sources. Do not claim metadata is verified unless the concrete file was actually read and the identifier was captured from tool output or a local command allowed by the controlling workflow.

## Rule source hierarchy

- `AGENT.md` controls project-adapted constraints for this repository.
- concrete Skill files under `skills/` named `SKILL.md` control reusable workflow behavior.
- `AGENTS.md` controls loading order, canonical path semantics, missing-rule behavior, metadata requirements, and continuation behavior.

The agent must obey all loaded rule sources.

If rule sources appear to conflict, stop and ask the user which rule controls. Do not silently choose one and do not continue with guessed precedence.

## No replacement rule generation

Agents must not generate, regenerate, replace, summarize over, or synthesize replacement rule files during ordinary repository work.

Forbidden unless the user explicitly asks to modify rule files:

- generating or replacing `AGENT.md`;
- creating an alternate project rule file;
- generating a replacement Skill;
- inventing a Skill path;
- creating case/plural variants of `AGENTS.md`, `AGENT.md`, or `SKILL.md`;
- copying Skill content into `AGENTS.md`;
- copying project-specific content into `AGENTS.md`.

## Required pre-work record

Before making repository changes, the agent must establish a working record containing:

- current branch;
- base branch;
- loaded concrete rule files;
- rule metadata;
- files expected to change;
- checks expected to run;
- checks that cannot be run locally;
- open user feedback that constrains the task.

If this working record cannot be established, stop before editing files.

## Allowed output boundary

For repository work, the agent may only generate or update files required by the user request, `AGENT.md`, loaded Skills, or existing repository convention.

Allowed output categories:

- production code in the existing source layout;
- tests in the existing test layout;
- migrations or schema files when data shape changes;
- configuration files when runtime behavior requires them;
- scripts only when they fit existing repository workflow or validation needs;
- documentation files required by loaded Skills;
- PR notes and final response evidence.

Forbidden output unless explicitly requested:

- duplicate project roots;
- sample applications;
- placeholder files;
- noop files;
- arbitrary reports;
- unrelated demos;
- generated artifacts outside repository convention;
- test helpers required by production runtime;
- fixture, mock, sample, or demo data inside production runtime paths.

## Documentation boundary

Documentation must follow loaded Skills and existing repository convention.

Do not invent documentation names when a loaded Skill defines fixed targets.
Do not merge separate document purposes into one loose summary.
Do not claim implemented or verified behavior unless backed by code, tests, CI, logs, reports, or artifacts within the evidence boundary of the controlling workflow.

## Context continuation

After compaction, model switch, long pause, new continuation, or loss of working memory, the agent must not continue from memory.

The agent must re-read:

1. the concrete file `AGENTS.md`;
2. the distinct concrete file `AGENT.md`;
3. relevant concrete Skill files discovered with `skills/**/SKILL.md`;
4. current branch and changed files.

Then rebuild the working record, including rule metadata, before editing or reporting completion.

## Final response requirements

Every final response for repository work must include:

- exact files changed with canonical capitalization;
- branch name and PR number when created;
- loaded concrete rule files with metadata identifiers when available;
- missing, unreadable, skipped, or blocked rule sources;
- checks run;
- checks not run;
- remaining evidence gaps or risks.

Do not present a glob as though it were a loaded file. Do not present generated artifacts or documentation under names different from their actual file paths.