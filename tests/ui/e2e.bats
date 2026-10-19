#!/usr/bin/env bats

bats_require_minimum_version 1.5.0

REPO_ROOT="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)"
E2E="$REPO_ROOT/ui/scripts/e2e.sh"

setup() {
    STUBS="$BATS_TEST_TMPDIR/bin"
    export STUB_LOG="$BATS_TEST_TMPDIR/calls"
    export STUB_BLOCKED="$BATS_TEST_TMPDIR/blocked"
    export STUB_DOWN_BLOCKED="$BATS_TEST_TMPDIR/down-blocked"
    mkdir -p "$STUBS"

    cat >"$STUBS/go" <<'EOF'
#!/bin/sh
echo "go $*" >>"$STUB_LOG"
case "$*" in
*"cmd/stack down"*)
    case "${STUB_DOWN:-}" in
    fail) exit 1 ;;
    block)
        : >"$STUB_DOWN_BLOCKED"
        exec sleep 600
        ;;
    esac
    ;;
*"cmd/stack up"*)
    echo "export E2E_COMPOSE_PROJECT='shellhub-e2e-stub'"
    echo "export E2E_EDITION='community'"
    echo "export E2E_BASE_URL='http://localhost:1'"
    echo "export E2E_STACK_EXPIRES='2026-10-20T12:00:00Z'"
    ;;
esac
EOF

    cat >"$STUBS/npx" <<'EOF'
#!/bin/sh
echo "npx $*" >>"$STUB_LOG"
case "$STUB_PLAYWRIGHT" in
pass) exit 0 ;;
fail) exit 1 ;;
block)
    : >"$STUB_BLOCKED"
    exec sleep 600
    ;;
esac
EOF

    chmod +x "$STUBS/go" "$STUBS/npx"
}

e2e() {
    exec env -u CI -u E2E_KEEP -u E2E_KEEP_TTL PATH="$STUBS:$PATH" E2E_RUNNER=host "$@" bash "$E2E" test
}

playwright_stopped() {
    for _ in $(seq 50); do
        pgrep -f "sleep 600" >/dev/null || return 0
        sleep 0.1
    done
    return 1
}

await() {
    for _ in $(seq 100); do
        [ -e "$1" ] && return 0
        sleep 0.1
    done
    return 1
}

start_blocked() {
    set -m
    e2e STUB_PLAYWRIGHT=block "$@" 2>"$BATS_TEST_TMPDIR/stderr" &
    pid=$!
    set +m

    await "$STUB_BLOCKED"
}

interrupt_with() {
    local signal=$1

    start_blocked "${@:2}"

    kill "-$signal" -- "-$pid"
    status=0
    wait "$pid" || status=$?
}

@test "test tears the stack down at once when only the script gets TERM" {
    e2e STUB_PLAYWRIGHT=block 2>/dev/null &
    local pid=$!

    await "$STUB_BLOCKED"

    kill -TERM "$pid"
    for _ in $(seq 50); do
        kill -0 "$pid" 2>/dev/null || break
        sleep 0.1
    done
    if kill -0 "$pid" 2>/dev/null; then
        pkill -f "sleep 600" || true
        wait "$pid" || true
        false
    fi

    status=0
    wait "$pid" || status=$?
    [ "$status" -eq 143 ]
    grep -q "cmd/stack down" "$STUB_LOG"
    playwright_stopped
}

@test "test keeps the tests' status when the teardown fails" {
    run e2e STUB_PLAYWRIGHT=pass STUB_DOWN=fail
    [ "$status" -eq 0 ]
    [[ "$output" == *"ui/scripts/e2e.sh down"* ]]
}

@test "test tears the stack down when the tests pass" {
    run e2e STUB_PLAYWRIGHT=pass
    [ "$status" -eq 0 ]
    grep -q "cmd/stack down" "$STUB_LOG"
}

@test "test tears the stack down when the tests fail and keeps their status" {
    run e2e STUB_PLAYWRIGHT=fail
    [ "$status" -eq 1 ]
    grep -q "cmd/stack down" "$STUB_LOG"
}

@test "test tears the stack down on INT and exits with its status" {
    interrupt_with INT
    [ "$status" -eq 130 ]
    grep -q "cmd/stack down" "$STUB_LOG"
}

@test "test tears the stack down on TERM and exits with its status" {
    interrupt_with TERM
    [ "$status" -eq 143 ]
    grep -q "cmd/stack down" "$STUB_LOG"
}

@test "test says how to drop the stack when its teardown is interrupted" {
    start_blocked STUB_DOWN=block

    kill -INT -- "-$pid"
    await "$STUB_DOWN_BLOCKED"
    kill -INT -- "-$pid"
    status=0
    wait "$pid" || status=$?

    [ "$status" -eq 130 ]
    grep -q "ui/scripts/e2e.sh down" "$BATS_TEST_TMPDIR/stderr"
    playwright_stopped
}

@test "the drop command names the runner it was run with" {
    run e2e STUB_PLAYWRIGHT=pass E2E_KEEP=1
    [ "$status" -eq 0 ]
    [[ "$output" == *"E2E_RUNNER=host ui/scripts/e2e.sh down"* ]]
}

@test "test tears the stack down on HUP and exits with its status" {
    interrupt_with HUP
    [ "$status" -eq 129 ]
    grep -q "cmd/stack down" "$STUB_LOG"
    playwright_stopped
}

@test "E2E_KEEP says how to drop the stack when the tests are interrupted" {
    interrupt_with INT E2E_KEEP=1
    [ "$status" -eq 130 ]
    grep -q "ui/scripts/e2e.sh down" "$BATS_TEST_TMPDIR/stderr"
    grep -q "2026-10-20T12:00:00Z" "$BATS_TEST_TMPDIR/stderr"
    run ! grep -q "cmd/stack down" "$STUB_LOG"
}

@test "E2E_KEEP keeps the stack and says how to drop it and when it expires" {
    run e2e STUB_PLAYWRIGHT=pass E2E_KEEP=1
    [ "$status" -eq 0 ]
    [[ "$output" == *"http://localhost:1"* ]]
    [[ "$output" == *"ui/scripts/e2e.sh down"* ]]
    [[ "$output" == *"2026-10-20T12:00:00Z"* ]]
    run ! grep -q "cmd/stack down" "$STUB_LOG"
}

@test "E2E_KEEP names the stack in the command that drops it" {
    run e2e STUB_PLAYWRIGHT=pass E2E_KEEP=1 E2E_STACK_NAME=mine
    [ "$status" -eq 0 ]
    [[ "$output" == *"E2E_STACK_NAME=mine ui/scripts/e2e.sh down"* ]]
}

@test "E2E_KEEP_TTL reaches stack up" {
    run e2e STUB_PLAYWRIGHT=pass E2E_KEEP_TTL=72h
    grep -q "cmd/stack up .*--ttl 72h" "$STUB_LOG"
}
