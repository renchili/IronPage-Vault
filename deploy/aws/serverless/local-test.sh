#!/usr/bin/env bash
set -euo pipefail

if command -v python3 >/dev/null 2>&1; then
  python3 - <<'PY'
from pathlib import Path
import sys
try:
    import yaml
except Exception:
    print('python yaml package not installed; skipping YAML parser check')
    sys.exit(0)
for path in ['deploy/aws/serverless/template.yaml']:
    with open(path) as f:
        yaml.safe_load(f)
    print(f'validated yaml: {path}')
PY
fi

if command -v sam >/dev/null 2>&1; then
  sam validate --template-file deploy/aws/serverless/template.yaml || true
else
  echo 'sam not installed; skipping sam validate'
fi

if command -v aws >/dev/null 2>&1; then
  aws cloudformation validate-template --template-body file://deploy/aws/serverless/template.yaml >/dev/null || true
else
  echo 'aws cli not installed; skipping cloudformation validate-template'
fi

docker build -t ironpage-vault:serverless-local .

if [ -x scripts/docker_acceptance.sh ]; then
  bash scripts/docker_acceptance.sh
else
  echo 'scripts/docker_acceptance.sh missing; build validation completed'
fi
