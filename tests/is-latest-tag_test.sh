#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT="$REPO_ROOT/.github/scripts/is-latest-tag.sh"

PASS=0
FAIL=0

run_test() {
  local name="$1"
  local tag="$2"
  local expected_exit="$3"
  shift 3
  local tags=("$@")

  local tmpdir
  tmpdir="$(mktemp -d)"
  trap "rm -rf '$tmpdir'" RETURN

  git -C "$tmpdir" init -q
  git -C "$tmpdir" -c user.name=test -c user.email=test@example.com commit --allow-empty -m "init" -q

  for t in "${tags[@]}"; do
    git -C "$tmpdir" tag "$t"
  done

  local actual_exit=0
  GIT_DIR="$tmpdir/.git" bash "$SCRIPT" "$tag" || actual_exit=$?

  if [[ "$actual_exit" -eq "$expected_exit" ]]; then
    echo "PASS: $name"
    PASS=$((PASS + 1))
  else
    echo "FAIL: $name (expected exit $expected_exit, got $actual_exit)"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== is-latest-tag.sh tests ==="

run_test "newest tag is latest" \
  "v0.26.0" 0 \
  "v0.24.0" "v0.25.0" "v0.26.0"

run_test "older patch after newer release" \
  "v0.25.1" 1 \
  "v0.24.0" "v0.25.0" "v0.25.1" "v0.26.0"

run_test "pre-release tag is never latest" \
  "v0.27.0-rc.1" 1 \
  "v0.24.0" "v0.25.0" "v0.26.0" "v0.27.0-rc.1"

run_test "lone pre-release" \
  "v0.27.0-rc.1" 1 \
  "v0.27.0-rc.1"

run_test "first-ever tag is latest" \
  "v0.1.0" 0 \
  "v0.1.0"

run_test "equal to current highest" \
  "v0.26.0" 0 \
  "v0.26.0"

run_test "pre-releases excluded from candidate set" \
  "v0.26.0" 0 \
  "v0.25.0" "v0.26.0" "v0.27.0-beta.1"

run_test "higher major version" \
  "v1.0.0" 0 \
  "v0.25.0" "v0.26.0" "v1.0.0"

run_test "lower major version" \
  "v0.26.0" 1 \
  "v0.26.0" "v1.0.0"

run_test "no argument prints usage" \
  "" 2

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ "$FAIL" -eq 0 ]]
