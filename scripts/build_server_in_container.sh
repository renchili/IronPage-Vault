#!/usr/bin/env sh
set -eu

# The application blank-imports the generated Swagger package. Create a
# temporary declaration first so module resolution can complete inside this
# image, then regenerate the package with the same project script used by CI.
mkdir -p docs/swagger
printf 'package swagger\n' > docs/swagger/docs.go
go mod tidy
go install github.com/swaggo/swag/cmd/swag@v1.16.4
bash scripts/generate_swagger.sh
go build -mod=mod -o /out/ironpage ./cmd/server
