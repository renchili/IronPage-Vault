#!/usr/bin/env python3
"""Reject duplicate or unsafe CI starts for the same repository revision.

GitHub Actions history is the persistent, auditable latch. A new commit SHA
clears a failed-revision latch. A reviewed workflow_dispatch may explicitly
unlock one replay of the same SHA.
"""

from __future__ import annotations

import datetime as dt
import json
import os
import sys
import urllib.error
import urllib.request

COOLDOWN_SECONDS = 10 * 60
FAILED_CONCLUSIONS = {
    "failure",
    "timed_out",
    "cancelled",
    "action_required",
    "stale",
}


def required(name: str) -> str:
    value = os.environ.get(name, "").strip()
    if not value:
        raise SystemExit(f"ERROR: {name} is required for the CI execution guard")
    return value


def parse_time(value: str | None) -> dt.datetime | None:
    if not value:
        return None
    return dt.datetime.fromisoformat(value.replace("Z", "+00:00"))


def main() -> int:
    repository = required("GITHUB_REPOSITORY")
    sha = required("GITHUB_SHA")
    token = required("GITHUB_TOKEN")
    current_run_id = int(required("GITHUB_RUN_ID"))
    current_attempt = int(os.environ.get("GITHUB_RUN_ATTEMPT", "1"))
    unlock = os.environ.get("IRONPAGE_CI_UNLOCK", "false").lower() == "true"

    if current_attempt > 1 and not unlock:
        print(
            "ERROR: automatic rerun of the same failed revision is locked; "
            "push a reviewed fix commit or use the explicit manual unlock input",
            file=sys.stderr,
        )
        return 1

    request = urllib.request.Request(
        f"https://api.github.com/repos/{repository}/actions/runs?head_sha={sha}&per_page=100",
        headers={
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {token}",
            "X-GitHub-Api-Version": "2022-11-28",
            "User-Agent": "ironpage-ci-execution-guard",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            payload = json.load(response)
    except (urllib.error.URLError, json.JSONDecodeError) as exc:
        print(f"ERROR: unable to inspect GitHub Actions history: {exc}", file=sys.stderr)
        return 1

    previous_runs = [
        run
        for run in payload.get("workflow_runs", [])
        if int(run.get("id", 0)) != current_run_id
    ]
    previous_runs.sort(
        key=lambda run: parse_time(run.get("run_started_at") or run.get("created_at"))
        or dt.datetime.min.replace(tzinfo=dt.timezone.utc),
        reverse=True,
    )

    if unlock:
        print("PASS: reviewed manual CI unlock accepted for this revision")
        return 0

    failed = [run for run in previous_runs if run.get("conclusion") in FAILED_CONCLUSIONS]
    if failed:
        run = failed[0]
        print(
            "ERROR: this revision is latched after a failed run "
            f"({run.get('html_url')}); push a fix commit before automatic verification",
            file=sys.stderr,
        )
        return 1

    now = dt.datetime.now(dt.timezone.utc)
    recent = []
    for run in previous_runs:
        started = parse_time(run.get("run_started_at") or run.get("created_at"))
        if started and 0 <= (now - started).total_seconds() < COOLDOWN_SECONDS:
            recent.append(run)
    if recent:
        run = recent[0]
        print(
            "ERROR: another run for this revision started within the 10-minute cooldown "
            f"({run.get('html_url')})",
            file=sys.stderr,
        )
        return 1

    print("PASS: CI cooldown and failed-revision latch")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
