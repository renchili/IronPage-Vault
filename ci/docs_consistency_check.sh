#!/usr/bin/env bash
set -euo pipefail

fail=0

check_absent() {
  local description="$1"
  local pattern="$2"
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
)

for path in "${required_docs[@]}"; do
  if [ ! -f "$path" ]; then
    echo "ERROR: required documentation file is missing: $path"
    fail=1
  fi
done

check_absent "stale test UI path; use /ui/" "/ui/manual-test\\.html|public/manual-test\\.html" README.md docs
check_absent "stale redaction behavior" "marker-only|marker only|does not perform forensic content removal" README.md docs
check_absent "stale Bates behavior" "does not draw visible Bates|no visible Bates|Bates numbers are not visible" README.md docs
check_absent "stale comparison behavior" "binary-only|does not perform text extraction|no bounding-box reporting" README.md docs
check_absent "stale backup behavior" "metadata-only backup|metadata snapshot is not a restore-capable backup|future worker.*backup" README.md docs
check_absent "obsolete security blocker" "unsafe runtime defaults.*(block|fail)|security-accepted until those defaults" docs/questions.md docs/requirement-check.md
check_absent "obsolete seeded-user guidance" "seeded users.*changed.*non-demo|seed users.*change.*deployment" docs/security.md
check_absent "untracked design process narration" "being refactored|temporarily keep|temporary app wrappers|follow-up PR|move callers directly.*later|extracted later" docs/design.md
check_absent "false no-limitations claim" "No product-scope recheck limitations|No limitations are currently tracked" docs/requirement-check.md
check_absent "work-log or next-action pollution" "Next action|Future work|Conversation record|agent process|tool failure|branch failure|PR failure" docs/questions.md

python3 - <<'PY' || fail=1
from pathlib import Path
import re
import sys

questions = Path("docs/questions.md").read_text(encoding="utf-8")
if not questions.startswith("# Requirement Clarifications\n"):
    print("ERROR: docs/questions.md must use the Requirement Clarifications title")
    raise SystemExit(1)

section_matches = list(re.finditer(r"^## (.+)$", questions, re.MULTILINE))
required_topics = {
    "Rolling failed-login lockout",
    "Initial administrator and acceptance fixtures",
    "Authentication state failures must fail closed",
    "Acceptance browser surface",
    "Regression and current-HEAD evidence",
}
actual_topics = {match.group(1).strip() for match in section_matches}
missing_topics = sorted(required_topics - actual_topics)
if missing_topics:
    print("ERROR: docs/questions.md is missing required clarification topics: " + ", ".join(missing_topics))
    raise SystemExit(1)

required_subsections = [
    "Easy-to-make interpretation",
    "Why it fails",
    "Correct requirement interpretation",
    "Required implementation",
    "Acceptance evidence",
]
for index, match in enumerate(section_matches):
    start = match.end()
    end = section_matches[index + 1].start() if index + 1 < len(section_matches) else len(questions)
    body = questions[start:end]
    headings = re.findall(r"^### (.+)$", body, re.MULTILINE)
    if headings != required_subsections:
        print(f"ERROR: clarification topic {match.group(1)!r} has invalid subsection order: {headings!r}")
        raise SystemExit(1)

required_paths = [
    Path("migrations/002_login_attempt_window.sql"),
    Path("API_tests/test_auth_lockout_docker.sh"),
    Path("ci/docker_acceptance.sh"),
]
for path in required_paths:
    if not path.is_file():
        print(f"ERROR: documented authentication evidence path is missing: {path}")
        raise SystemExit(1)

print("PASS: requirement clarification structure and evidence paths")
PY

python3 - <<'PY' || fail=1
from pathlib import Path
import re
import sys

roots = [Path("README.md"), Path("docs"), Path("public"), Path("docker-compose.yml")]
extensions = {".md", ".html", ".htm", ".js", ".json", ".yml", ".yaml"}
files = []
for root in roots:
    if root.is_file():
        files.append(root)
    elif root.is_dir():
        files.extend(path for path in root.rglob("*") if path.is_file() and path.suffix.lower() in extensions)

sensitive_name = r"(?:password|passphrase|secret|token|api_key|signing_key|encryption_key)"
quoted_assignment = re.compile(
    rf"(?i)\\b[A-Za-z0-9_]*{sensitive_name}[A-Za-z0-9_]*\\b\\s*[:=]\\s*[\\\"']([^\\\"']+)[\\\"']"
)
env_assignment = re.compile(
    r"^\\s*([A-Z0-9_]*(?:PASSWORD|PASSPHRASE|SECRET|TOKEN|API_KEY|SIGNING_KEY|ENCRYPTION_KEY)[A-Z0-9_]*)\\s*[:=]\\s*(.+?)\\s*$"
)
password_input = re.compile(
    r"(?i)<input(?=[^>]*type=[\"']password[\"'])(?=[^>]*value=[\"']([^\"']+)[\"'])[^>]*>"
)

placeholder_fragments = (
    "${",
    "$",
    "<",
    "placeholder",
    "example",
    "sample",
    "required",
    "from_env",
    "from-environment",
    "change_me",
    "change-me",
    "redacted",
    "***",
    "document.getelementbyid",
    "process.env",
    "os.getenv",
)

def is_placeholder(value: str) -> bool:
    normalized = value.strip().strip("`\\\"'").lower()
    if not normalized:
        return True
    if normalized in {"null", "none", "unset", "empty", "n/a", "na", "true", "false"}:
        return True
    return any(fragment in normalized for fragment in placeholder_fragments)

findings = []
for path in sorted(set(files)):
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except UnicodeDecodeError:
        continue

    credential_columns = None
    for line_number, line in enumerate(lines, start=1):
        for match in quoted_assignment.finditer(line):
            if not is_placeholder(match.group(1)):
                findings.append((path, line_number, "credential-like quoted assignment contains a fixed literal"))

        env_match = env_assignment.match(line)
        if env_match and not is_placeholder(env_match.group(2)):
            findings.append((path, line_number, "sensitive environment assignment contains a fixed literal"))

        input_match = password_input.search(line)
        if input_match and not is_placeholder(input_match.group(1)):
            findings.append((path, line_number, "password input contains a fixed value"))

        if line.count("|") >= 2:
            cells = [cell.strip() for cell in line.strip().strip("|").split("|")]
            normalized = [cell.lower().replace(" ", "_") for cell in cells]
            indexes = [i for i, cell in enumerate(normalized) if re.fullmatch(sensitive_name, cell)]
            if indexes:
                credential_columns = indexes
                continue
            if credential_columns and not all(re.fullmatch(r":?-+:?", cell) for cell in cells):
                for index in credential_columns:
                    if index < len(cells) and not is_placeholder(cells[index]):
                        findings.append((path, line_number, "credential table contains a fixed literal"))
        elif not line.strip():
            credential_columns = None

if findings:
    for path, line_number, message in findings:
        print(f"ERROR: {path}:{line_number}: {message}")
    raise SystemExit(1)

print("PASS: no fixed credential-like values found in documentation, browser assets, or Compose")
PY

if [ "$fail" -ne 0 ]; then
  exit 1
fi

echo "PASS: documentation consistency checks"
