#!/usr/bin/env bash
set -euo pipefail

SWAG_BIN="${SWAG_BIN:-swag}"

mkdir -p docs/swagger
if [ ! -s docs/swagger/docs.go ]; then
  printf 'package swagger\n' > docs/swagger/docs.go
fi

"$SWAG_BIN" init \
  -g cmd/server/main.go \
  -o docs/swagger \
  --parseInternal \
  --parseDependency
