# Agent Execution Bootstrap

This file is the mandatory repository entrypoint for agents.

It does not define product requirements. Project-adapted rules belong in `AGENT.md`.
It does not define reusable workflow details. Reusable workflow rules belong in repository Skill files.

## Tolerant rule-source discovery

Rule references are discovery hints, not brittle filename tests.

Agents must:

- discover root agent rule files whose basename is `agent` or `agents`, case-insensitively;
- discover Skill files recursively under a root directory named `skill` or `skills`, with a filename matching `skill.md` case-insensitively;
- treat `**` as a recursive discovery marker, not literal path characters;
- read every relevant concrete match;
- use the repository's actual stored path when opening, editing, citing, or reporting a file;
- tolerate references such as `AGENT.md`, `AGENTS.md`, `agent.md`, `agents.md`, `skill/**/skill.md`, and `skills/**/SKILL.md` during discovery;
- avoid reporting harmless capitalization, singular/plural, or glob differences as missing sources, conflicts, naming defects, or blockers.

When both `AGENTS.md` and `AGENT.md` exist, read both. Their different names do not justify stopping or asking which one is correct.

A rule may be reported missing only after case-insensitive, singular/plural-tolerant discovery finds no semantically matching concrete file. Reports must name concrete paths actually read; a glob is not a loaded file.

## Static-only task interpretation

When a loaded Skill defines generation, repair, review, validation, acceptance, or reporting as static-only, that boundary is absolute.

The agent must not use any of the following to complete or validate the task:

- shell or terminal commands;
- Python, a code interpreter, notebook, or generated helper program;
- local `git`, `gh`, `curl`, `wget`, package managers, compilers, formatters, or linters;
- project code, repository scripts, tests, binaries, generators, or build commands;
- containers, databases, browsers, emulators, deployments, or local runtime services;
- temporary workflows, execution branches, remote runners, or CI jobs created to outsource validation;
- requests that the user run commands and return results.

No network, unavailable internet, an absent local checkout, missing tools, missing dependencies, unavailable Docker, an unavailable database, or an unavailable browser never relaxes a static-only boundary and never authorizes a different execution path.

Use read-only repository APIs, connector file reads, uploaded-file viewers, archive viewers that do not execute content, or equivalent non-executing source retrieval. Repository writes may use a source-control file API only when the user explicitly requests repository changes; write access does not authorize execution.

Within a static-only task, wording such as `test`, `full test`, `complete test`, `完整测试`, `regression`, `acceptance`, or `generate a test report` means complete static source inspection under the loaded Skill unless the user explicitly requests a separate runtime workflow and the controlling repository rules permit runtime execution. Ambiguous wording is not execution permission.

Missing runtime evidence, CI, logs, screenshots, builds, deployment results, or test output does not by itself cause failure, conditional status, or a waiting state. If required source remains inaccessible through non-executing retrieval, record the exact concrete source gap and complete every other statically inspectable part.

A downloadable-report request does not authorize shell, Python, notebooks, helper code, or repository report-file creation. Deliver the report in the response or through a genuinely non-executing artifact capability. Do not add a report to the repository unless the user explicitly names the repository path.

## Mandatory rule loading

Before planning, editing, generating files, reviewing, committing, opening a PR, or reporting completion, the agent must load rule sources in this order:

1. all root-level agent rule files discovered by the tolerant rule-source rules above;
2. relevant concrete Skill files discovered recursively;
3. `README.md`, when present;
4. existing docs, tests, scripts, CI, deployment files, configuration files, and source layout relevant to the task.

For this repository, both `AGENTS.md` and `AGENT.md` exist and have different roles, so both are mandatory for ordinary repository work.

At least one relevant concrete Skill must be read when the task involves project generation, repair, validation, documentation, repository hygiene, PR creation, or final delivery.

Do not use `.chatgpt/skills/...` as a repository Skill path unless the repository itself explicitly establishes that location.

## Missing project-rule behavior

If tolerant discovery finds no root-level project rule file during ordinary implementation, repair, validation, documentation, or PR work:

- stop before editing files;
- report the concrete locations and tolerant discovery forms inspected;
- report which operation failed;
- ask the user whether to provide or create a project rule file;
- do not invent project constraints;
- do not create an alternate rule file unless explicitly requested;
- do not return a raw tool error such as `404` as the final answer.

Do not report `AGENT.md` or `AGENTS.md` as missing when a semantically matching case or singular/plural variant exists.

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

- `AGENT.md` controls project-adapted constraints for this repository.
- relevant concrete Skill files control reusable workflow behavior.
- `AGENTS.md` controls loading order, tolerant discovery, missing-rule behavior, metadata requirements, file-creation boundaries, static-only interpretation, and continuation behavior.

The agent must obey all loaded rule sources.

If substantive rules conflict, stop and ask the user which rule controls. Different filename capitalization, singular/plural spelling, or glob notation is not a substantive conflict.

