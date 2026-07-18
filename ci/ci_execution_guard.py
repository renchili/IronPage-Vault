#!/usr/bin/env python3
"""Enforce target cooldown and an auditable failed-revision latch.

GitHub Actions history is the persistent control record. A reviewed fix commit
changes the SHA and clears the failure latch. A deliberate workflow_dispatch
with the unlock input may replay the same SHA once.
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


def run_target(run: dict[str, object]) -> str:
    pull_requests = run.get("pull_requests")
    if isinstance(pull_requests, list) and pull_requests:
        first = pull_requests[0]
        if isinstance(first, dict) and first.get("number") is not None:
            return f"pr-{first['number']}"
    head_branch = run.get("head_branch")
    return str(head_branch or "")


def main() -> int:
    repository = required("GITHUB_REPOSITORY")
    target = required("IRONPAGE_CI_TARGET")
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
        f"https://api.github.com/repos/{repository}/actions/runs?per_page=100",
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
        if isinstance(run, dict)
        and int(run.get("id", 0)) != current_run_id
        and run_target(run) == target
    ]
    previous_runs.sort(
        key=lambda run: parse_time(run.get("run_started_at") or run.get("created_at"))
        or dt.datetime.min.replace(tzinfo=dt.timezone.utc),
        reverse=True,
    )

    if unlock:
        print(f"PASS: reviewed manual CI unlock accepted for target {target}")
        return 0

    failed_same_revision = [
        run
        for run in previous_runs
        if run.get("head_sha") == sha and run.get("conclusion") in FAILED_CONCLUSIONS
    ]
    if failed_same_revision:
        run = failed_same_revision[0]
        print(
            "ERROR: this target and revision are latched after a failed run "
            f"({run.get('html_url')}); push a reviewed fix commit before automatic verification",
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
            f"ERROR: target {target} already started verification within the 10-minute cooldown "
            f"({run.get('html_url')})",
            file=sys.stderr,
        )
        return 1

    print(f"PASS: CI cooldown and failed-revision latch for target {target}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
