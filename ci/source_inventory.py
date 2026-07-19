#!/usr/bin/env python3
"""Generate a retained tracked-source manifest and reject repository hazards."""

from __future__ import annotations

from collections import defaultdict
import hashlib
import json
from pathlib import Path
import stat
import subprocess
import sys

FORBIDDEN_PARTS = {
    "__pycache__",
    ".pytest_cache",
    "node_modules",
    ".idea",
    ".vscode",
}
FORBIDDEN_SUFFIXES = {
    ".pyc",
    ".class",
    ".o",
    ".db",
    ".sqlite",
    ".log",
}
FORBIDDEN_NAMES = {
    ".env",
    ".DS_Store",
}
GENERATED_PREFIXES = (
    "artifacts/",
    "coverage/",
    "reports/",
)
ALLOWED_NEAR_DUPLICATE_PAIRS = {
    frozenset({"AGENT.md", "AGENTS.md"}),
}


def tracked_files() -> list[Path]:
    result = subprocess.run(
        ["git", "ls-files", "-z"],
        check=True,
        stdout=subprocess.PIPE,
    )
    return sorted(
        (Path(item.decode("utf-8")) for item in result.stdout.split(b"\0") if item),
        key=lambda path: path.as_posix(),
    )


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def differs_by_one_edit(left: str, right: str) -> bool:
    """Return true only when two names differ by exactly one edit."""
    if left == right or abs(len(left) - len(right)) > 1:
        return False
    if len(left) > len(right):
        left, right = right, left
    if len(left) == len(right):
        differences = sum(a != b for a, b in zip(left, right, strict=True))
        return differences == 1

    left_index = 0
    right_index = 0
    skipped = False
    while left_index < len(left) and right_index < len(right):
        if left[left_index] == right[right_index]:
            left_index += 1
            right_index += 1
            continue
        if skipped:
            return False
        skipped = True
        right_index += 1
    return True


def path_hygiene_findings(files: list[Path]) -> tuple[list[str], list[str]]:
    findings: list[str] = []
    allowed_exceptions: list[str] = []
    casefolded: dict[str, list[str]] = defaultdict(list)
    siblings: dict[tuple[str, str], list[Path]] = defaultdict(list)

    for path in files:
        relative = path.as_posix()
        casefolded[relative.casefold()].append(relative)

        if not relative.isascii():
            findings.append(f"non-ASCII tracked path: {relative}")
        if any(character.isspace() for character in relative):
            findings.append(f"whitespace in tracked path: {relative}")
        if any(ord(character) < 32 or ord(character) == 127 for character in relative):
            findings.append(f"control character in tracked path: {relative!r}")
        if any("-" in part and "_" in part for part in path.parts):
            findings.append(f"mixed hyphen/underscore naming in path segment: {relative}")

        parent = path.parent.as_posix()
        suffix = path.suffix.casefold()
        siblings[(parent, suffix)].append(path)

    for values in casefolded.values():
        if len(values) > 1:
            findings.append(f"case-only path collision: {', '.join(sorted(values))}")

    for paths in siblings.values():
        ordered = sorted(paths, key=lambda path: path.name.casefold())
        for index, left in enumerate(ordered):
            for right in ordered[index + 1 :]:
                left_name = left.name.casefold()
                right_name = right.name.casefold()
                if min(len(left_name), len(right_name)) < 5:
                    continue
                if not differs_by_one_edit(left_name, right_name):
                    continue

                pair = frozenset({left.as_posix(), right.as_posix()})
                if pair in ALLOWED_NEAR_DUPLICATE_PAIRS:
                    allowed_exceptions.append(
                        "explicit rule-entrypoint exception: "
                        + ", ".join(sorted(pair))
                    )
                else:
                    findings.append(
                        "near-duplicate sibling paths: "
                        + ", ".join(sorted(pair))
                    )

    return sorted(set(findings)), sorted(set(allowed_exceptions))


def contamination_findings(files: list[Path]) -> list[str]:
    findings: list[str] = []
    for path in files:
        relative = path.as_posix()
        parts = set(path.parts)
        if parts & FORBIDDEN_PARTS:
            findings.append(f"forbidden tracked directory: {relative}")
        if path.name in FORBIDDEN_NAMES:
            findings.append(f"forbidden tracked runtime/system file: {relative}")
        if path.suffix.lower() in FORBIDDEN_SUFFIXES:
            findings.append(f"forbidden tracked generated/runtime suffix: {relative}")
        if relative.startswith(GENERATED_PREFIXES):
            findings.append(f"forbidden tracked generated artifact: {relative}")
        if not path.is_file():
            findings.append(f"tracked path is not a regular file: {relative}")
    return sorted(set(findings))


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: source_inventory.py OUTPUT_JSON", file=sys.stderr)
        return 2

    output = Path(sys.argv[1])
    files = tracked_files()
    path_findings, allowed_exceptions = path_hygiene_findings(files)
    contamination = contamination_findings(files)
    entries: list[dict[str, object]] = []

    for path in files:
        if not path.is_file():
            continue
        relative = path.as_posix()
        mode = stat.S_IMODE(path.stat().st_mode)
        entries.append(
            {
                "path": relative,
                "size_bytes": path.stat().st_size,
                "mode": f"{mode:04o}",
                "sha256": sha256(path),
            }
        )

    payload = {
        "commit": subprocess.check_output(
            ["git", "rev-parse", "HEAD"], text=True
        ).strip(),
        "tracked_file_count": len(entries),
        "files": entries,
        "contamination_findings": contamination,
        "path_hygiene_findings": path_findings,
        "allowed_path_exceptions": allowed_exceptions,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")

    all_findings = contamination + path_findings
    if all_findings:
        for finding in all_findings:
            print(f"ERROR: {finding}", file=sys.stderr)
        return 1
    print(f"PASS: source inventory contains {len(entries)} tracked files")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
