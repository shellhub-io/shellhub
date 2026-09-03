# Helpers for install.bats and shellhub-agent-wrapper.bats.
#
# Strategy: run install.sh (or the wrapper it emits) with a PATH holding only
# a curated set of real utilities plus a directory of stubs that append their
# own argv to $CALLS. Asserting on that log covers the decisions — which
# container flags are passed, which env vars are mapped, when the script
# aborts — without a container runtime, a systemd, or a network.
#
# A binary left out of both lists is genuinely absent, which is how the
# "systemctl is missing" and "xdg-open is missing" branches are reached.

INSTALL_SH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/install.sh"

# Utilities install.sh needs to do its work. Deliberately excludes everything
# whose presence a branch keys off: docker, podman, snap, systemctl, sudo,
# wslinfo, find, xdg-open, curl, wget.
INSTALL_TEST_REAL_BINS="sh sed grep awk tr wc cat head tail cut mktemp chmod mkdir rm cp mv ln gzip sleep uname tee env dirname basename"

# Build the curated PATH once per file, then a per-test stub dir on top.
setup_install_env() {
    REAL_BIN="$BATS_FILE_TMPDIR/real-bin"

    if [ ! -d "$REAL_BIN" ]; then
        mkdir -p "$REAL_BIN"
        for bin in $INSTALL_TEST_REAL_BINS; do
            if path=$(command -v "$bin"); then
                ln -sf "$path" "$REAL_BIN/$bin"
            fi
        done
    fi

    STUB_DIR="$BATS_TEST_TMPDIR/stub-bin"
    CALLS="$BATS_TEST_TMPDIR/calls"
    mkdir -p "$STUB_DIR"
    : > "$CALLS"

    export REAL_BIN STUB_DIR CALLS
}

# Put an executable named $1 on the stubbed PATH. Its body is $2, or a
# default that just records the invocation. Bodies are POSIX sh and can use
# $CALLS, "$@" and exit codes.
# Usage: stub_bin docker 'echo "docker $*" >> "$CALLS"; exit 0'
stub_bin() {
    local name="$1" body="${2-}"

    if [ -z "$body" ]; then
        body="echo \"$name \$*\" >> \"\$CALLS\""
    fi

    {
        echo '#!/bin/sh'
        echo "$body"
    } > "$STUB_DIR/$name"
    chmod +x "$STUB_DIR/$name"
}

# Source install.sh as a library in a fresh /bin/sh and call one of its
# functions. Runs under the real POSIX shell rather than bats' bash, so the
# functions are exercised the way a user's `curl | sh` runs them.
# Usage: call_install check_podman_boot_restart
call_install() {
    run env PATH="$STUB_DIR:$REAL_BIN" INSTALL_SH_LIB=1 \
        sh -c '. "$1"; shift; "$@"' sh "$INSTALL_SH" "$@"
}

# Run install.sh end to end, exercising the detection flow in main().
run_install() {
    run env PATH="$STUB_DIR:$REAL_BIN" "$INSTALL_SH" "$@"
}

assert_called() {
    grep -qF -- "$1" "$CALLS" && return 0

    echo "expected a recorded call containing: $1"
    echo "--- recorded calls ---"
    cat "$CALLS"
    return 1
}

refute_called() {
    grep -qF -- "$1" "$CALLS" || return 0

    echo "unexpected recorded call containing: $1"
    echo "--- recorded calls ---"
    cat "$CALLS"
    return 1
}

assert_file_contains() {
    grep -qF -- "$2" "$1" && return 0

    echo "expected $1 to contain: $2"
    echo "--- $1 ---"
    cat "$1"
    return 1
}

refute_file_contains() {
    grep -qF -- "$2" "$1" || return 0

    echo "expected $1 not to contain: $2"
    echo "--- $1 ---"
    cat "$1"
    return 1
}

assert_output_contains() {
    [[ "$output" == *"$1"* ]] && return 0

    echo "expected output to contain: $1"
    echo "--- output ---"
    echo "$output"
    return 1
}

refute_output_contains() {
    [[ "$output" != *"$1"* ]] && return 0

    echo "expected output not to contain: $1"
    echo "--- output ---"
    echo "$output"
    return 1
}
