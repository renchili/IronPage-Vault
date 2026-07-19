# CI Boundary

IronPage Vault has one GitHub Actions workflow: `.github/workflows/ci.yml`.

## Static workflow scope

The workflow performs repository static acceptance only. It does not run project tests, full regression, Docker, databases, API interactions, browser interactions, or deployment.

Runtime and interaction evidence remain separate manual or lifecycle artifacts. A static workflow conclusion must never be presented as runtime acceptance.

## Target admission

All pull request, merge-group, `main` push, and manual events resolve to one explicit target key. The same target key is used by the workflow concurrency group and by admission history matching.

The workflow uses:

- `cancel-in-progress: true` to collapse superseded running or queued events for the same target;
- an admission job before checkout or repository-controlled code;
- complete Actions-history pagination rather than a first-page scan;
- immediate cancellation when admission is denied, never a sleep inside the runner;
- a ten-minute target cooldown based on the latest completed non-cancelled run;
- a failed target/revision latch;
- rejection of ordinary GitHub rerun attempts;
- a one-time manual unlock bound to the exact target, current revision, failed run ID, and reviewed reason;
- the run-name marker `unlock-<failed-run-id>` as the auditable consumption record.

A second manual dispatch using the same failed-run authorization is rejected.

## Platform limitation

Repository workflow YAML is evaluated only after GitHub creates a workflow-run object. The repository can prevent checkout and repository-controlled validation before admission succeeds, collapse active duplicates through concurrency, and cancel denied work. It cannot by itself prove that GitHub never creates a run object or allocates an admission runner.

A requirement for true pre-dispatch prevention needs separate repository ruleset, GitHub App, or external admission evidence. Static reports must state that limitation instead of claiming repository YAML provides pre-dispatch blocking.

## Static gates

After admission, one sequential job checks:

- workflow syntax;
- shell syntax;
- Python syntax;
- Go formatting;
- tracked source inventory and path hygiene;
- documentation consistency;
- repository and structure contracts;
- scheduled-backup contracts;
- metadata-storage contracts;
- Swagger route coverage.

Normal step failure prevents later steps from starting. There is no `if: always()` post-failure path.

## Retained evidence

`ci/source_inventory.py` writes:

```text
artifacts/static-acceptance/source-inventory.json
```

The manifest records the checked commit, every tracked path, mode, size, SHA-256, contamination findings, path-hygiene findings, and explicit naming exceptions. It is uploaded only after every static gate succeeds and is retained for 90 days.

## Runtime verification boundary

The following remain manual or normal-lifecycle entrypoints and are not called by the static workflow:

- `run_tests.sh`
- `ci/run_full_regression.sh`
- `ci/docker_acceptance.sh`
- `ci/run_project_api_regression.sh`
- stateful scripts under `tests/api/`

Their existence is static evidence only. A runtime claim requires a pre-existing completed artifact tied to the exact inspected revision.
