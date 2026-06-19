#!/usr/bin/env bash
set -euo pipefail

# CI may parse project-owned shell scripts to catch syntax errors, but this
# checker deliberately does not execute run_tests.sh, API_tests, or unit_tests.

paths=()
for root in ci scripts API_tests unit_tests; do
  if [ -d "$root" ]; then
    while IFS= read -r -d '' file; do
      paths+=("$file")
    done < <(find "$root" -name '*.sh' -print0)
  fi
done

if [ -f run_tests.sh ]; then
  paths+=(run_tests.sh)
fi

if [ "${#paths[@]}" -eq 0 ]; then
  echo "no shell scripts found"
  exit 0
fi

for file in "${paths[@]}"; do
  echo "bash -n $file"
  bash -n "$file"
done