## No replacement rule generation

Agents must not generate, regenerate, replace, summarize over, or synthesize replacement rule files during ordinary repository work.

Forbidden unless the user explicitly asks to modify rule files:

- generating or replacing `AGENT.md`;
- creating an alternate project rule file;
- generating a replacement Skill;
- inventing a Skill path;
- copying Skill content into `AGENTS.md`;
- copying project-specific content into `AGENTS.md`.

## Required pre-work record

Before making repository changes, the agent must establish a working record containing:

- current branch;
- base branch;
- loaded concrete rule files;
- rule metadata;
- files expected to change;
- files expected to be created, which should normally be none;
- exact requirement source and ownership justification for every proposed new file;
- checks expected under the controlling workflow;
- checks prohibited or unavailable;
- open user feedback that constrains the task.

If this working record cannot be established, stop before editing files.

## Strict file-creation boundary

Updating an existing file is the default. New files are forbidden unless one of these conditions is satisfied:

1. the user explicitly requests the exact new repository path;
2. a loaded rule contains a task-specific mandatory statement naming the exact required path and that path is genuinely absent;
3. a required production, schema, migration, or test component has no existing owner file and cannot be implemented correctly in the existing repository layout.

Generic, default, example, fallback, suggested, or illustrative output-path lists in a Skill do not satisfy condition 2 and do not authorize creating files. A heading such as `Default outputs`, a reusable example, or a path mentioned only as a possible convention is not a task-specific mandate.

A broad request such as `update documentation`, `add tests`, `generate a report`, `document the change`, `complete the project`, or `fix everything` does not authorize creating arbitrary files.

Before creating any file, the agent must record:

- the exact path;
- the requirement source that mandates it;
- why no existing file can own the content;
- the file's long-term owner and repository convention;
- whether the same purpose already exists elsewhere.

If that record cannot be established, do not create the file.

Documentation rules are stricter:

- update existing canonical documentation instead of creating another document;
- do not create implementation-status, follow-up, review-fix, roadmap, questions, report, acceptance-summary, handoff, notes, checklist, or assistant-history documents unless the user explicitly names the exact path or a task-specific mandatory rule requires that exact path;
- do not create duplicate README files, alternate API documents, versioned copies, timestamped reports, or one-document-per-finding files;
- do not turn a temporary working record, review report, tool output, or assistant explanation into repository documentation;
- the existence of a `docs/` directory is not permission to add more files;
- repository convention means an established exact path or an unmistakable existing file family, not merely a directory where a new file could fit.

Completeness means implementing all required behavior across the correct existing owners. It does not mean maximizing file count.

## Allowed output boundary

For repository work, the agent may only update files required by the user request, `AGENT.md`, loaded Skills, or existing repository convention, subject to the strict file-creation boundary above.

Allowed output categories:

- production code in the existing source layout;
- tests in the existing test layout;
- migrations or schema files when data shape changes;
- configuration files when runtime behavior requires them;
- scripts only when they fit existing repository workflow or validation needs;
- updates to existing documentation required by the implementation;
- PR notes and final response evidence that remain outside repository files unless explicitly requested.

Forbidden output unless explicitly requested and justified:

- duplicate project roots;
- sample applications;
- placeholder files;
- noop files;
- arbitrary reports;
- unrelated demos;
- generated artifacts outside repository convention;
- new documentation created only to narrate the agent's work;
- test helpers required by production runtime;
- fixture, mock, sample, or demo data inside production runtime paths.

## Documentation boundary

Documentation must follow loaded Skills and existing repository convention.

Do not invent documentation names when a loaded Skill defines fixed targets.
Do not create a new document when an existing document already owns the topic.
Do not merge separate document purposes into one loose summary merely to avoid understanding their ownership.
Do not claim implemented or verified behavior unless backed by evidence allowed by the controlling workflow.

## Context continuation

After compaction, model switch, long pause, new continuation, or loss of working memory, the agent must not continue from memory.

The agent must re-read:

1. all root-level agent rule files found by tolerant discovery;
2. relevant concrete Skill files found by tolerant discovery;
3. current branch and changed files.

Then rebuild the working record, including rule metadata and proposed file-creation justifications, before editing or reporting completion.

## Final response requirements

Every final response for repository work must include:

- exact files changed using stored repository paths;
- exact files created, normally `none`, with a requirement justification for each created file;
- branch name and PR number when created;
- loaded concrete rule files with metadata identifiers when available;
- genuinely missing, unreadable, skipped, or blocked rule sources;
- checks run;
- checks not run;
- repository reads and writes actually performed;
- remaining evidence gaps or risks.

Do not present generated artifacts or documentation under names different from their actual file paths. Do not present a discovery glob as a loaded file. Do not treat harmless case, singular/plural, or glob differences as defects.
