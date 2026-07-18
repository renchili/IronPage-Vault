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

`tests/api/test_auth_lockout_docker.sh`, executed by `ci/docker_acceptance.sh`, must prove that expired failures do not combine with fresh failures, the fifth fresh failure returns `423 ACCOUNT_LOCKED`, the correct password remains blocked during the lock, an expired lock permits login, and successful login clears both event rows and compatibility fields.

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

Configuration tests and repository contracts must prove missing, conflicting, and bcrypt-incompatible values are rejected. Docker evidence must prove an empty normal-mode generated volume creates one administrator, removing bootstrap values and restarting preserves that identity, acceptance mode creates only its execution-scoped fixtures, and normal mode does not expose `/ui/`.

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

Docker fault-injection tests must force the failed-attempt, login-reset, blacklist, replay, session, and logout persistence paths to fail. Each request must return the documented internal error. A forced logout failure must roll back token blacklisting and leave the session usable; a later successful logout must revoke the same token.

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

Rendering evidence must be described as rendering only. Interaction acceptance requires an executed browser flow covering missing input, incorrect credentials, successful login, network failure and retry, keyboard navigation, visible focus, and understandable result status, with evidence tied to the tested revision.

## Regression and current-HEAD evidence

### Easy-to-make interpretation

A passing historical run, a passing targeted job, or a generated reviewer report can be presented as full current-HEAD acceptance.

### Why it fails

Evidence from another revision does not prove the inspected tree. A reviewer report summarizes evidence rather than generating it. A local entrypoint probe proves only the rows it executes.

### Correct requirement interpretation

Every full-regression claim must identify the exact tested commit, workflow run, generated summary, and retained artifact. A later commit may reuse earlier evidence only for unchanged behavior and must be labelled accordingly; it cannot be called a fresh current-HEAD run.

### Required implementation

The sole workflow must run the complete regression sequentially, stop after the first failure, publish a summary only after success, and retain the complete successful artifact without pushing generated reports to protected `main`. The same workflow must enforce shared concurrency, cooldown, and failed-revision latching.

### Acceptance evidence

Acceptance requires a generated `summary.json` with `overall_status=passed`, all recorded stage statuses equal to zero, and an artifact tied to the tested SHA. When the inspected revision differs from that SHA, the difference and its validation scope must be stated rather than hidden.
