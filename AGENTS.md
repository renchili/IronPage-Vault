# Agent Execution Bootstrap

This file is the mandatory repository entrypoint for agents.

It does not define product requirements. Project-adapted rules belong in the repository's project rule file.
It does not define reusable workflow details. Reusable workflow rules belong in repository Skill files.

## Tolerant rule-path resolution

Agents must resolve rule files by meaning and repository contents, not reject work because a reference differs only by capitalization, singular/plural spelling, or glob notation.

The following references are discovery hints, not brittle literal-path requirements:

- `AGENT.md` and `AGENTS.md` mean root-level agent rule files whose basename matches `agent` or `agents`, case-insensitively.
- `skill/**/skill.md` and `skills/**/SKILL.md` mean Skill files under a root directory named `skill` or `skills`, with a file named `skill.md`, all matched case-insensitively and at any nested depth.
- `**` is a recursive discovery marker, not characters that must appear in a real path.

Required behavior:

- inspect the repository and resolve every semantically matching concrete file;
- when both singular and plural root agent files exist, read both;
- prefer the repository's actual stored spelling when opening, editing, citing, or reporting a concrete file;
- tolerate references such as `agent.md`, `AGENT.md`, `agents.md`, `AGENTS.md`, `skill/**/skill.md`, `skills/**/SKILL.md`, and case variants during discovery;
- do not stop, fail, ask the user, or create a replacement merely because the reference used a different case, singular/plural form, or glob spelling;
- do not treat a case-insensitive filesystem collision as two independent rule sources;
- fail for a missing rule only after case-insensitive, singular/plural-tolerant discovery finds no semantically matching concrete file.

A report must name the concrete repository paths actually read. It must not claim that a glob itself was a loaded file.

## Mandatory rule loading

Before planning, editing, generating files, reviewing, committing, opening a PR, or reporting completion, the agent must load rule sources in this order:

1. all root-level agent rule files matching `agent.md` or `agents.md`, case-insensitively;
2. relevant concrete Skill files discovered from `skill/**/skill.md`, using case-insensitive and singular/plural-tolerant matching;
3. `README.md`, when present, using the repository's actual spelling;
4. existing docs, tests, scripts, CI, deployment files, configuration files, and source layout.

For this repository, both `AGENTS.md` and `AGENT.md` exist and have different roles, so both must be read. Their different names are not a reason to stop or ask the user which one is correct.

At least one relevant concrete Skill must be read when the task involves project generation, repair, validation, documentation, repository hygiene, PR creation, or final delivery.

Do not use `.chatgpt/skills/...` as a repository Skill path unless the repository itself explicitly establishes that location.

## Missing project-rule behavior

If tolerant discovery finds no root-level project rule file during ordinary implementation, repair, validation, documentation, or PR work:

- stop before editing files;
- report every discovery form attempted, including singular/plural and case-insensitive matching;
- report which operation failed;
- ask the user whether to provide or create a project rule file;
- do not invent project constraints;
- do not create an alternate rule file unless explicitly requested;
- do not return a raw tool error such as `404` as the final answer.

Do not report `AGENT.md` as missing when a semantically matching case or singular/plural variant exists. Do not report `AGENTS.md` as missing merely because only the project rule file is required for the task.

## Missing Skill behavior

If the task requires a Skill and tolerant discovery finds no relevant concrete Skill file:

- stop before editing files;
- report the conceptual discovery pattern `skill/**/skill.md`;
- report the concrete directories and files inspected;
- report candidate concrete Skill files if any, and why they were not selected;
- ask the user which Skill applies or whether a Skill should be created;
- do not invent a Skill path;
- do not generate a replacement Skill unless explicitly requested.

A capitalization or `skill` versus `skills` difference alone is never a missing-Skill condition.

## Rule metadata integrity

The agent must prove which concrete rules were loaded. A bare statement such as `read the rules` is not enough.

Before editing files, the working record must include rule metadata for every loaded, missing, unreadable, skipped, or blocked rule source:

- concrete repository path as stored;
- role;
- required status;
- read status;
- stable identifier when available: commit SHA, blob SHA, checksum, branch, or ref;
- reason the rule source applies to the current task.

The final response and PR body must include the concrete rule file paths actually read and any genuinely missing, unreadable, skipped, or blocked rule sources. Do not list harmless case, singular/plural, or glob differences as missing sources.

## Rule source hierarchy

- the project-adapted root rule file controls project-specific constraints;
- concrete Skill files control reusable workflow behavior;
- this bootstrap file controls loading, tolerant discovery, metadata, continuation, and reporting behavior.

The agent must obey all loaded rule sources.

If substantive rules conflict, stop and ask the user which rule controls. Different filename capitalization, singular/plural spelling, or glob notation is not a substantive conflict.

## No replacement rule generation

Agents must not generate, regenerate, replace, summarize over, or synthesize replacement rule files during ordinary repository work.

Forbidden unless the user explicitly asks to modify rule files:

- generating or replacing a project rule file;
- creating an alternate project rule file;
- generating a replacement Skill;
- inventing a Skill path;
- copying Skill content into the bootstrap file;
- copying project-specific content into the bootstrap file.

## Required pre-work record

Before making repository changes, the agent must establish a working record containing:

- current branch;
- base branch;
- loaded concrete rule files;
- rule metadata;
- files expected to change;
- checks expected to run under the controlling workflow;
- checks prohibited or unavailable;
- open user feedback that constrains the task.

If this working record cannot be established after tolerant rule discovery, stop before editing files.

## Allowed output boundary

For repository work, the agent may only generate or update files required by the user request, loaded project rules, loaded Skills, or existing repository convention.

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
Do not claim implemented or verified behavior unless backed by evidence allowed by the controlling workflow.

## Context continuation

After compaction, model switch, long pause, new continuation, or loss of working memory, the agent must not continue from memory.

The agent must re-read:

1. all root-level agent rule files found by tolerant discovery;
2. relevant concrete Skill files found by tolerant discovery;
3. current branch and changed files.

Then rebuild the working record, including rule metadata, before editing or reporting completion.

## Final response requirements

Every final response for repository work must include:

- exact files changed using their stored repository paths;
- branch name and PR number when created;
- loaded concrete rule files with metadata identifiers when available;
- genuinely missing, unreadable, skipped, or blocked rule sources;
- checks run;
- checks not run;
- remaining evidence gaps or risks.

Do not present a discovery glob as a loaded file. Do not treat harmless case, singular/plural, or glob differences as defects.