# Test Effectiveness Follow-up

The uploaded review correctly identified that the earlier API tests were too weak.

Changes made:

- `API_tests/test_api_flow.sh` no longer prints static coverage claims.
- `API_tests/lib.sh` provides reusable request helpers.
- `API_tests/test_auth_rbac.sh` now contains real authenticated RBAC checks when tokens are supplied.
- `API_tests/test_document_upload.sh` now checks reviewer upload denial, editor upload, document ID extraction, document read, and version list.
- `public/manual-test.html` now calls backend APIs directly instead of only describing test steps.
- `/swagger/swagger.yaml` is now served from public assets so it is available in the container without copying the full docs tree.

Remaining manual step:

- Finish the seeded login bootstrap in `API_tests/test_api_flow.sh` so the script creates Admin, Editor, and Reviewer tokens before running sub-suites.

No commands were executed while making these changes.
