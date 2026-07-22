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
    "docs/pitr.md", "docs/deployment-offline.md", "docs/api-spec.md",
    "docs/metadata-security.md", "docs/swagger-artifacts.md",
    "skills/full-project-acceptance-hard-gates/SKILL.md",
    "skills/project-generation-workflow/SKILL.md",
]
for name in required:
    if not Path(name).is_file():
        stop(f"missing required file: {name}")
for obsolete in ["docs/implementation-status.md", "docs/test-effectiveness-followup.md"]:
    if Path(obsolete).exists():
        stop(f"obsolete process/status document remains: {obsolete}")
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
    "execution-gated static verdict": r"complete acceptance result requires executed evidence|Missing runtime or interaction evidence is `NOT VERIFIED`|Full acceptance requires a pre-existing generated",
    "workflow overstatement": r"static workflow.*complete regression|sole workflow.*complete regression|workflow.*uploads evidence only after the complete regression",
    "best-effort backup": r"best-effort.*(?:pg_dump|backup)|metadata-only backup",
    "manual-only restore isolation": r"^\s*1\.\s*Stop application writes\.",
    "false interrupted failure": r"Requested.*(?:converts|changes).*Failed|Requested journal.*Failed",
    "overlay redaction": r"draw filled black rectangles|overlay-style redaction|marker-only redaction",
    "obsolete compare limitation": r"not true bbox-level|no bounding-box reporting|binary-only compare",
    "old local runner claim": r"run_tests\.sh directly runs go test",
    "acceptance process residue": r"acceptance fix bundle|implementation followup|test-effectiveness followup",
}
for label, pattern in stale.items():
    if re.search(pattern, docs, re.IGNORECASE | re.MULTILINE):
        stop(f"stale documentation claim: {label}")

questions = Path("docs/questions.md").read_text(encoding="utf-8")
sections = list(re.finditer(r"^## (.+)$", questions, re.MULTILINE))
topics = {m.group(1).strip() for m in sections}
expected_topics = {
    "Rolling failed-login lockout",
    "Initial administrator and acceptance fixtures",
    "Authentication state failures must fail closed",
    "Acceptance browser surface",
    "Buildable frontend design and interaction handoff",
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
    "exact icon library and icon name",
    "developers must still choose material icons",
    "visual-only definitions of special interactions",
    "UI implementation readiness and platform review",
]:
    if phrase not in acceptance:
        stop(f"acceptance Skill missing rule: {phrase}")
for phrase in [
    "static source-completion tasks",
    "must not optimize for the smallest change count",
    "Do not stop scanning after the first P0",
    "Continue until no known in-scope static defect is deferred",
    "CI triggered or awaited: `none`",
    "Frontend design and implementation contract",
    "exact icon library and icon name",
    "Special-interaction contract",
    "Platform and app-review compliance",
    "Do not invent YAML, JSON, schema, manifest, token-registry, or “review pack” deliverables",
]:
    if phrase not in generation:
        stop(f"generation Skill missing rule: {phrase}")
for phrase in [
    "cancel-in-progress: true", "github.paginate", "failedSameRevision",
    "latestCompletedSameRevision", "run.head_sha === currentSha",
    "same-revision admission cooldown", "alreadyConsumed",
    "sameRevision", "sameBranch", "sameRepository",
    "exact open PR revision", "must equal the selected branch",
]:
    if phrase not in workflow:
        stop(f"workflow missing admission rule: {phrase}")
if workflow.index("actions/github-script@v7") >= workflow.index("actions/checkout@v4"):
    stop("admission must precede checkout")
if re.search(r"time\.sleep|\bsleep\s+\d+", workflow):
    stop("admission must reject rather than sleep")

for path, phrases in {
    "README.md": [
        "PUT /api/admin/workflow-statuses", "same-repository open PR",
        "audit source IP and structured metadata", "pg_restore --single-transaction",
        "application recovery boundary", "Interrupted", "PGPASSFILE",
        "under review", "redaction pending", "approved",
    ],
    "docs/design.md": [
        "same transaction", "page-number range", "safe archive extraction",
        "operation coordination", "system principal", "Interrupted",
        "same-repository open PR", "Draft -> Under Review -> Redaction Pending -> Approved -> Finalized",
    ],
    "docs/requirement-check.md": [
        "admin get/put route", "deterministic source-ip lookup/backfill",
        "canonical manual target validation", "ordered validation",
        "backup recovery boundary", "restore maintenance", "password-free argv",
        "buildable frontend generation rules", "exact component, icon, size",
        "arbitrary yaml/json packages cannot substitute",
    ],
    "docs/testing.md": [
        "exact same-repository open PR", "audit source ip/metadata",
        "staged restore", "ordered definitions", "backup/restore integrity",
        "Requested becomes Interrupted/unknown",
    ],
    "docs/api-spec.md": [
        "put | `/api/admin/workflow-statuses`", "source_ip",
        "start_number", "end_number", "complete ordered chain",
        "/api/admin/backup/restore/:id/resolve", "CONFIG_KEY_READ_ONLY",
    ],
    "docs/backup-recovery.md": [
        "safe", "--single-transaction", "requested", "completed", "failed",
        "application mutation barrier", "Interrupted", "PGPASSFILE",
    ],
    "docs/security.md": [
        "Mandatory acting-user audit", "PGPASSFILE", "database passwords therefore do not appear in subprocess argv",
        "configuration integrity", "Interrupted",
    ],
    "docs/pitr.md": [
        "exclusive application mutation barrier", "code-enforced maintenance", "Interrupted",
    ],
}.items():
    text = Path(path).read_text(encoding="utf-8").lower()
    for phrase in phrases:
        if phrase.lower() not in text:
            stop(f"{path} missing current implementation claim: {phrase}")

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
