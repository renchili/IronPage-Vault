# Questions and Decisions

This document records project-related questions from the conversation and explains the decisions made. It avoids duplicated rule text and explains why each decision was made.

## 1. What should the project be called?

Answer: **IronPage Vault**.

Why this name was chosen:

- `IronPage` keeps the PDF/legal document identity.
- `Vault` communicates local containment, security, auditability, and controlled access.
- It sounds more appropriate for legal and compliance work than a casual document editor name.

## 2. What is the short English description?

Answer:

> IronPage Vault is an air-gapped legal PDF lifecycle system for secure document review, redaction, version control, Bates numbering, workflow tracking, audit logging, notifications, and role-based access control.

Why this description was chosen:

- It stays short enough for metadata and project listings.
- It names the core lifecycle capabilities.
- It emphasizes air-gapped legal PDF management rather than generic file storage.

## 3. What should be in `metadata.json`?

Answer:

`metadata.json` must include:

- `project_type`
- `prompt`
- `backend_language`
- `backend_framework`
- `frontend_language`
- `frontend_framework`
- `cwd`
- `database`

Why this structure was chosen:

- It gives automated agents enough context to understand the required stack.
- It preserves the full original prompt so future implementation steps can be checked against the actual requirement.
- It avoids relying only on README summaries.

## 4. Should `CLAUDE.md` duplicate `AGENT.md`?

Answer: No.

Why this decision was made:

- Duplicating two rule files creates drift.
- The user explicitly clarified that Claude should reference `AGENT.md`, not maintain two edited copies.
- `CLAUDE.md` now points to `AGENT.md` as the single rule source.

## 5. Why is `AGENT.md` the single source of truth?

Answer:

The project has many strict requirements: runnability, RBAC, workflow, audit, PDF processing, testing, Docker deployment, and offline operation. Keeping those rules in one file prevents mismatched instructions.

Why this matters:

- Implementation agents can read one rule file first.
- Requirement updates are less likely to be missed.
- Acceptance reviewers can compare code against one canonical project rule set.

## 6. Why is the backend written in Go with Echo and sqlx?

Answer:

The prompt explicitly requires Go, Echo, sqlx, and PostgreSQL.

Why this matters:

- Using another framework or ORM would be off-prompt.
- Echo gives a direct REST API layer.
- sqlx keeps database interaction explicit while still improving scanning and query ergonomics.

## 7. Why is PostgreSQL the only database?

Answer:

The prompt states that PostgreSQL is the sole persistence layer for metadata, audit records, configuration dictionaries, notification queues, and related system records.

Why this matters:

- It keeps the deployment simple for air-gapped use.
- It avoids adding Redis, SQLite, Elasticsearch, or other stores that would complicate acceptance.
- PostgreSQL can handle sessions, audit logs, workflow history, and metadata in one local system.

## 8. Why are PDF binaries stored on the filesystem?

Answer:

The prompt requires PDF binary assets to reside on the local filesystem and be referenced by database pointers.

Why this matters:

- Large PDFs are not forced into PostgreSQL rows.
- File paths, hashes, versions, and metadata remain queryable.
- Local volume backup and filesystem snapshot strategies are easier to document.

## 9. Why is the project single-container?

Answer:

The prompt requires a single Docker container deployable on a standalone machine with no external network dependencies.

Why this matters:

- The evaluator can run one service locally.
- PostgreSQL and the API are packaged together.
- This matches the air-gapped requirement even though multi-container deployment is more common in general production environments.

## 10. Why provide `docker-compose.yml` if the prompt says single container?

Answer:

`docker-compose.yml` provides the one-command startup path while still defining only one service container.

Why this matters:

- Acceptance instructions require simple startup.
- `docker compose up --build` is easy to verify.
- Volumes for PostgreSQL data, PDF storage, and backups can be declared clearly.

## 11. Why are there three seeded users?

Answer:

The system needs Admin, Editor, and Reviewer test coverage.

Why this matters:

- Acceptance can immediately test RBAC without manually creating accounts first.
- The role matrix can be verified locally.
- Seeded users make API tests deterministic.

## 12. Why is Admin not automatically allowed to edit documents?

Answer:

The prompt separates Admin management capabilities from Editor document manipulation capabilities.

Why this matters:

- It prevents accidental privilege expansion.
- It keeps legal workflow duties separated.
- It makes RBAC acceptance tests more meaningful.

## 13. Why require `X-Request-ID`?

Answer:

