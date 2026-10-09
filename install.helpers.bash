INSTALL_SH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/install.sh"

INSTALL_TEST_REAL_BINS="sh sed grep awk tr wc cat head tail cut mktemp chmod mkdir rm cp mv ln gzip sha256sum sleep uname tee env dirname basename"

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

call_install() {
    run env PATH="$STUB_DIR:$REAL_BIN" INSTALL_SH_LIB=1 \
        sh -c '. "$1"; shift; "$@"' sh "$INSTALL_SH" "$@"
}

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
