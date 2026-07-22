# Requirement Clarifications

This document records stable requirement interpretations that affect implementation and acceptance. Each topic states the easy mistake, why it fails, the correct boundary, the required implementation, and the evidence needed to accept it.

## Rolling failed-login lockout

### Easy-to-make interpretation

A single cumulative `failed_attempts` counter can appear to satisfy the requirement to lock an account after five failures.

### Why it fails

A cumulative counter has no time boundary. Failures separated by hours or days can be added together even though the requirement is five failures inside a rolling window. It also gives no durable event history for proving that old attempts expired.

### Correct requirement interpretation

Only failed attempts whose timestamps fall within the preceding 15 minutes count toward lockout. The fifth in-window failure locks the account for 15 minutes. Attempts older than the window do not count, an active lock rejects even the correct password, and a successful login after the lock expires clears the failed-attempt state.

### Required implementation

Persist failed attempts in `login_attempts`. Serialize updates for one user, delete events older than the rolling cutoff, insert the new event, count current events, and update `users.failed_attempts` and `users.locked_until` in one transaction. Successful login must atomically clear the attempt events and user lock fields while creating the server-side session.

### Acceptance evidence

`tests/api/test_auth_lockout_docker.sh` defines the expiry, fifth-attempt lock, active-lock rejection, recovery, and state-clearing paths. Static acceptance inspects those definitions; an existing lifecycle artifact may be cited only as optional execution context.

## Initial administrator and acceptance fixtures

### Easy-to-make interpretation

The requirement for usable local accounts can be interpreted as embedding one administrator and several reusable seeded credentials in normal product configuration.

### Why it fails

Fixed identities make separate installations share long-lived credentials and mix acceptance fixtures with product state. Changing those credentials after startup does not remove the unsafe initialization path.

### Correct requirement interpretation

Normal mode initializes one generated administrator only when the user table is empty. Existing users are never overwritten on restart. Acceptance fixtures are allowed only when explicit acceptance mode is enabled, and their values must be supplied by the execution environment. The browser probe under `/ui/` belongs to that acceptance-only mode.

### Required implementation

The deployment layer must generate installation-specific bootstrap and complete local runtime configuration. Application validation must reject missing or conflicting values. Acceptance startup must require execution-scoped fixture values and must not accept bootstrap values at the same time. Product code, image configuration, Compose, and browser assets must contain no fixed credential.

### Acceptance evidence

Configuration, source, and repository contracts define empty-installation creation, restart preservation, acceptance fixtures, and normal-mode UI exclusion. Existing Docker evidence is optional context and does not control the static verdict.

## Authentication state failures must fail closed

### Easy-to-make interpretation

Database updates around authentication can be treated as best-effort because password verification and token parsing already succeeded in memory.

### Why it fails

Ignoring a failed lockout update can allow repeated guesses without durable enforcement. Ignoring blacklist, replay, session, or logout persistence failures can return authenticated or logged-out success when the database state says otherwise.

### Correct requirement interpretation

Authentication and logout succeed only when their required state reads and writes succeed. A database error in failed-attempt recording, successful-login reset, blacklist lookup, replay recording, session activity, or logout revocation must return the standard internal-error envelope and stop the request. Logout must update blacklist and session state atomically.

### Required implementation

Check every database read, write, transaction commit, and affected-row result on the authentication path. Use transactions for rolling-window updates, successful-login state creation, and logout revocation. Do not emit a token or `logged_out` response after an incomplete state mutation.

### Acceptance evidence

Source and test definitions cover failed-attempt, login-reset, blacklist, replay, session, and logout persistence failures, including rollback of an incomplete logout. Existing fault-injection artifacts are optional context rather than static acceptance prerequisites.

## Acceptance browser surface

### Easy-to-make interpretation

The presence of files under `public/` or one rendered screenshot can be interpreted as a product frontend or as complete browser interaction coverage.

### Why it fails

IronPage Vault is a backend deliverable. The browser surface is an acceptance aid, and a screenshot proves only static rendering. It does not prove submission, validation errors, network recovery, retry, keyboard operation, focus management, or accessible status updates. Multiple served HTML pages also create an ambiguous acceptance contract.

### Correct requirement interpretation

