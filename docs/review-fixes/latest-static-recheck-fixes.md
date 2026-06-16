# Latest Static Recheck Fixes

This patch addresses the reject items from `ironpage_latest_static_recheck_report`.

## Fixed test blockers

- `test_admin_ops.sh` now expects strict restore empty body to return 400.
- `test_admin_ops.sh` validates backup response fields and artifact file existence.
- `test_admin_ops.sh` performs a restore call using real artifacts returned by backup.
- `test_compare_acceptance.sh` delegates to the self-contained compare flow and no longer requires external version IDs.
- `test_finalized_immutability.sh` creates its own document and walks Draft -> Under Review -> Redaction Pending -> Approved -> Finalized before asserting immutability.

## Added content and side-effect checks

- `test_pdf_content_acceptance.sh` generates a PDF containing known text, confirms redaction, downloads the resulting PDF, and verifies `pdftotext` no longer extracts the target text.
- The same content test applies Bates and verifies the expected Bates label can be extracted.
- `test_notification_mention_side_effect.sh` creates an annotation with `@editor` and verifies a notification side effect is visible to the mentioned user.
- Audit action filtering now verifies returned rows match the requested action.
