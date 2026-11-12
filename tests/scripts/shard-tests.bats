#!/usr/bin/env bats

SHARD="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)/.github/scripts/shard-tests.sh"

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

names_in() {
    tr -d '^()$' <<<"$1" | tr '|' '\n'
}

shard_names() {
    names_in "$(printf '%s\n' "$NAMES" | bash "$SHARD" "$@")"
}

shard() {
    local input="$1"
    shift
    run bash -c 'printf "%s" "$1" | bash "$2" "${@:3}"' _ "$input" "$SHARD" "$@"
}

@test "the shards together hold every test exactly once" {
    union=$(for s in 1 2 3; do shard_names 3 "$s"; done | LC_ALL=C sort)
    [ "$union" = "$(printf '%s\n' "$NAMES" | LC_ALL=C sort)" ]
}

@test "the input order does not change a shard" {
    reversed=$(names_in "$(printf '%s\n' "$NAMES" | tac | bash "$SHARD" 3 2)")
    [ "$reversed" = "$(shard_names 3 2)" ]
}

@test "removing other names does not move a name to another shard" {
    dropped='^Test(Alpha|Beta|Gamma)$'
    for s in 1 2 3; do
        full=$(shard_names 3 "$s" | grep -vE "$dropped" || true)
        partial=$(names_in "$(printf '%s\n' "$NAMES" | grep -vE "$dropped" | bash "$SHARD" 3 "$s" 2>/dev/null || true)")
        [ "$full" = "$partial" ]
    done
}

@test "names matching the first-shard regex land in shard 1" {
    pinned=$(shard_names 3 1 '^TestAgentDockerSelfUpdate')
    grep -qx TestAgentDockerSelfUpdate <<<"$pinned"
    grep -qx TestAgentDockerSelfUpdateFailure <<<"$pinned"
}

@test "names matching the first-shard regex land in no other shard" {
    others=$(for s in 2 3; do shard_names 3 "$s" '^TestAgentDockerSelfUpdate'; done)
    ! grep -q TestAgentDockerSelfUpdate <<<"$others"
}

@test "a single shard selects an anchored pattern" {
    shard TestOnly 1 1
    [ "$status" -eq 0 ]
    [ "$output" = '^(TestOnly)$' ]
}

@test "no arguments prints usage" {
    shard "$NAMES"
    [ "$status" -eq 2 ]
    [[ "$output" == *"Usage:"* ]]
}

@test "a shard above the count fails" {
    shard "$NAMES" 3 4
    [ "$status" -eq 2 ]
}

@test "shard zero fails" {
    shard "$NAMES" 3 0
    [ "$status" -eq 2 ]
}

@test "an empty list fails" {
    shard "" 3 1
    [ "$status" -eq 1 ]
    [[ "$output" == *"no test names on stdin"* ]]
}

@test "a name that is not a Go identifier fails" {
    shard "TestA|TestB" 1 1
    [ "$status" -eq 1 ]
    [[ "$output" == *"not a Go test name"* ]]
}

@test "a shard that holds no tests fails" {
    shard TestOnly 2 2 '^TestOnly$'
    [ "$status" -eq 1 ]
    [[ "$output" == *"holds no tests"* ]]
}