`/ui/` is served only in acceptance mode and is backed by one canonical source, `public/index.html`. Browser interaction evidence improves acceptance confidence but does not create a required production frontend.

### Required implementation

Keep the canonical UI behind `ACCEPTANCE_MODE`, do not embed fixture values, remove duplicate served surfaces, and preserve API behavior as the source of truth. Browser tests operate the same canonical page.

### Acceptance evidence

Static inspection traces missing input, incorrect credentials, successful login, network failure and retry, keyboard navigation, visible focus, and understandable result status through the canonical page and its test definitions. The reviewer does not execute that flow.

## Buildable frontend design and interaction handoff

### Easy-to-make interpretation

A list of screens, a polished screenshot, generic component names, or prose such as “use an appropriate icon” can be treated as enough design output for later frontend development. Adding several YAML or JSON files can also appear to make an underspecified design implementation-ready.

### Why it fails

The frontend engineer still has to choose the component library, icon, dimensions, spacing, state behavior, special-interaction rules, accessibility behavior, and host-application conventions. Those decisions change the product and cannot be reconstructed reliably from a static image or generic prose. Arbitrary structured files add format without resolving the missing design decisions.

### Correct requirement interpretation

When a requested project includes a production UI or implementation-guiding prototype, the result must resolve the visual, component, interaction, responsive, accessibility, and target-platform decisions needed to build it. Exact icon choice and visual size, control and layout dimensions, full state behavior, special interaction lifecycle, and applicable app-review format are part of the requirement. Backend-only projects and explicitly acceptance-only probes remain outside that production-frontend obligation.

### Required implementation

`skills/project-generation-workflow/SKILL.md` must require the agent to inspect the target framework, host application, design system, component library, icon library, viewport constraints, and platform review rules before generating UI. It must require exact components, icons, dimensions, state variants, special-interaction commit/cancel/failure behavior, API and permission mapping, and implementation-ready output. It must forbid vague design placeholders and must not invent YAML, JSON, manifests, registries, or review packs unless the request, platform, loaded rules, or repository convention requires them. `skills/full-project-acceptance-hard-gates/SKILL.md` must reject UI output that still requires material redesign before coding.

### Acceptance evidence

Static contracts confirm both Skills contain the implementation-readiness, exact icon and sizing, special-interaction, platform-review, accessibility, traceability, and artifact-format boundaries. A UI acceptance report must either map those items to current source and design evidence or give a justified `N/A`; a screenshot, prose-only mock, or arbitrary structured package cannot establish readiness by itself.

## Rule path discovery without filename nitpicking

### Easy-to-make interpretation

An agent can treat `AGENT.md`, `AGENTS.md`, `agent.md`, `agents.md`, `skill/**/skill.md`, and `skills/**/SKILL.md` as unrelated literal requirements, then stop because the reference capitalization, singular/plural spelling, or glob notation does not exactly match one repository path.

### Why it fails

Those spelling differences often express the same discovery intent. Treating `**` as literal path characters or treating a harmless case difference as a missing rule creates false blockers, especially across case-sensitive and case-insensitive clients. It also encourages creation of duplicate rule files instead of reading the files already present.

### Correct requirement interpretation

Rule discovery is tolerant, while concrete reporting is precise. Root files whose basename is `agent` or `agents` are discovered case-insensitively. Skill files are discovered recursively under `skill` or `skills` directories with a case-insensitive `skill.md` filename. When multiple semantically relevant concrete files exist, read them all. After discovery, use the repository's stored path when opening, editing, citing, or reporting.

### Required implementation

`AGENTS.md` and both repository Skills must define case-insensitive, singular/plural-tolerant discovery and must state that `**` is a recursive marker rather than a literal filename component. Agents must not stop, fail, ask the user, or synthesize a replacement solely because a reference uses a different harmless spelling. A missing-rule result is allowed only after tolerant discovery finds no semantically matching concrete file.

### Acceptance evidence

Static repository contracts verify tolerant discovery phrases, prohibit literal-glob and case-only missing-rule behavior, and require reports to name concrete repository paths actually read. The current repository still reports the stored paths `AGENTS.md`, `AGENT.md`, `skills/project-generation-workflow/SKILL.md`, and `skills/full-project-acceptance-hard-gates/SKILL.md` after resolving them tolerantly.

## Static reviewer acceptance

### Easy-to-make interpretation

