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

## Static reviewer acceptance

### Easy-to-make interpretation

A reviewer can run tests, build containers, trigger CI, or generate reports to close whatever evidence gaps appear during acceptance.

### Why it fails

That changes the target environment, creates reviewer-owned evidence, can trigger unrelated work, and turns external execution into a stopping condition instead of completing source inspection.

### Correct requirement interpretation

Acceptance is static and read-only. Missing runtime, deployment, interaction, CI, or full-regression evidence does not alter the static verdict. `NOT VERIFIED` is reserved for required source, package content, or rule material that was inaccessible or not inspected.

### Required implementation

`skills/full-project-acceptance-hard-gates/SKILL.md` must prohibit reviewer execution of project code, scripts, tests, builds, containers, databases, browsers, deployments, and CI. It must require complete source inspection, forbid waiting for CI, and continue through every applicable Gate 0–27 after each blocker.

### Acceptance evidence

The acceptance report states reviewer execution as none, distinguishes optional external artifacts from static proof, completes all static gates, and does not claim that source inspection is an executed runtime result.

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
