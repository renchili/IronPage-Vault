# API Tests

This directory contains API functional acceptance material.

The API test suite is expected to cover:

- health endpoint
- local login
- authenticated principal lookup
- role denial paths
- Admin user and configuration endpoints
- Editor document upload endpoints
- Reviewer annotation endpoints
- workflow transitions
- finalized document immutability
- audit log query
- notification query
- backup job metadata

Run from the repository root with:

```bash
./run_tests.sh
```

The scripts assume the service is already available at `BASE_URL`, defaulting to:

```text
http://localhost:8080
```
