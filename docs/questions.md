# Questions and Implementation Answers

This document records practical project questions and implementation answers. It focuses on what was built, what tradeoff was made, what remains incomplete, and how the result should be reviewed.

## Q1. What is the real project scope?

IronPage Vault is a backend-first legal PDF lifecycle API. The deliverable is a local Go service with PostgreSQL metadata, filesystem PDF storage, workflow, versioning, annotations, redaction metadata, Bates processing records, audit logs, notifications, and backup records.

The manual HTML page under `public/` is only a backend testing aid. It is not a production frontend and should not change the acceptance scope.

## Q2. What does the Docker-based delivery model mean?

The intended build path is Docker builder. The service should be evaluated through the Docker build and compose path instead of a developer's local Go installation. This keeps acceptance isolated from local machine state.

The project still needs runtime verification through Docker. The code was edited but commands were not executed during these changes.

## Q3. What is the difference between role access and document object access?

Role access decides whether a role can call a category of endpoint. Object access decides whether that user can access that particular document.

The implementation now separates these concerns. Admin has broad read access, document owners can read their own records, Reviewers can read non-Draft documents, Editors can mutate owned non-Finalized documents, and Reviewers can review non-Draft non-Finalized documents.

The remaining task is to keep testing every mutation route with both allowed and denied cases.

## Q4. How should the document lifecycle be reviewed?

The lifecycle is:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

The implementation enforces ordered transitions and treats Finalized as the immutable end state. Reviewers should test the happy path, invalid transitions, finalization, and post-finalization rejection across mutating APIs.

## Q5. What is now complete about batch upload?

Batch upload now uses the same persistence path as single upload. Each file creates a document, a version, a stored PDF file, and audit output.

It is no longer an accepted-count placeholder. The missing part is API coverage: tests should upload multiple sample files and verify the returned count and resulting document records.

## Q6. What is the current encryption status?

The code includes AES-GCM helpers and unit tests. Annotation comments and redaction reasons are encrypted before storage.

This is partial column-level protection, not full sensitive-field encryption. Coordinates remain numeric database columns, so the feature should not be marked complete if coordinates are considered sensitive.

## Q7. What is still missing from redaction?

The redaction flow can stage regions, confirm a redaction, create a new version, and record audit.

It still does not perform forensic content removal. The current PDF marker helper only appends metadata-like marker bytes to a copied PDF file; it does not rewrite content streams, remove text, or prove unrecoverability.

Final acceptance requires a real PDF content removal engine and tests proving the removed content cannot be extracted.

## Q8. What does Bates processing currently provide?

Bates processing now creates a new document version and records the job. That is an improvement over metadata-only behavior.

It still does not draw visible Bates numbers on each PDF page and does not allocate cross-document batch sequences. The current PDF marker helper only appends marker bytes to the copied PDF file. The correct status is Partial until page-visible numbering exists.

## Q9. What does document comparison currently provide?

Comparison now reads real version files and reports metadata and binary-level differences. It also checks that both versions are within the caller's readable document scope.

It does not perform text extraction, real added/removed segment detection, page-level text comparison, or bounding-box reporting. It should remain Partial.

## Q10. What changed in audit logging?

Audit writes are no longer only direct per-handler inserts. A shared audit helper exists, and the audit list route has a filtered implementation for actor, document, action, request, source, and time-range review.

The audit list route must remain Admin-only. Non-Admin users should receive 403 when calling `/api/audit-logs`.

The feature still needs API tests that create known events and then verify the filters return the expected records.

## Q11. What changed in notifications?

Notifications remain local in PostgreSQL. Workflow updates create notifications. The helper also enforces an unread ceiling.

Annotation mention support was added through local username parsing. When a comment mentions another local user, the system can create an in-app notification for that user.

The remaining work is API validation of mention creation and read acknowledgement.

## Q12. What does backup currently do?

The backup endpoint now creates a local backup artifact marker and records a completed backup job. This is better than only inserting queued metadata.

It is still not a full backup implementation. Final acceptance requires a real database dump, a filesystem artifact or snapshot for stored PDFs, and a restore-oriented verification path.

## Q13. What API coverage is still missing?

Current API tests cover only part of the system. The remaining high-value coverage includes batch upload, rollback, workflow transitions, invalid transitions, finalization, redaction, annotation, Bates version creation, comparison, audit filters, notification read acknowledgement, backup execution, and finalization immutability.

The important point is that a missing flow should not appear as a passing suite.

## Q14. How should documentation stay honest?

`requirement-check.md` should describe actual status, not intended status. A handler that only creates a marker or partial artifact must not be described as complete.

Current honest statuses include:

- Redaction: incomplete for forensic removal.
- Bates: partial because version creation exists but visible page numbering does not.
- Backup: partial because a local artifact exists but full dump and snapshot do not.
- Compare: partial because it is not text and bounding-box aware.
- API coverage: partial because many mutating routes are not fully exercised.

