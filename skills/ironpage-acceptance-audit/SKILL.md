---
name: ironpage-acceptance-audit
description: Re-run the IronPage Vault prompt-completion acceptance audit against the current repository state, collect GitHub Actions/run_tests/full-regression evidence, and produce an honest pass/partial/fail report.
---

# IronPage Vault Acceptance Audit Skill

Use this skill when the task is to verify whether **IronPage Vault** currently satisfies the original prompt, to re-check prior audit findings against the latest code, or to generate an acceptance report with GitHub Actions evidence.

The goal is not to defend old conclusions. Always re-check the current repository state first.

## Repository

- Repository: `renchili/IronPage-Vault`
- Default branch: `main`
- Primary evidence sources:
  - `metadata.json`
  - `docs/requirement-check.md`
  - `docs/metadata-security.md`
  - `docs/role-field-visibility.md`
  - `ci/run_full_regression.sh`
  - `run_tests.sh`
  - `ci/run_tests_contract_check.sh`
  - latest PRs and workflow artifacts
  - generated regression summaries under `reports/regression/**/summary.md` and `summary.json`

## Non-negotiable rules

1. **Do not rely on prior audit conclusions.** The project changes quickly. Re-read the current files.
2. **Separate evidence types clearly:**
   - static code evidence;
   - PR CI probe evidence;
   - full regression evidence;
   - local `run_tests.sh` acceptance evidence;
   - manually generated report from the assistant.
3. **Do not confuse `run_tests.sh` with `ci/run_full_regression.sh`.**
   - `run_tests.sh` writes `artifacts/local-acceptance/**`.
   - `ci/run_full_regression.sh` writes `artifacts/regression/**` and may be copied into `reports/regression/**`.
4. **Do not claim a test ran unless a workflow/job/summary/log proves it.**
5. **Do not call a generated assistant report a test result.** It is only a review artifact.
6. **If a job is probe-only, say so.** `IRONPAGE_RUN_TESTS_CONTRACT_PROBE=1 bash run_tests.sh` verifies the local entrypoint/report-generation path, not the full API acceptance suite.
7. **If the result is missing, say what is missing and how to obtain it.**

## Audit workflow

### 1. Reconstruct the original prompt requirements

Read `metadata.json` and extract the acceptance requirements:

- air-gapped legal document lifecycle system;
- Go + Echo;
- sqlx + PostgreSQL;
- local filesystem PDF binary storage;
- Admin / Editor / Reviewer RBAC;
- Draft -> Under Review -> Redaction Pending -> Approved -> Finalized;
- immutable Finalized records;
- two-phase forensic redaction with burn-in;
- Bates numbering with prefix/suffix/padding/sequential batch;
- version history / rollback / revision ceiling;
- batch import limits;
- structured comparison with added/removed/modified text;
- bcrypt local auth;
- lockout after failed login attempts;
- JWT session / jti blacklist / replay protection;
- timestamp request freshness;
- AES-256 column-level protection for sensitive fields;
- contextual role-based masking;
- audit trails;
- notifications;
- configurable templates and workflow definitions;
- pagination and uniform error envelope;
- PDF limits;
- scheduled backup and restore/PITR docs;
- single Docker container / no external dependency posture.

### 2. Check security storage against current code

Read these files:

- `docs/metadata-security.md`
- `migrations/001_schema.sql`
- `internal/platform/crypto.go`
- `internal/app/credential_storage.go`
- `internal/app/pii_storage.go`
- `internal/app/auth.go`
- `internal/app/admin.go`
- `ci/metadata_storage_check.sh`

The expected current implementation should show:

- AES-GCM helper with `enc:v1:` prefix in `internal/platform/crypto.go`;
- `sealPasswordHash` and `openPasswordHash` wrapping bcrypt verifiers;
- deterministic lookup key for username lookup, e.g. `lookup:v1:`;
- ciphertext columns for username, display name, document title, notification message, audit source IP, and audit metadata;
- redaction geometry and reason protected;
- annotation comment protected;
- metadata contract checking these requirements.

Important: old conclusions that password hashes or PII are not sealed may be stale. Verify the current code.

### 3. Check role-based field visibility

Read:

- `docs/role-field-visibility.md`
- response types in `internal/app/types.go`
- admin/document/review/notification handlers

Confirm whether:

- password hashes are hidden from all roles;
- ciphertext companion columns are never serialized;
- notification messages are recipient-only;
- audit source IP and metadata are Admin-only;
- document titles are opened only after object-level access checks;
- redaction coordinates/reasons remain hidden from list responses.

