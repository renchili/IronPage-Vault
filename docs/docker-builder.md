# Docker Builder Workflow

IronPage Vault is intended to be built through Docker rather than a local Go toolchain.

## Why

The project is a pure backend service for offline acceptance. Docker keeps the build environment isolated and repeatable.

## Build path

The Dockerfile uses a two-stage build:

1. Go builder stage downloads modules and compiles the backend binary.
2. Runtime stage packages PostgreSQL, the API binary, migrations, public test UI assets, and startup script.

## go.sum policy

This delivery does not require a committed `go.sum` file. Module resolution occurs inside the Docker builder stage during image build.

## Acceptance helper

A Docker-only acceptance helper exists at:

```text
scripts/docker_acceptance.sh
```

It is intended for evaluators and was not executed during generation.

## Local environment policy

Do not rely on local absolute paths, local Go versions, or globally installed Go packages. Use Docker for build and acceptance.
