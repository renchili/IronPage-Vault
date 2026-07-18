#!/usr/bin/env bash
set -euo pipefail

fail=0

check_absent() {
  local description="$1" pattern="$2"
  shift 2
  local matches
  matches=$(grep -RInE "$pattern" "$@" 2>/dev/null || true)
  if [ -n "$matches" ]; then
    echo "ERROR: $description"
    echo "$matches"
    fail=1
  fi
}

required_docs=(
  README.md
  docs/questions.md
  docs/requirement-check.md
  docs/security.md
  docs/design.md
  docs/backup-recovery.md
  docs/testing.md
  docs/usage.md
  docs/pitr.md
  docs/deployment-offline.md
)
for path in "${required_docs[@]}"; do
  if [ ! -f "$path" ]; then
    echo "ERROR: required documentation file is missing: $path"
    fail=1
  fi
done

check_absent "legacy test directory path" "API_tests/|unit_tests/" README.md docs ci tests run_tests.sh .github
check_absent "removed duplicate UI path" "/ui/manual-test\\.html|public/manual-test\\.html" README.md docs ci tests public
check_absent "stale redaction behavior" "marker-only|marker only|does not perform forensic content removal" README.md docs
check_absent "stale Bates behavior" "does not draw visible Bates|no visible Bates|Bates numbers are not visible" README.md docs
check_absent "stale comparison behavior" "binary-only|does not perform text extraction|no bounding-box reporting" README.md docs
check_absent "stale backup behavior" "metadata-only backup|metadata snapshot is not a restore-capable backup|future worker.*backup" README.md docs
check_absent "obsolete security blocker" "unsafe runtime defaults.*(block|fail)|security-accepted until those defaults" docs/questions.md docs/requirement-check.md
check_absent "untracked process narration" "Next action|Future work|Conversation record|agent process|tool failure|branch failure|PR failure|This patch addresses|This patch closes" README.md docs
check_absent "false current evidence claim" "current HEAD.*(passed|PASS)|full regression.*current.*passed" README.md docs
check_absent "cloud deployment outside project scope" "AWS|EKS|Lambda|CloudFormation|serverless deployment" README.md docs ci scripts tests

if [ -e deploy/aws ] || [ -e docs/aws-deployment.md ]; then
  echo "ERROR: cloud deployment material conflicts with the air-gapped single-container scope"
  fail=1
fi
if [ -d docs/review-fixes ]; then
  if find docs/review-fixes -type f -print -quit | grep -q .; then
    echo "ERROR: obsolete review-fix process documents remain under docs/review-fixes"
    fail=1
  fi
fi
if [ ! -f public/index.html ] || [ -e public/manual-test.html ]; then
  echo "ERROR: public/ must contain one canonical acceptance UI at public/index.html"
  fail=1
fi
if [ "$(find .github/workflows -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \) | wc -l | tr -d ' ')" != "1" ]; then
  echo "ERROR: exactly one GitHub Actions workflow is required"
  fail=1
fi

python3 - <<'PY' || fail=1
from pathlib import Path
import re

questions = Path('docs/questions.md').read_text(encoding='utf-8')
if not questions.startswith('# Requirement Clarifications\n'):
    raise SystemExit('ERROR: docs/questions.md must use the Requirement Clarifications title')
section_matches = list(re.finditer(r'^## (.+)$', questions, re.MULTILINE))
required_topics = {
    'Rolling failed-login lockout',
    'Initial administrator and acceptance fixtures',
    'Authentication state failures must fail closed',
    'Acceptance browser surface',
    'Regression and current-HEAD evidence',
}
actual_topics = {match.group(1).strip() for match in section_matches}
missing = sorted(required_topics - actual_topics)
if missing:
    raise SystemExit('ERROR: missing clarification topics: ' + ', '.join(missing))
required_subsections = [
    'Easy-to-make interpretation',
    'Why it fails',
    'Correct requirement interpretation',
    'Required implementation',
    'Acceptance evidence',
]
for index, match in enumerate(section_matches):
    start = match.end()
    end = section_matches[index + 1].start() if index + 1 < len(section_matches) else len(questions)
    headings = re.findall(r'^### (.+)$', questions[start:end], re.MULTILINE)
    if headings != required_subsections:
        raise SystemExit(f'ERROR: clarification topic {match.group(1)!r} has invalid subsection order: {headings!r}')
for path in (
    Path('migrations/002_login_attempt_window.sql'),
    Path('tests/api/test_auth_lockout_docker.sh'),
    Path('ci/docker_acceptance.sh'),
):
    if not path.is_file():
        raise SystemExit(f'ERROR: documented evidence path is missing: {path}')
print('PASS: requirement clarification structure and evidence paths')
PY

python3 - <<'PY' || fail=1
from pathlib import Path
import re

roots = [Path('README.md'), Path('docs'), Path('public'), Path('docker-compose.yml')]
extensions = {'.md', '.html', '.htm', '.js', '.json', '.yml', '.yaml'}
files = []
for root in roots:
    if root.is_file():
        files.append(root)
    elif root.is_dir():
        files.extend(path for path in root.rglob('*') if path.is_file() and path.suffix.lower() in extensions)

sensitive_name = r'(?:password|passphrase|secret|token|api_key|signing_key|encryption_key)'
quoted_assignment = re.compile(rf"(?i)\b[A-Za-z0-9_]*{sensitive_name}[A-Za-z0-9_]*\b\s*[:=]\s*[\"']([^\"']+)[\"']")
env_assignment = re.compile(r"^\s*([A-Z0-9_]*(?:PASSWORD|PASSPHRASE|SECRET|TOKEN|API_KEY|SIGNING_KEY|ENCRYPTION_KEY)[A-Z0-9_]*)\s*[:=]\s*(.+?)\s*$")
password_input = re.compile(r'''(?i)<input(?=[^>]*type=["']password["'])(?=[^>]*value=["']([^"']+)["'])[^>]*>''')
placeholder_fragments = (
    '${', '$', '<', 'placeholder', 'example', 'sample', 'required', 'generated',
    'random', 'from_env', 'change_me', 'redacted', '***',
    'document.getelementbyid', 'process.env', 'os.getenv',
)

def is_placeholder(value: str) -> bool:
    normalized = value.strip().strip('`"\'').lower()
    if not normalized or normalized in {'null', 'none', 'unset', 'empty', 'n/a', 'na', 'true', 'false'}:
        return True
    return any(fragment in normalized for fragment in placeholder_fragments)

findings = []
for path in sorted(set(files)):
    try:
        lines = path.read_text(encoding='utf-8').splitlines()
    except UnicodeDecodeError:
        continue
    for line_number, line in enumerate(lines, start=1):
        for match in quoted_assignment.finditer(line):
            if not is_placeholder(match.group(1)):
                findings.append((path, line_number, 'credential-like quoted assignment contains a fixed literal'))
        env_match = env_assignment.match(line)
        if env_match and not is_placeholder(env_match.group(2)):
            findings.append((path, line_number, 'sensitive environment assignment contains a fixed literal'))
        input_match = password_input.search(line)
        if input_match and not is_placeholder(input_match.group(1)):
            findings.append((path, line_number, 'password input contains a fixed value'))
if findings:
    for path, line_number, message in findings:
        print(f'ERROR: {path}:{line_number}: {message}')
    raise SystemExit(1)
print('PASS: no fixed credential-like values in documentation, browser assets, or Compose')
PY

if [ "$fail" -ne 0 ]; then
  exit 1
fi

echo "PASS: documentation consistency checks"