## Q15. What kind of tests are still needed?

Helper-level unit tests are useful but insufficient. The project needs handler or API integration tests that assert database and filesystem side effects.

Important side effects include document version creation, audit row creation, notification creation, object-access denial, backup job creation, workflow history creation, and Finalized immutability.

## Q16. What is the current delivery status?

The current delivery is stronger than a skeleton but still not final-complete.

A fair label is:

```text
Partial backend implementation with improved persistence, access control, audit, notifications, and tests, but still missing compliance-grade PDF processing and sufficient E2E coverage.
```

## Q17. Why move rules and access policy out of internal/app?

`internal/app` is the API adapter layer. It should own Echo routes, middleware, request binding, and response mapping. It should not own the domain policy that decides whether a role may mutate a document, whether a workflow state is final, or whether a particular principal can access a particular document.

Moving pure rules and object-access policy into `internal/core` makes those decisions deterministic, framework-free, and testable without Echo or a database.

## Q18. Why keep app wrapper functions temporarily?

The migration is intentionally incremental. Existing handlers already call functions such as `canReadDocumentObject` and `IsValidBatesPadding` from the app package. Rewriting every handler in the same PR would create a large risky diff.

The temporary wrapper strategy is:

```text
handler -> internal/app compatibility wrapper -> internal/core policy
```

The next cleanup step is:

```text
handler/service -> internal/core policy
```

After callers are moved, the app wrappers can be deleted.

## Q19. Why does documentListWhereClause move to internal/store?

A SQL WHERE clause is not domain policy. It is persistence/query adapter behavior. It includes database column names and SQL placeholder syntax, so `internal/core` should not emit it.

It also should not remain owned by `internal/app`, because API handlers should not own query construction. The real implementation now belongs in `internal/store` as `DocumentListWhereClause`.

The app package keeps a temporary wrapper only to preserve existing handler calls while larger repository functions are extracted later.

## Q20. Why move crypto and digest helpers into internal/platform?

Crypto and digest helpers are implementation adapters. They provide AES-GCM encryption and SHA-256 digest capabilities, but they do not decide domain policy and they do not map HTTP requests.

Keeping them inside `internal/app` made the API package look like it owned encryption and file hashing. Moving the real implementations to `internal/platform` separates low-level infrastructure capability from API handling.

## Q21. Why keep app crypto/digest wrappers temporarily?

Existing code still calls `encryptString`, `decryptString`, and `fileDigest` from the app package. Rewriting all callers in the same PR would mix infrastructure migration with handler/service changes.

The temporary wrapper strategy is:

```text
app caller -> internal/app compatibility wrapper -> internal/platform implementation
```

The follow-up cleanup is:

```text
app/service caller -> internal/platform implementation
```

Then the app wrappers can be deleted.

## Q22. Why move PDF helpers into internal/platform?

PDF inspection and PDF marker helpers are filesystem/PDF adapter code. They read files, validate PDF headers, count page markers, compute file digests, and write copied PDF artifacts.

That is not API handler responsibility and not domain policy. The real implementation belongs in `internal/platform`.

The temporary compatibility path is:

```text
app caller -> internal/app PDF wrapper -> internal/platform PDF implementation
```

The cleanup path is:

```text
app/service caller -> internal/platform PDF implementation
```

After callers move, the app wrapper can be deleted.

## Q23. Why move workflow chain rules into internal/core?

The workflow chain is pure domain policy:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

It does not need Echo, SQL, filesystem access, or request/response mapping. Moving `NextWorkflowStatus` and `WorkflowStatusChain` into `internal/core` makes the lifecycle rule testable without API handlers.

The temporary wrapper strategy is:

```text
handler -> internal/app compatibility wrapper -> internal/core workflow rule
```

The follow-up cleanup is:

```text
handler/service -> internal/core workflow rule
```

## Q24. Why move notification unread-cap policy into internal/core?

The rule that decides how many unread notifications must be trimmed is deterministic domain policy:

```text
trim = unread - limit + 1 when unread >= limit
```

It does not need Echo, SQL, or HTTP response formatting. It belongs in `internal/core`.

The database update that marks rows as read remains outside core because SQL and persistence side effects belong in the app/store migration path.

## Q25. Why move mention parsing into internal/core?

Mention parsing is pure text policy. It takes an annotation comment and returns candidate usernames. It does not need Echo, SQL, or notification persistence.

The database lookup for users and the creation of notification rows remain outside core. Those side effects should later move through service/store boundaries.

The temporary wrapper strategy is:

```text
annotation handler -> internal/app wrapper -> internal/core text parser
```

The follow-up cleanup is:

```text
annotation service -> internal/core text parser
```
