#!/usr/bin/env bash
set -euo pipefail

BASE_SHA=${BASE_SHA:-${GITHUB_BASE_REF:-}}
if [ -z "${BASE_SHA:-}" ]; then
  BASE_SHA=${1:-origin/main}
fi

if git rev-parse --verify "$BASE_SHA" >/dev/null 2>&1; then
  base_ref="$BASE_SHA"
elif [ -n "${GITHUB_BASE_REF:-}" ]; then
  git fetch --no-tags --depth=1 origin "$GITHUB_BASE_REF"
  base_ref="origin/$GITHUB_BASE_REF"
else
  git fetch --no-tags --depth=1 origin main
  base_ref="origin/main"
fi

mapfile -t changed < <(git diff --name-status --find-renames "$base_ref"...HEAD)

failures=()
notes=()
changed_paths=()

for row in "${changed[@]}"; do
  status=$(awk '{print $1}' <<<"$row")
  path=$(awk '{print $NF}' <<<"$row")
  changed_paths+=("$path")
  notes+=("$status $path")
done

path_changed() {
  local pattern="$1"
  local path
  for path in "${changed_paths[@]}"; do
    [[ "$path" == $pattern ]] && return 0
  done
  return 1
}

same_dir_test_changed_or_exists() {
  local src="$1"
  local dir
  dir=$(dirname "$src")
  local path
  for path in "${changed_paths[@]}"; do
    [[ "$path" == "$dir"/*_test.go ]] && return 0
  done
  compgen -G "$dir/*_test.go" >/dev/null
}

same_dir_test_changed() {
  local src="$1"
  local dir
  dir=$(dirname "$src")
  local path
  for path in "${changed_paths[@]}"; do
    [[ "$path" == "$dir"/*_test.go ]] && return 0
  done
  return 1
}

api_contract_changed() {
  path_changed 'internal/app/swagger_*.go' || path_changed 'docs/swagger/**' || path_changed 'API_tests/**' || path_changed 'ci/swagger_contract_check.sh'
}

workflow_executes_script() {
  local script="$1"
  grep -R -F "bash $script" .github/workflows >/dev/null 2>&1
}

workflow_executes_contract() {
  local contract="$1"
  grep -R -F "bash $contract" .github/workflows >/dev/null 2>&1
}

contract_executes_script() {
  local script="$1"
  local contract
  compgen -G 'ci/*contract*_check.sh' >/dev/null || return 1
  for contract in ci/*contract*_check.sh; do
    grep -F "$script" "$contract" >/dev/null 2>&1 || continue
    workflow_executes_contract "$contract" && return 0
  done
  return 1
}

ci_script_execution_covered() {
  local script="$1"
  workflow_executes_script "$script" || contract_executes_script "$script"
}

printf 'Changed files inspected by CI impact gate:\n'
printf '  %s\n' "${notes[@]}"

for row in "${changed[@]}"; do
  status=$(awk '{print $1}' <<<"$row")
  path=$(awk '{print $NF}' <<<"$row")

  case "$path" in
    cmd/*.go|cmd/**/*.go|internal/*.go|internal/**/*.go)
      [[ "$path" == *_test.go ]] && continue
      if [[ "$status" == A* ]]; then
        if ! same_dir_test_changed "$path"; then
          failures+=("new Go source $path must include a changed same-package *_test.go")
        fi
      elif [[ "$status" == M* || "$status" == R* ]]; then
        if ! same_dir_test_changed_or_exists "$path"; then
          failures+=("modified Go source $path has no same-package *_test.go coverage")
        fi
      fi
      ;;
    ci/*.sh)
      if ! ci_script_execution_covered "$path"; then
        failures+=("changed CI script $path is not executed directly by PR CI or by a workflow-executed contract check")
      fi
      ;;
  esac

done

if git diff --unified=0 "$base_ref"...HEAD -- internal/app ':!**/*_test.go' 2>/dev/null | grep -E '^\+.*(@Router|\.GET\(|\.POST\(|\.PATCH\(|\.PUT\(|\.DELETE\()' >/dev/null; then
  if ! api_contract_changed; then
    failures+=("API route/handler contract changed without swagger/API contract test update")
  fi
fi

if path_changed 'migrations/**'; then
  if ! path_changed 'internal/repository/**' && ! path_changed 'internal/store/**' && ! path_changed 'API_tests/**'; then
    failures+=("migration changed without repository/store or API test update")
  fi
fi

if [ "${#failures[@]}" -ne 0 ]; then
  echo "CI impact gate failed:"
  printf '  - %s\n' "${failures[@]}"
  exit 1
fi

echo "PASS CI impact gate"
