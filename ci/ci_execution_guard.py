#!/usr/bin/env python3
"""Enforce target cooldown and an auditable failed-revision latch.

GitHub Actions history is the persistent control record. A reviewed fix commit
changes the SHA and clears the failure latch. A deliberate workflow_dispatch
with the unlock input may replay the same SHA once. New fix revisions wait out
any remaining target cooldown instead of being recorded as false failures.
"""

from __future__ import annotations

import datetime as dt
import json
import os
import sys
import time
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


def fetch_runs(repository: str, token: str) -> list[dict[str, object]]:
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
        raise RuntimeError(f"unable to inspect GitHub Actions history: {exc}") from exc
    runs = payload.get("workflow_runs", [])
    return [run for run in runs if isinstance(run, dict)]


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

    try:
        runs = fetch_runs(repository, token)
    except RuntimeError as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1

    previous_runs = [
        run
        for run in runs
        if int(run.get("id", 0)) != current_run_id and run_target(run) == target
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
    latest_started = next(
        (
            parse_time(run.get("run_started_at") or run.get("created_at"))
            for run in previous_runs
            if parse_time(run.get("run_started_at") or run.get("created_at"))
        ),
        None,
    )
    if latest_started is not None:
        elapsed = max(0.0, (now - latest_started).total_seconds())
        remaining = max(0, int(COOLDOWN_SECONDS - elapsed + 0.999))
        if remaining:
            print(
                f"WAIT: target {target} has {remaining}s remaining in the 10-minute "
                "start cooldown; no validation stage will begin before it expires"
            )
            time.sleep(remaining)

    print(f"PASS: CI cooldown and failed-revision latch for target {target}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
