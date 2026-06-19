#!/usr/bin/env bash
set -euo pipefail

bash scripts/generate_swagger.sh

test -s docs/swagger/docs.go
test -s docs/swagger/swagger.yaml

python3 - <<'PY'
import re
import sys
from pathlib import Path

source_contracts = []
for path in sorted(Path('internal/app').glob('swagger_*.go')):
    pending_codes = set()
    pending_start = None
    for lineno, line in enumerate(path.read_text(encoding='utf-8').splitlines(), 1):
        code_match = re.search(r'@(Success|Failure)\s+(\d{3})\b', line)
        if code_match:
            pending_codes.add(code_match.group(2))
            pending_start = pending_start or lineno
            continue
        router_match = re.search(r'@Router\s+(\S+)\s+\[(\w+)\]', line)
        if router_match:
            route, method = router_match.groups()
            source_contracts.append((route, method.lower(), set(pending_codes), f'{path}:{pending_start or lineno}'))
            pending_codes = set()
            pending_start = None

if not source_contracts:
    raise SystemExit('no router annotations found')

swagger_lines = Path('docs/swagger/swagger.yaml').read_text(encoding='utf-8').splitlines()
paths = {}
current_path = None
current_method = None
for line in swagger_lines:
    path_match = re.match(r'^\s{2}(["\']?/[^:"\']+["\']?):\s*$', line)
    if path_match:
        current_path = path_match.group(1).strip('"\'')
        paths.setdefault(current_path, {})
        current_method = None
        continue
    method_match = re.match(r'^\s{4}(get|post|put|patch|delete|options|head):\s*$', line)
    if current_path and method_match:
        current_method = method_match.group(1)
        paths[current_path].setdefault(current_method, set())
        continue
    code_match = re.match(r'^\s{8}["\']?(\d{3})["\']?:\s*$', line)
    if current_path and current_method and code_match:
        paths[current_path][current_method].add(code_match.group(1))

missing = []
for route, method, codes, source in source_contracts:
    generated_codes = paths.get(route, {}).get(method)
    if generated_codes is None:
        missing.append(f'missing route {method.upper()} {route} from {source}')
        continue
    for code in sorted(codes):
        if code not in generated_codes:
            missing.append(f'missing response {code} for {method.upper()} {route} from {source}')

if missing:
    print('Generated swagger does not match source annotations:', file=sys.stderr)
    for item in missing:
        print(f'  - {item}', file=sys.stderr)
    sys.exit(1)

joined = '\n'.join(swagger_lines)
if 'BearerAuth' not in joined and 'bearerAuth' not in joined and 'ApiKeyAuth' not in joined:
    raise SystemExit('generated swagger does not include an auth security scheme')

print(f'PASS generated Swagger contract: {len(source_contracts)} annotated routes verified')
PY