A reviewer can run tests, shell commands, Python helpers, build containers, trigger CI, or create a temporary remote workflow to close evidence gaps. Missing network or local tools can be treated as a reason to outsource execution, and wording such as “完整测试” can be treated as implicit permission to run the repository.

### Why it fails

That changes the target environment, creates reviewer-owned evidence, can trigger unrelated work, and turns external execution into a stopping condition instead of completing source inspection. An offline or incomplete environment makes such execution less trustworthy, not more permissible. Ambiguous test wording does not override the repository's static-only workflow.

### Correct requirement interpretation

Acceptance is static and read-only. Missing runtime, network, local checkout, deployment, interaction, CI, or full-regression evidence does not alter the static verdict. Under the static Skill, “test”, “full test”, “完整测试”, regression, acceptance, and report-generation requests mean a complete Gate 0–27 source review unless the user explicitly requests a separate runtime workflow and controlling rules allow it. `NOT VERIFIED` is reserved for required source or rule material that remains inaccessible after non-executing retrieval attempts.

### Required implementation

`skills/full-project-acceptance-hard-gates/SKILL.md` and `skills/project-generation-workflow/SKILL.md` must prohibit shell, terminal, Python, code interpreter, notebooks, local `git`/`gh`, package managers, project scripts, builds, containers, databases, browsers, deployments, CI actions, temporary workflows, and asking the user to run commands. They must require read-only repository APIs or equivalent source viewers, complete source inspection, and continuation through every applicable Gate 0–27.

### Acceptance evidence

The acceptance report states every execution category as `none`, distinguishes optional pre-existing artifacts from static proof, completes all static gates, and does not claim that source inspection is an executed runtime result. Static contracts require the no-network/no-tool fallback and ambiguous-test-wording rules.

## CI admission and one-time unlock

### Easy-to-make interpretation

A target-wide cooldown and failure state can be applied to every later commit in the same pull request to prevent repeated execution.

### Why it fails

That blocks the corrective revision needed to repair a failed check. A new source revision is not a repeat of the failed revision. GitHub also creates a workflow-run object before repository YAML executes; a guard after checkout is too late, sleeping consumes a runner, a boolean unlock is repeatable, and a first-page history scan is not durable.

### Correct requirement interpretation

Repository CI must collapse superseded active target runs, admit before checkout and repository-controlled code, reject rather than sleep, paginate relevant history, latch failed target/revision pairs, deny ordinary reruns, and consume one exact unlock tied to target, revision, failed run ID, and reviewed reason. Cooldown and failure latching apply to the exact revision; a different corrective revision must be admitted immediately.

### Required implementation

Use one shared target key, `cancel-in-progress: true`, an admission job before checkout, complete Actions pagination, a ten-minute same-revision cooldown, exact failed-run authorization, and an auditable run-name consumption marker. Filter cooldown history by `head_sha === currentSha`, and upload the static source manifest only after every static gate succeeds.

Repository YAML must not claim literal pre-dispatch prevention. That stronger property needs external platform controls.

### Acceptance evidence

Static inspection confirms the workflow and documentation agree that the same failed revision is blocked while a new revision is admitted. A claim that no workflow-run object or admission runner can ever start requires separate platform-level evidence.

## Regression and current-revision evidence

### Easy-to-make interpretation

A passing historical run, a passing static job, or a generated reviewer report can be presented as executed runtime proof for the current revision.

### Why it fails

Evidence from another revision does not prove execution of the inspected tree. A reviewer report summarizes static findings rather than generating runtime evidence. A static workflow proves repository properties, not executed product behavior.

### Correct requirement interpretation

Static acceptance is decided from current source and test definitions without requiring a full-regression artifact. When an existing execution artifact is cited for a runtime claim, it must identify the exact tested commit and scope; tree equivalence must be explicit rather than assumed.

### Required implementation

Keep `ci/run_full_regression.sh` as a separate manual or normal-lifecycle entrypoint. Keep `.github/workflows/ci.yml` static-only. The static workflow must not run or claim full regression, Docker, API, browser, database, or deployment acceptance.

### Acceptance evidence

The full-regression definition statically includes all required stages, fail-fast propagation, truthful summaries, revision fields, and artifacts. Existing generated summaries are optional read-only context and do not alter the static verdict.