#!/usr/bin/env bash
set -euo pipefail

if [ -n "${SWAG_BIN:-}" ]; then
  swag_bin="$SWAG_BIN"
elif command -v swag >/dev/null 2>&1; then
  swag_bin="$(command -v swag)"
else
  swag_bin="$(go env GOPATH)/bin/swag"
fi

if [ ! -x "$swag_bin" ]; then
  go install github.com/swaggo/swag/cmd/swag@v1.16.4
fi

mkdir -p docs/swagger
if [ ! -s docs/swagger/docs.go ]; then
  printf 'package swagger\n' > docs/swagger/docs.go
fi

"$swag_bin" init \
  -g cmd/server/main.go \
  -o docs/swagger \
  --parseInternal \
  --parseDependency
