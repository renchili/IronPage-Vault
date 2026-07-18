#!/usr/bin/env python3
"""Generate a tracked-source manifest and reject repository contamination."""

from __future__ import annotations

import hashlib
import json
import os
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


def tracked_files() -> list[Path]:
    result = subprocess.run(
        ["git", "ls-files", "-z"],
        check=True,
        stdout=subprocess.PIPE,
    )
    return [Path(item.decode("utf-8")) for item in result.stdout.split(b"\0") if item]


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: source_inventory.py OUTPUT_JSON", file=sys.stderr)
        return 2

    output = Path(sys.argv[1])
    files = tracked_files()
    findings: list[str] = []
    entries: list[dict[str, object]] = []

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
            continue

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
        "commit": subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip(),
        "tracked_file_count": len(entries),
        "files": entries,
        "contamination_findings": findings,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")

    if findings:
        for finding in findings:
            print(f"ERROR: {finding}", file=sys.stderr)
        return 1
    print(f"PASS: source inventory contains {len(entries)} tracked files")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
