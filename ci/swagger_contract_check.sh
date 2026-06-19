#!/usr/bin/env bash
set -euo pipefail

test -s docs/swagger/docs.go
test -s docs/swagger/swagger.yaml

python3 - <<'PY'
import re
import sys
from pathlib import Path

source_routes = []
for path in sorted(Path('internal/app').glob('swagger_*.go')):
    text = path.read_text(encoding='utf-8')
    for match in re.finditer(r'@Router\s+(\S+)\s+\[(\w+)\]', text):
        route, method = match.groups()
        source_routes.append((route, method.lower(), str(path)))

if not source_routes:
    raise SystemExit('no @Router annotations found under internal/app/swagger_*.go')

swagger = Path('docs/swagger/swagger.yaml').read_text(encoding='utf-8')
paths = {}
current_path = None
for line in swagger.splitlines():
    path_match = re.match(r'^  (/[^:]+):\s*$', line)
    if path_match:
        current_path = path_match.group(1).strip('"\'')
        paths.setdefault(current_path, set())
        continue
    method_match = re.match(r'^    (get|post|put|patch|delete|options|head):\s*$', line)
    if current_path and method_match:
        paths[current_path].add(method_match.group(1))

missing = []
for route, method, source in source_routes:
    if method not in paths.get(route, set()):
        missing.append(f'{method.upper()} {route} from {source}')

if missing:
    print('Generated swagger is missing annotated routes:', file=sys.stderr)
    for item in missing:
        print(f'  - {item}', file=sys.stderr)
    sys.exit(1)

required_responses = {
    ('/api/admin/backup/restore', 'post'): {'200', '400'},
    ('/api/documents', 'get'): {'200', '401'},
    ('/api/documents', 'post'): {'201', '400', '401'},
    ('/api/audit-logs', 'get'): {'200', '401', '403'},
}

lines = swagger.splitlines()
for (route, method), codes in required_responses.items():
    try:
        route_index = next(i for i, line in enumerate(lines) if line == f'  {route}:')
    except StopIteration:
        raise SystemExit(f'missing required route section {route}')

    next_route = next((i for i in range(route_index + 1, len(lines)) if re.match(r'^  /[^:]+:\s*$', lines[i])), len(lines))
    route_block = lines[route_index:next_route]
    try:
        method_index = next(i for i, line in enumerate(route_block) if line == f'    {method}:')
    except StopIteration:
        raise SystemExit(f'missing method {method.upper()} for {route}')

    next_method = next((i for i in range(method_index + 1, len(route_block)) if re.match(r'^    (get|post|put|patch|delete|options|head):\s*$', route_block[i])), len(route_block))
    method_block = '\n'.join(route_block[method_index:next_method])
    for code in sorted(codes):
        if not re.search(rf'\n\s+"?{code}"?:', method_block):
            raise SystemExit(f'missing response {code} for {method.upper()} {route}')

if 'BearerAuth' not in swagger and 'ApiKeyAuth' not in swagger:
    raise SystemExit('generated swagger does not include an auth security scheme')

print(f'PASS generated Swagger contract: {len(source_routes)} annotated routes verified')
PY
