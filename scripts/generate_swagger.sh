#!/usr/bin/env bash
set -euo pipefail

SWAG_BIN="${SWAG_BIN:-swag}"
"$SWAG_BIN" init \
  -g cmd/server/main.go \
  -o docs/swagger \
  --parseInternal \
  --parseDependency