### 4. Check `run_tests.sh` local acceptance report

Read:

- `run_tests.sh`
- `ci/run_tests_contract_check.sh`
- PRs touching local acceptance report behavior, especially PRs titled like `Generate visual local acceptance report` or `Upload local acceptance report artifact from PR CI`.

Expected `run_tests.sh` generated outputs:

```text
artifacts/local-acceptance/results.tsv
artifacts/local-acceptance/summary.json
artifacts/local-acceptance/summary.md
artifacts/local-acceptance/report.html
artifacts/local-acceptance/logs/*
```

Expected contract check behavior:

```bash
rm -rf docs/swagger artifacts/local-acceptance
IRONPAGE_RUN_TESTS_CONTRACT_PROBE=1 bash run_tests.sh

test -s docs/swagger/docs.go
test -s docs/swagger/swagger.yaml
test -s artifacts/local-acceptance/results.tsv
test -s artifacts/local-acceptance/summary.json
test -s artifacts/local-acceptance/summary.md
test -s artifacts/local-acceptance/report.html
grep -q 'IronPage Local Acceptance Report' artifacts/local-acceptance/report.html
grep -q 'local_entrypoint_contract' artifacts/local-acceptance/results.tsv
```

If a workflow artifact exists, download and inspect it. Typical artifact name:

```text
local-acceptance-<run_id>-<sha>.zip
```

When reporting the result, state whether the artifact is:

- probe-only evidence; or
- full `./run_tests.sh` acceptance suite evidence.

### 5. Check full regression evidence

Read:

- `ci/run_full_regression.sh`
- `.github/workflows/full-regression-reusable.yml`
- `.github/workflows/post-merge-regression.yml`
- latest generated summaries under `reports/regression/**/summary.md` and `summary.json`

Expected full regression stages include:

- `prepare_workspace`
- `swagger_contract`
- `swagger_route_coverage`
- `scheduled_backup_contract`
- `metadata_storage_contract`
- `gofmt`
- `go_vet`
- `go_test_race`
- `shell_syntax`
- `docker_build`
- `docker_acceptance`

A valid full-regression evidence summary should explicitly show `Overall: PASSED` or `overall_status: passed`, with all stage statuses successful.

### 6. Check simple UI/manual validation aid

Read:

- `public/manual-test.html`
- `internal/app/server.go`
- `Dockerfile`

Expected evidence:

- manual backend test UI exists at `public/manual-test.html`;
- server exposes `e.Static("/ui", cfg.PublicDir)`;
- Docker image copies `public` into runtime image;
- UI covers login, document upload/list, annotation, redaction, Bates, workflow, audit, notifications, backup, and Swagger YAML.

Report this as a manual validation aid, not a production frontend.

## Acceptance decision rubric

Use this rubric unless the user gives a stricter one:

### Pass / accepted

Use this when:

- current code satisfies the main original prompt requirements;
- current security storage/masking docs and code align;
- full regression has passed;
- `run_tests.sh` report generation has passed and artifact availability is understood;
- remaining issues are only clarity/follow-up items.

Recommended wording:

> Engineering acceptance: passed. Original prompt main requirements: satisfied by current code and evidence. Note: local `run_tests.sh` artifact may be probe-only unless a full local API acceptance workflow is explicitly run.

### Conditional pass

Use this when:

- implementation is mostly complete;
- but one of the required evidence sources is missing or probe-only;
- or a prompt requirement is documented but not test-backed.

### Partial / not accepted

Use this when:

- a core prompt feature is absent;
- full regression failed;
- security storage contradicts the prompt;
- or the report relies on stale conclusions rather than current code.

## Report structure

Generate the report with these sections:

1. **Scope and timestamp**
2. **Current repository/branch/PR evidence**
3. **Original prompt checklist**
4. **Security storage and PII matrix**
5. **Role-based response visibility**
6. **`run_tests.sh` evidence**
7. **Full regression evidence**
8. **Manual UI evidence**
9. **Remaining caveats**
10. **Final acceptance decision**

Be precise about artifact names, workflow run IDs, PR numbers, and whether each result is full-suite or probe-only.

## Common mistakes to avoid

- Saying password hashes are not AES sealed without re-reading current code.
- Treating old reports as current evidence.
- Confusing `reports/regression/**` with `artifacts/local-acceptance/**`.
- Treating `IRONPAGE_RUN_TESTS_CONTRACT_PROBE=1` as full local API acceptance.
- Saying a generated assistant report is a test artifact.
- Failing to download the PR CI artifact when it exists.
- Ignoring newly merged PRs after a previous audit.
