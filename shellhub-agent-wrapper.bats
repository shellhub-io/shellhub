#!/usr/bin/env bats
# Tests for the shellhub-agent wrapper that install.sh generates. The wrapper
# is a separate artifact with its own runtime behaviour — it locates the agent
# container and proxies commands into it — so install.sh appears here only in
# setup_file, to produce the fixture. Coverage of the generation itself lives
# in install.bats.

load install.helpers

setup_file() {
    fixture_bin="$BATS_FILE_TMPDIR/fixture-bin"
    mkdir -p "$fixture_bin"
    printf '#!/bin/sh\necho 0\n' > "$fixture_bin/id"
    chmod +x "$fixture_bin/id"

    for runtime in docker podman; do
        mkdir -p "$BATS_FILE_TMPDIR/$runtime"
        env PATH="$fixture_bin:$PATH" INSTALL_SH_LIB=1 INSTALL_DIR="$BATS_FILE_TMPDIR/$runtime" \
            sh -c '. "$1"; install_agent_wrapper "$2"' sh "$INSTALL_SH" "$runtime" > /dev/null
    done
}

setup() {
    setup_install_env

    RUNTIME=docker
    WRAPPER="$BATS_FILE_TMPDIR/docker/shellhub-agent"
}

use_podman() {
    RUNTIME=podman
    WRAPPER="$BATS_FILE_TMPDIR/podman/shellhub-agent"
}

# Stub the container runtime the wrapper drives.
# Usage: stub_runtime <ps output> [exec output] [exec exit status]
stub_runtime() {
    export STUB_CONTAINERS="${1-}" STUB_EXEC_OUTPUT="${2-}" STUB_EXEC_STATUS="${3:-0}"

    stub_bin "$RUNTIME" "echo \"$RUNTIME \$*\" >> \"\$CALLS\"
case \"\$1\" in
ps) [ -n \"\$STUB_CONTAINERS\" ] && printf '%s\n' \"\$STUB_CONTAINERS\" ;;
exec)
    [ -n \"\$STUB_EXEC_OUTPUT\" ] && printf '%s\n' \"\$STUB_EXEC_OUTPUT\"
    exit \"\$STUB_EXEC_STATUS\"
    ;;
esac
exit 0"
}

run_wrapper() {
    run env PATH="$STUB_DIR:$REAL_BIN" "$WRAPPER" "$@"
}

@test "the wrapper looks the agent container up by its role label" {
    stub_runtime shellhub

    run_wrapper status

    [ "$status" -eq 0 ]
    assert_called "docker ps --filter label=shellhub.role=agent --format {{.Names}}"
}

@test "the wrapper proxies its arguments into the single agent container" {
    stub_runtime shellhub

    run_wrapper status --format json

    [ "$status" -eq 0 ]
    assert_called "docker exec shellhub agent status --format json"
}

@test "the wrapper reports when no agent container is running" {
    stub_runtime ""

    run_wrapper status

    [ "$status" -eq 1 ]
    assert_output_contains "no running agent container found"
    assert_output_contains "install.sh uninstall"
    refute_called "docker exec"
}

@test "the wrapper refuses to guess between several agent containers" {
    stub_runtime "$(printf 'shellhub\nshellhub-spare')"

    run_wrapper status

    [ "$status" -eq 1 ]
    assert_output_contains "multiple agent containers found"
    assert_output_contains "  shellhub"
    assert_output_contains "  shellhub-spare"
    refute_called "docker exec"
}

@test "the wrapper propagates the agent's exit status" {
    stub_runtime shellhub "" 3

    run_wrapper status

    [ "$status" -eq 3 ]
}

@test "login opens the accept-device URL in the host browser" {
    stub_runtime shellhub "https://cloud.example.test/accept-device?code=ABC123"
    stub_bin xdg-open

    run_wrapper login

    [ "$status" -eq 0 ]
    assert_called "xdg-open https://cloud.example.test/accept-device?code=ABC123"
    assert_output_contains "(opened in your browser)"
}

@test "login says nothing about a browser it could not open" {
    stub_runtime shellhub "https://cloud.example.test/accept-device?code=ABC123"
    stub_bin xdg-open 'echo "xdg-open $*" >> "$CALLS"; exit 1'

    run_wrapper login

    [ "$status" -eq 0 ]
    assert_called "xdg-open https://cloud.example.test/accept-device?code=ABC123"
    refute_output_contains "opened in your browser"
}

@test "login strips the whitespace around the URL before handing it to the browser" {
    stub_runtime shellhub "  https://cloud.example.test/accept-device?code=ABC123  "
    stub_bin xdg-open

    run_wrapper login

    [ "$status" -eq 0 ]
    assert_called "xdg-open https://cloud.example.test/accept-device?code=ABC123"
}

@test "login streams the agent output while watching for the URL" {
    stub_runtime shellhub "$(printf 'waiting for acceptance\nhttps://cloud.example.test/accept-device?code=ABC123\ndone')"
    stub_bin xdg-open

    run_wrapper login

    [ "$status" -eq 0 ]
    assert_output_contains "waiting for acceptance"
    assert_output_contains "done"
}

@test "login propagates the agent's exit status through the browser path" {
    stub_runtime shellhub "https://cloud.example.test/accept-device?code=ABC123" 3
    stub_bin xdg-open

    run_wrapper login

    [ "$status" -eq 3 ]
}

@test "login proxies straight through when no browser is available" {
    stub_runtime shellhub "https://cloud.example.test/accept-device?code=ABC123"

    run_wrapper login

    [ "$status" -eq 0 ]
    assert_called "docker exec shellhub agent login"
    refute_output_contains "opened in your browser"
}

@test "login propagates the agent's exit status when no browser is available" {
    stub_runtime shellhub "" 3

    run_wrapper login

    [ "$status" -eq 3 ]
}

@test "the podman wrapper drives podman" {
    use_podman
    stub_runtime shellhub

    run_wrapper status

    [ "$status" -eq 0 ]
    assert_called "podman ps --filter label=shellhub.role=agent"
    assert_called "podman exec shellhub agent status"
}
