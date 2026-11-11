#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT="$REPO_ROOT/.github/scripts/shard-tests.sh"

PASS=0
FAIL=0

NAMES="TestAlpha
TestBeta
TestGamma
TestDelta
TestEpsilon
TestZeta
TestEta
TestTheta
TestAgentDockerSelfUpdate
TestAgentDockerSelfUpdateFailure"

check() {
  local name="$1" ok="$2"
  if [[ "$ok" == true ]]; then
    echo "PASS: $name"
    PASS=$((PASS + 1))
  else
    echo "FAIL: $name"
    FAIL=$((FAIL + 1))
  fi
}

expect_exit() {
  local name="$1" expected="$2" input="$3"
  shift 3
  local actual=0
  printf '%s' "$input" | bash "$SCRIPT" "$@" >/dev/null 2>&1 || actual=$?
  check "$name (expected exit $expected, got $actual)" "$([[ "$actual" -eq "$expected" ]] && echo true || echo false)"
}

shard_names() {
  printf '%s\n' "$NAMES" | bash "$SCRIPT" "$@" | tr -d '^()$' | tr '|' '\n'
}

echo "=== shard-tests.sh tests ==="

union=$(for s in 1 2 3; do shard_names 3 "$s"; done | LC_ALL=C sort)
check "the shards together hold every test exactly once" \
  "$([[ "$union" == "$(printf '%s\n' "$NAMES" | LC_ALL=C sort)" ]] && echo true || echo false)"

reversed=$(printf '%s\n' "$NAMES" | tac | bash "$SCRIPT" 3 2 | tr -d '^()$' | tr '|' '\n')
check "the input order does not change a shard" \
  "$([[ "$reversed" == "$(shard_names 3 2)" ]] && echo true || echo false)"

dropped='^Test(Alpha|Beta|Gamma)$'
stable=true
for s in 1 2 3; do
  full=$(shard_names 3 "$s" | grep -vE "$dropped" || true)
  partial=$(printf '%s\n' "$NAMES" | grep -vE "$dropped" | bash "$SCRIPT" 3 "$s" 2>/dev/null | tr -d '^()$' | tr '|' '\n' || true)
  [[ "$full" == "$partial" ]] || stable=false
done
check "removing other names does not move a name to another shard" "$stable"

pinned=$(shard_names 3 1 '^TestAgentDockerSelfUpdate')
check "names matching the first-shard regex land in shard 1" \
  "$(grep -qx TestAgentDockerSelfUpdate <<<"$pinned" && grep -qx TestAgentDockerSelfUpdateFailure <<<"$pinned" && echo true || echo false)"

others=$(for s in 2 3; do shard_names 3 "$s" '^TestAgentDockerSelfUpdate'; done)
check "names matching the first-shard regex land in no other shard" \
  "$(grep -q TestAgentDockerSelfUpdate <<<"$others" && echo false || echo true)"

pattern=$(printf 'TestOnly\n' | bash "$SCRIPT" 1 1)
check "a single shard selects an anchored pattern" \
  "$([[ "$pattern" == '^(TestOnly)$' ]] && echo true || echo false)"

expect_exit "no arguments prints usage" 2 "$NAMES"
expect_exit "shard above the count" 2 "$NAMES" 3 4
expect_exit "shard zero" 2 "$NAMES" 3 0
expect_exit "empty list" 1 "" 3 1
expect_exit "a name that is not a Go identifier" 1 "TestA|TestB" 1 1
expect_exit "a shard that holds no tests" 1 "TestOnly" 2 2 '^TestOnly$'

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ "$FAIL" -eq 0 ]]
