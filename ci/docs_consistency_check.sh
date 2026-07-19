#!/usr/bin/env bash
set -euo pipefail

python3 - <<'PY'
from pathlib import Path
import re

def stop(message):
    raise SystemExit("ERROR: " + message)

required = [
    "README.md", "ci/BOUNDARY.md", "docs/questions.md",
    "docs/requirement-check.md", "docs/security.md", "docs/design.md",
    "docs/backup-recovery.md", "docs/testing.md", "docs/usage.md",
    "docs/pitr.md", "docs/deployment-offline.md",
    "docs/swagger-artifacts.md",
    "skills/full-project-acceptance-hard-gates/SKILL.md",
    "skills/project-generation-workflow/SKILL.md",
]
for name in required:
    if not Path(name).is_file():
        stop(f"missing required file: {name}")
if Path("deploy/aws").exists() or Path("docs/aws-deployment.md").exists():
    stop("cloud deployment material conflicts with the air-gapped scope")
if not Path("public/index.html").is_file() or Path("public/manual-test.html").exists():
    stop("public/ must contain only public/index.html")
workflows = list(Path(".github/workflows").glob("*.y*ml"))
if len(workflows) != 1 or workflows[0].name != "ci.yml":
    stop("exactly one workflow at .github/workflows/ci.yml is required")

doc_paths = [Path("README.md"), Path("ci/BOUNDARY.md")]
doc_paths += [p for p in Path("docs").rglob("*") if p.is_file()]
docs = "\n".join(p.read_text(encoding="utf-8", errors="ignore") for p in doc_paths)
stale = {
    "legacy test path": r"API_tests/|unit_tests/",
    "duplicate UI path": r"/ui/manual-test\.html|public/manual-test\.html",
    "cloud deployment": r"AWS|EKS|Lambda|CloudFormation|serverless deployment",
    "obsolete guard documentation": r"ci_execution_guard\.py",
    "target-wide cooldown": r"ten-minute target cooldown|completed non-cancelled runs enforce a ten-minute target cooldown",
    "execution-gated verdict": r"Missing runtime or interaction evidence is `NOT VERIFIED`|Missing runtime, deployment, interaction, or full-regression evidence is recorded as `NOT VERIFIED`|Full acceptance requires a pre-existing generated",
    "workflow overstatement": r"static workflow.*complete regression|sole workflow.*complete regression|workflow.*uploads evidence only after the complete regression",
}
for label, pattern in stale.items():
    if re.search(pattern, docs, re.IGNORECASE):
        stop(f"stale documentation claim: {label}")

questions = Path("docs/questions.md").read_text(encoding="utf-8")
sections = list(re.finditer(r"^## (.+)$", questions, re.MULTILINE))
topics = {m.group(1).strip() for m in sections}
expected_topics = {
    "Rolling failed-login lockout",
    "Initial administrator and acceptance fixtures",
    "Authentication state failures must fail closed",
    "Acceptance browser surface",
    "Static reviewer acceptance",
    "CI admission and one-time unlock",
    "Regression and current-revision evidence",
}
if topics != expected_topics:
    stop(f"clarification topics differ: {sorted(topics)}")
expected_subsections = [
    "Easy-to-make interpretation", "Why it fails",
    "Correct requirement interpretation", "Required implementation",
    "Acceptance evidence",
]
for index, match in enumerate(sections):
    end = sections[index + 1].start() if index + 1 < len(sections) else len(questions)
    actual = re.findall(r"^### (.+)$", questions[match.end():end], re.MULTILINE)
    if actual != expected_subsections:
        stop(f"invalid clarification structure for {match.group(1)!r}")

acceptance = Path("skills/full-project-acceptance-hard-gates/SKILL.md").read_text(encoding="utf-8")
generation = Path("skills/project-generation-workflow/SKILL.md").read_text(encoding="utf-8")
workflow = Path(".github/workflows/ci.yml").read_text(encoding="utf-8")
gates = {int(v) for v in re.findall(r"^## Gate (\d+):", acceptance, re.MULTILINE)}
if gates != set(range(28)):
    stop(f"acceptance gates differ from 0-27: {sorted(gates)}")
for phrase in [
    "Absolute static-only boundary",
    "Every applicable Gate 0–27 must be completed",
    "Inspect workflows only; never trigger or wait for CI",
    "a new revision must be admissible immediately",
    "Missing test execution, CI, build, deployment, runtime logs, screenshots, or external artifacts does not alter the verdict",
]:
    if phrase not in acceptance:
        stop(f"acceptance Skill missing rule: {phrase}")
for phrase in [
    "static source-completion tasks",
    "must not optimize for the smallest change count",
    "Do not stop scanning after the first P0",
    "Continue until no known in-scope static defect is deferred",
    "CI triggered or awaited: `none`",
]:
    if phrase not in generation:
        stop(f"generation Skill missing rule: {phrase}")
for phrase in [
    "cancel-in-progress: true", "github.paginate", "failedSameRevision",
    "latestCompletedSameRevision", "run.head_sha === currentSha",
    "same-revision admission cooldown", "alreadyConsumed",
    "push a new revision or authorize that exact run once",
]:
    if phrase not in workflow:
        stop(f"workflow missing admission rule: {phrase}")
if workflow.index("actions/github-script@v7") >= workflow.index("actions/checkout@v4"):
    stop("admission must precede checkout")
if re.search(r"time\.sleep|\bsleep\s+\d+", workflow):
    stop("admission must reject rather than sleep")

credential = re.compile(r'''(?i)\b[A-Za-z0-9_]*(?:password|passphrase|secret|token|api_key|signing_key|encryption_key)[A-Za-z0-9_]*\b\s*[:=]\s*["']([^"']+)["']''')
allowed = ("${", "$", "<", "placeholder", "example", "sample", "required", "generated", "random", "redacted", "***")
for path in doc_paths + [Path("public/index.html"), Path("docker-compose.yml")]:
    for number, line in enumerate(path.read_text(encoding="utf-8", errors="ignore").splitlines(), 1):
        for match in credential.finditer(line):
            value = match.group(1).strip().lower()
            if value and not any(marker in value for marker in allowed):
                stop(f"fixed credential-like value at {path}:{number}")

print("PASS: documentation consistency checks")
PY
