# Final Report Hardening

This patch addresses the remaining static recheck findings after strict fallback semantics.

## Coordinate encryption

Redaction proposal requests still accept geometry, but the service no longer persists those geometry values in plaintext. The legacy numeric columns are written as zero placeholders for compatibility while encrypted coordinate columns are the source of truth. Burn-in decrypts coordinates only inside the confirmation path.

## API exposure

Redaction list responses no longer expose plaintext coordinate or reason fields.

## CI and validation

The repository now includes a GitHub Actions workflow that runs `go test ./...`, static rule checks, and Docker build on pull requests and main pushes.

## Static rules

`unit_tests/test_rules.sh` now guards against reintroducing plaintext redaction coordinate persistence, non-strict PDF/backup paths, missing CI, missing self-contained compare coverage, and non-strict restore paths.