The prompt requires anti-replay behavior. A unique request ID lets the server detect repeated requests for the same token/session.

Why this matters:

- It provides deterministic replay checks without external infrastructure.
- It gives audit logs a request correlation key.
- It supports troubleshooting and acceptance validation.

## 14. Why require `X-Request-Timestamp`?

Answer:

The prompt requires rejecting old request payloads.

Why this matters:

- It reduces the replay window.
- It keeps request validation local and deterministic.
- It gives the API a clear security contract.

## 15. Why keep server-side session records with JWT?

Answer:

The prompt requires eight-hour inactivity expiration and token blacklist support. A purely stateless JWT is not enough for those requirements.

Why this matters:

- Inactivity can be enforced through `last_seen_at`.
- Logout can revoke a token immediately through the blacklist.
- Session state supports audit and local control.

## 16. Why use a mandatory workflow chain?

Answer:

The prompt defines this chain:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Why this matters:

- Legal document state must be predictable.
- Finalized records must be protected.
- Tests can verify valid and invalid transitions.

## 17. Why make Finalized documents immutable?

Answer:

The prompt states that Finalized status permanently renders the record immutable.

Why this matters:

- It preserves legal and compliance integrity.
- It prevents post-finalization edits, rollback, redaction, annotation, or numbering changes.
- It is one of the highest-priority acceptance gates.

## 18. Why is redaction two-phase?

Answer:

The prompt requires staged coordinate-bound metadata first and Editor burn-in confirmation second.

Why this matters:

- Reviewers or Editors can inspect proposed regions before irreversible action.
- Burn-in confirmation is auditable.
- It separates planning from permanent document change.

## 19. Why is a visual overlay not enough for redaction?

Answer:

A visual overlay can hide text but may leave underlying content extractable. The prompt requires the underlying region to be permanently stripped after confirmation.

Why this matters:

- Legal redaction must protect actual content, not only appearance.
- The API contract must distinguish staged regions from confirmed burn-in.

## 20. Why create audit rows for every mutating action?

Answer:

The prompt requires structured audit trails with indefinite retention.

Why this matters:

- Legal teams need traceability.
- Reviewers can verify who changed what and when.
- Audit records make acceptance and debugging easier.

## 21. Why use in-app notifications instead of email?

Answer:

The environment is air-gapped and cannot rely on external notification delivery providers.

Why this matters:

- Notifications remain local.
- No SMTP, SaaS, Slack, SMS, or cloud service is required.
- Notification templates and messages can be stored in PostgreSQL.

## 22. Why add an informal frontend if the project says no frontend?

Answer:

The user later requested a non-formal frontend for testing. The project therefore includes an informal manual page only.

Why this matters:

- It helps human acceptance testers.
- It does not change the backend-first product scope.
- It avoids introducing frontend framework dependencies.

## 23. Why provide sample PDFs and CSV files?

Answer:

The user requested real testable mock data, including PDF and CSV data.

Why this matters:

- Acceptance can run offline.
- Upload and batch scenarios have local fixtures.
- Test scripts do not need internet downloads.

## 24. Why create `api-spec.md`?

Answer:

The user requested a dedicated interface description document.

Why this matters:

- README stays focused on startup and acceptance.
- API details have a stable location.
- Swaggo annotations and generated docs can be compared against the Markdown API spec.

## 25. Why create `design.md`?

Answer:

The user requested an explanation of why the system is designed this way.

Why this matters:

- It documents architectural tradeoffs.
- It explains why certain non-default decisions, such as single-container PostgreSQL plus API, are intentional.
- It helps future agents avoid changing core assumptions accidentally.

## 26. Why create `requirement-check.md`?

Answer:

The user requested a verification document comparing implementation with the prompt.

Why this matters:

- It makes completion status explicit.
- It prevents hiding partial implementation behind broad claims.
- It gives acceptance reviewers a checklist.

## 27. Why use the phrase “why this was done” instead of “fix”?

Answer:

The user asked not to use the word `fix` in `questions.md`, and to describe why decisions were made.

Why this matters:

- The document becomes a decision record rather than a bug log.
- It records reasoning and tradeoffs from the conversation.
- It avoids sounding like the project is only a list of corrections.

## 28. Why did API tests need seeded login bootstrap?

Answer:

The original API tests were not effective because they did not authenticate as real Admin, Editor, and Reviewer users before exercising protected routes. The user manually added seeded login bootstrap to `API_tests/test_api_flow.sh`.

Why this matters:

- RBAC cannot be validated without real tokens.
- Object-level authorization cannot be tested with anonymous requests.
- API coverage must start with a real authenticated flow.

## 29. Why remove SKIP-as-success behavior from API suites?

Answer:

A suite that prints SKIP and exits successfully can make incomplete coverage appear to pass. The admin and document review suites were changed to avoid silent success.

Why this matters:

- Acceptance output should not hide missing coverage.
- A missing test should fail or be explicitly documented as incomplete.
- This makes coverage gaps visible to reviewers.

## 30. Why add object-level authorization?

Answer:

Route-level RBAC is not enough for legal documents. A Reviewer or Editor should not automatically see every document just because they have a valid role.

Why this matters:

- Draft documents should remain restricted to their owner/editor.
- Reviewer access should start after the document leaves Draft.
- Legal document isolation requires per-object checks, not only role checks.

## 31. Why can Reviewer no longer read Draft documents?

Answer:

Draft documents are not yet ready for review. The updated object access rule allows Reviewers to read non-Draft documents only.

Why this matters:

- It preserves the workflow boundary between authoring and review.
- It prevents premature disclosure of draft legal material.
- It gives API tests a meaningful object-level authorization case.

## 32. Why does Bates now create a new document version?

Answer:

A Bates operation changes the legal artifact. Recording only job metadata is insufficient, so the Bates route now produces a new document version and updates `current_version`.

Why this matters:

- Bates output becomes traceable as a versioned artifact.
- The job can be audited against a resulting document version.
- Acceptance can verify version count changes after Bates.

## 33. Why is Bates still marked Partial?

Answer:

The current implementation creates a new PDF version using an append-style transform marker. It does not yet draw visible Bates numbers onto each PDF page or allocate cross-document batch sequences.

Why this matters:

- It is better than metadata-only but not full Bates numbering.
- The requirement expects page-visible numbering.
- The document must not claim a stronger implementation than exists.

## 34. Why was backup changed to create a local file?

Answer:

A backup endpoint that only inserts queued metadata cannot prove that a backup artifact exists. The enhanced handler writes a local backup marker file and records a Completed job.

Why this matters:

- The endpoint now has a filesystem side effect.
- Reviewers can inspect `target_path`.
- It is still not a real `pg_dump` or filesystem snapshot, so it remains Partial.

## 35. Why add audit query filters?

Answer:

The prompt requires audit logs to be queryable by meaningful fields. The enhanced audit route supports actor, document, action, request ID, source IP, and date range filters.

Why this matters:

- Legal audits need targeted searches.
- Reviewers can validate individual workflows without scanning all logs.
- It aligns audit output with request IDs and action types.

## 36. Why add annotation mention notification?

Answer:

The prompt requires notification behavior beyond workflow transitions. Annotation comments can contain `@username`; the mention helper resolves local users and creates in-app notifications.

Why this matters:

- It gives annotations a collaborative review signal.
- It remains offline and local.
- The notification path reuses the same unread-ceiling helper.

## 37. Why is redaction still not marked Complete?

Answer:

The current redaction transform still appends a marker to a PDF-like file. It does not permanently remove underlying content from the original PDF object streams.

Why this matters:

- A visual or append-only change is not forensic redaction.
- Legal redaction must make the hidden content unrecoverable.
- Without a proper PDF content removal engine, the project must honestly mark this as incomplete.

## 38. Why is compare still marked Partial?

Answer:

The compare endpoint now reads real version files and reports hash, size, page-count, and binary differences. It does not extract PDF text or return true page/bbox text segments.

Why this matters:

- It is no longer a fixed placeholder.
- It still does not satisfy full legal PDF comparison requirements.
- The requirement expects structured text-level added, removed, and modified segments.

## 39. Why update `requirement-check.md` after implementation changes?

Answer:

The static reviews found that code and documentation could drift. `requirement-check.md` must reflect actual implementation status rather than optimistic claims.

Why this matters:

- Reviewers depend on the document for acceptance status.
- Partial features should stay marked Partial or Planned.
- It prevents a mismatch between code behavior and written claims.

## 40. Why are more API tests still required?

Answer:

The current API tests cover login, some RBAC, document upload, object read checks, admin read endpoints, audit list, and backup job reads. They still do not cover the full mutating flow.

Why this matters:

- Workflow, finalize, redaction, annotation, Bates, compare, batch, notification read, backup run, logout, replay, and timestamp expiry need API assertions.
- Endpoint coverage is still below the requested threshold.
- Final acceptance depends on exercised behavior, not only implemented handlers.
