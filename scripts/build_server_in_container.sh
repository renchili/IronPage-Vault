#!/usr/bin/env sh
set -eu

mkdir -p docs/swagger
printf 'package swagger\n' > docs/swagger/docs.go
go mod tidy
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/server/main.go -o docs/swagger --parseInternal --parseDependency
go build -mod=mod -o /out/ironpage ./cmd/server
